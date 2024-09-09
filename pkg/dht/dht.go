package dht

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"

)

// DHT represents a Distributed Hash Table.
type DHT struct {
	RoutingTable *RoutingTable
	Storage      *storage.Storage
	Network      *message.Network
	stopOnce sync.Once
}

type SuccessMessageResponse struct {
	Value string
	Nodes []*KNode
}

// Serialize SuccessMessageResponse to byte array
func (smr *SuccessMessageResponse) Serialize() []byte {
	var buf bytes.Buffer

	// Serialize Value field with null-termination
	buf.WriteString(smr.Value)
	buf.WriteByte(0) // Null terminator for Value

	// Serialize each node
	for _, node := range smr.Nodes {
		buf.WriteString(node.ID)
		buf.WriteByte(0) // Null terminator for ID

		buf.WriteString(node.IP)
		buf.WriteByte(0) // Null terminator for IP

		// Serialize Port as 4 bytes (BigEndian format)
		portBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(portBytes, uint32(node.Port))
		buf.Write(portBytes)
	}

	return buf.Bytes()
}

// Deserialize byte array back to SuccessMessageResponse
func Deserialize(data []byte) (*SuccessMessageResponse, error) {
	smr := &SuccessMessageResponse{}
	buf := bytes.NewBuffer(data)

	// Deserialize Value (read until null terminator)
	value, err := buf.ReadString(0)
	if err != nil {
		return nil, err
	}
	smr.Value = value[:len(value)-1] // Remove the null terminator

	// Deserialize Nodes (continue reading until the buffer is empty)
	var nodes []*KNode
	for buf.Len() > 0 {
		node := &KNode{}

		// Deserialize ID
		id, err := buf.ReadString(0)
		if err != nil {
			return nil, err
		}
		node.ID = id[:len(id)-1] // Remove the null terminator

		// Deserialize IP
		ip, err := buf.ReadString(0)
		if err != nil {
			return nil, err
		}
		node.IP = ip[:len(ip)-1] // Remove the null terminator

		// Deserialize Port (read 4 bytes for Port)
		portBytes := make([]byte, 4)
		if _, err := buf.Read(portBytes); err != nil {
			return nil, err
		}
		node.Port = int(binary.BigEndian.Uint32(portBytes))

		nodes = append(nodes, node)
	}

	smr.Nodes = nodes
	return smr, nil
}

// NewDHT creates a new instance of DHT.
func NewDHT(cleanup_interval time.Duration, encryptionKey []byte, id string, ip string, port int) *DHT {
	return &DHT{
		RoutingTable: NewRoutingTable(id),
		Storage:      storage.NewStorage(cleanup_interval, encryptionKey),
		Network:      message.NewNetwork(id, ip, port),
	}
}

// PUT stores a value in the DHT.
func (d *DHT) PUT(key, value string, ttl int) error {
	return d.Storage.Put(key, value, ttl)
}

// GET retrieves a value from the DHT.
func (d *DHT) GET(key string) (string, error) {
	return d.Storage.Get(key)
}

func (d *DHT) FindValue(originID string, targetKeyID string) (string, []*KNode, error) {
	value, err := d.GET(targetKeyID)
	if err != nil {
		return "", nil, err
	}

	// Found value
	if value != "" {
		return value, nil, nil
	}

	// If the value is not found in the closest nodes, perform a FindNode RPC
	nodes := d.RoutingTable.GetClosestNodes(originID, targetKeyID)
	if len(nodes) == 0 {
		return "", nil, errors.New("no nodes found")
	}

	return "", nodes, nil
}

func (d *DHT) FindNode(originID string, targetID string) ([]*KNode, error) {
	nodes := d.RoutingTable.GetClosestNodes(originID, targetID)
	if len(nodes) == 0 {
		return nil, errors.New("no nodes found")
	}

	return nodes, nil
}

func (d *DHT) IterativeFindNode(targetID string) []*KNode {
	rt := d.RoutingTable
	shortlist := rt.GetClosestNodes(rt.NodeID, targetID)
	closestNodeDistance := XOR(shortlist[0].ID, targetID)
	lastClosestNode := shortlist[0]
	queriedNodes := make(map[string]bool)

	for len(shortlist) > 0 {
		minQueryCount := min(len(shortlist), Alpha)
		alphaNodes := shortlist[:minQueryCount]
		shortlist = shortlist[minQueryCount:]

		for _, node := range alphaNodes {
			if queriedNodes[node.ID] {
				continue
			}
			queriedNodes[node.ID] = true

			msg, msgCreationErr := message.CreateMessage(message.DHT_FIND_NODE, []byte(targetID))
			rpcMessage, serializationErr := msg.Serialize()

			if msgCreationErr != nil || serializationErr != nil { //TODO handle error
				log.Printf("Error creating/serializing message: %v, %v", msgCreationErr, serializationErr)
				return nil
			}

			//TODO wait for msgResponse
			msgResponse, msgResponseErr := d.Network.SendMessage(node.IP, node.Port, rpcMessage)
			if msgResponseErr != nil { //TODO handle error
				log.Printf("Error sending message: %v", msgResponseErr)
				return nil
			}

			deserializedResponse, deserializationErr := message.DeserializeMessage(msgResponse)
			// Deserialize the response to check if it contains a value or nodes
			if deserializationErr != nil {
				log.Printf("Error deserializing response: %v", deserializationErr)
				return nil
			}

			//Check if the deserialized response has value or nodes
			switch response := deserializedResponse.(type) {
			case *message.DHTSuccessMessage:
				serializedSuccessResponse := response.Value
				smr, deserializationErr := Deserialize(serializedSuccessResponse)

				if deserializationErr != nil {
					log.Printf("Error deserializing SuccessMessageResponse: %v", deserializationErr)
					return nil
				}

				for _, node := range smr.Nodes {
					if !queriedNodes[node.ID] { //Don't add nodes that have already been queried
						shortlist = append(shortlist, node)
					}
				}
				SortNodes(shortlist, targetID)

				if len(shortlist) >= K { // TODO Not sure about this
					break
				}
			case *message.DHTFailureMessage:
				// TODO handle failure
				log.Printf("Failure message received")
			default:
				// TODO handle unknown message
				log.Printf("Unknown message type received")
			}
		}

		if XOR(shortlist[0].ID, targetID).Cmp(closestNodeDistance) >= 0 { //If the closest node is not closer than the last closest node
			if lastClosestNode.ID == shortlist[0].ID {
				break
			}
		}
	}

	return shortlist[:min(len(shortlist), K)]
}

func (d *DHT) IterativeFindValue(targetID string) (string, []*KNode) {
	rt := d.RoutingTable
	shortlist := rt.GetClosestNodes(rt.NodeID, targetID)
	closestNodeDistance := XOR(shortlist[0].ID, targetID)
	lastClosestNode := shortlist[0]
	queriedNodes := make(map[string]bool)

	for len(shortlist) > 0 {
		minQueryCount := min(len(shortlist), Alpha)
		alphaNodes := shortlist[:minQueryCount]
		shortlist = shortlist[minQueryCount:]

		for _, node := range alphaNodes {
			if queriedNodes[node.ID] {
				continue
			}
			queriedNodes[node.ID] = true

			msg, msgCreationErr := message.CreateMessage(message.DHT_FIND_VALUE, []byte(targetID))
			rpcMessage, serializationErr := msg.Serialize()

			if msgCreationErr != nil || serializationErr != nil { //TODO handle error
				log.Printf("Error creating/serializing message: %v, %v", msgCreationErr, serializationErr)
				return "", nil
			}

			//TODO wait for msgResponse
			msgResponse, msgResponseErr := d.Network.SendMessage(node.IP, node.Port, rpcMessage)
			if msgResponseErr != nil { //TODO handle error
				log.Printf("Error sending message: %v", msgResponseErr)
				return "", nil
			}

			deserializedResponse, deserializationErr := message.DeserializeMessage(msgResponse)
			// Deserialize the response to check if it contains a value or nodes
			if deserializationErr != nil {
				log.Printf("Error deserializing response: %v", deserializationErr)
				return "", nil
			}

			//Check if the deserialized response has value or nodes
			switch response := deserializedResponse.(type) {
			case *message.DHTSuccessMessage:
				serializedSuccessResponse := response.Value
				smr, deserializationErr := Deserialize(serializedSuccessResponse)

				if deserializationErr != nil {
					log.Printf("Error deserializing SuccessMessageResponse: %v", deserializationErr)
					return "", nil
				}

				if smr.Value != "" {
					return smr.Value, nil
				}

				for _, node := range smr.Nodes {
					if !queriedNodes[node.ID] { //Don't add nodes that have already been queried
						shortlist = append(shortlist, node)
					}
				}
				SortNodes(shortlist, targetID)

				if len(shortlist) >= K { // TODO Not sure about this
					break
				}
			case *message.DHTFailureMessage:
				// TODO handle failure
				log.Printf("Failure message received")
			default:
				// TODO handle unknown message
				log.Printf("Unknown message type received")
			}
		}

		if XOR(shortlist[0].ID, targetID).Cmp(closestNodeDistance) >= 0 { //If the closest node is not closer than the last closest node
			if lastClosestNode.ID == shortlist[0].ID {
				break
			}
		}
	}

	return "", shortlist[:min(len(shortlist), K)]
}

// mock rpc call
func (kn *KNode) FindNodeRPC(targetID string) []*KNode {
	return []*KNode{}
}

func (rt *RoutingTable) IterativeFindValue2(targetID string) (string, []*KNode) {
	shortlist := rt.GetClosestNodes(rt.NodeID, targetID)
	closestNodeDistance := XOR(shortlist[0].ID, targetID)
	lastClosestNode := shortlist[0]
	queriedNodes := make(map[string]bool)

	for len(shortlist) > 0 {
		minQueryCount := min(len(shortlist), Alpha)
		alphaNodes := shortlist[:minQueryCount]
		shortlist = shortlist[minQueryCount:]

		for _, node := range alphaNodes {
			if queriedNodes[node.ID] {
				continue
			}
			queriedNodes[node.ID] = true

			value, foundNodes := node.FindValueRPC(targetID) //Simulate RPC call

			if value != "" {
				return value, nil
			}

			for _, foundNode := range foundNodes {
				if !queriedNodes[foundNode.ID] { //Don't add nodes that have already been queried
					shortlist = append(shortlist, foundNode)
				}
			}
			SortNodes(shortlist, targetID)

			if len(shortlist) >= K { // TODO Not sure about this
				break
			}
		}

		if XOR(shortlist[0].ID, targetID).Cmp(closestNodeDistance) >= 0 { //If the closest node is not closer than the last closest node
			if lastClosestNode.ID == shortlist[0].ID {
				break
			}
		}
	}

	return "", shortlist[:min(len(shortlist), K)]
}

// mock rpc call
func (kn *KNode) FindValueRPC(targetID string) (string, []*KNode) {
	return "", []*KNode{}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TODO we dont need join leave here?
// Join allows the node to join the DHT network.
func (d *DHT) Join() {
	// Mock implementation
}

// Leave gracefully leaves the DHT network.
func (d *DHT) Leave() error {
	// Mock implementation
	return nil
}


// Stop gracefully shuts down the DHT and cleanup routines.
func (d *DHT) Stop() error {
	// Ensure the stop is called only once
	d.stopOnce.Do(func() {
		d.Storage.StopCleanup()
		log.Println("DHT node stopped gracefully.")
	})
	return nil
}
