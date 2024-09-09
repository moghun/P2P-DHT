package dht

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

// DHT represents a Distributed Hash Table.
type DHT struct {
	RoutingTable *RoutingTable
	Storage      *storage.Storage
	Network      *message.Network
	stopOnce sync.Once
}

// NewDHT creates a new instance of DHT.
func NewDHT(cleanup_interval time.Duration, encryptionKey []byte, id string, ip string, port int) *DHT {
	return &DHT{
		RoutingTable: NewRoutingTable(id),
		Storage:      storage.NewStorage(cleanup_interval, encryptionKey),
		Network:      message.NewNetwork(id, ip, port),
	}
}

// HashKey hashes the given key and returns the hash as a hex string.
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	hexEncodedHash := hex.EncodeToString(hash[:20])
	// util.Log().Print("Hashed key length: ", len(hexEncodedHash))
	// decodedHash, _ := message.HexStringToByte32(hexEncodedHash)
	// util.Log().Print("Decoded hash length: ", len(decodedHash))
	return hexEncodedHash // Use the first 160 bits (20 bytes) of the hash
}

// IsHashedKey checks if the provided key is already a hashed key in hex string format.
func IsHashedKey(key string) bool {
	// Hex string of 160-bit (20 bytes) hash should be 40 characters long
	const hexPattern = "^[a-fA-F0-9]{40}$"
	matched, _ := regexp.MatchString(hexPattern, key)
	return matched
}

// EnsureKeyHashed returns the hashed key as a hex string if the provided key is not already hashed.
func EnsureKeyHashed(key string) string {
	if IsHashedKey(key) {
		return key // Already hashed
	}
	return HashKey(key) // Hash the key
}

// PUT stores a value in the DHT.
func (d *DHT) PUT(key, value string, ttl int) error {
	key = EnsureKeyHashed(key)
	nodesToStore, err := d.FindNode(key)
	if err != nil {
		return err
	}

	failedNodes := make(map[string]*KNode)

	if len(nodesToStore) == 0 {
		util.Log().Print("No nodes found to store the data, storing on this node")
		err = d.StoreToStorage(key, value, ttl)
		if err != nil {
			return err
		}
		return nil
	}
	for _, node := range nodesToStore {
		response, err := d.SendStoreMessage(key, value, *node)

		util.Log().Infof("Response: %v", response)

		if err != nil {
			util.Log().Errorf("Error sending message: %v", err)
			return err
		}

		if response.GetType() == message.DHT_FAILURE {
			failedNodes[node.ID] = node
		}
	}

	for i := 0; i < RetryLimit; i++ { // Retry failed nodes
		for _, node := range failedNodes {
			response, err := d.SendStoreMessage(key, value, *node)

			if err != nil {
				util.Log().Errorf("Error sending retry message: %v", err)
				return err
			}

			if response.GetType() == message.DHT_SUCCESS {
				delete(failedNodes, node.ID)
			}
		}
	}

	if len(failedNodes) > 0 {
		util.Log().Infof("Failed nodes: %v", failedNodes)
	}

	return nil
}

func (d *DHT) CreateStoreMessage(key, value string) ([]byte, error) {
	byte32Key, err := message.HexStringToByte32(key)
	if err != nil {
		util.Log().Errorf("Error converting key to byte32: %v", err)
		return nil, err
	}
	msg := message.NewDHTStoreMessage(10000, 2, byte32Key, []byte(value))
	rpcMessage, serializationErr := msg.Serialize()

	if serializationErr != nil { //TODO handle error
		util.Log().Errorf("Error serializing message: %v", serializationErr)
		return nil, serializationErr
	}

	return rpcMessage, nil
}

func (d *DHT) SendStoreMessage(key, value string, targetNode KNode) (message.Message, error) {
	rpcMessage, err := d.CreateStoreMessage(key, value)

	if err != nil {
		util.Log().Errorf("Error creating/serializing message: %v", err)
		return nil, err
	}

	responseChan := make(chan []byte)
	errorChan := make(chan error)

	// Send message asynchronously
	go func() {
		util.Log().Print("sending this from dht.sendStoreMessage: ", rpcMessage)
		msgResponse, err := d.Network.SendMessage(targetNode.IP, targetNode.Port, rpcMessage)
		util.Log().Print("dht.sendStoreMessage response: ", msgResponse)
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- msgResponse
	}()

	select {
	case msgResponse := <-responseChan:
		// Deserialize the response
		deserializedResponse, deserializationErr := message.DeserializeMessage(msgResponse)
		if deserializationErr != nil {
			util.Log().Errorf("Error deserializing response: %v", deserializationErr)
			return nil, deserializationErr
		}
		return deserializedResponse, nil

	case msgResponseErr := <-errorChan:
		util.Log().Errorf("Error sending message: %v", msgResponseErr)
		return nil, msgResponseErr
	}
}

// GET retrieves a value from the DHT.
func (d *DHT) GET(key string) (string, []*KNode, error) {
	value, nodes, err := d.FindValue(key)

	if err != nil {
		return "", nil, err
	}

	if value != "" {
		return value, nil, nil
	}

	if nodes == nil {
		return "", nil, errors.New("no nodes found")
	}

	return "", nodes, nil
}

func (d *DHT) GetFromStorage(targetKeyID string) (string, error) {
	value, err := d.Storage.Get(targetKeyID)
	util.Log().Print("Getting from storage: ", value)
	if err != nil {
		return "", err
	}

	if value != "" {
		return value, nil
	} else {
		return "", nil
	}

	//return "", errors.New("unexpected return path")
}

func (d *DHT) StoreToStorage(key, value string, ttl int) error {
	util.Log().Print("Storing to storage")
	key = EnsureKeyHashed(key)
	return d.Storage.Put(key, value, ttl)
}

func (d *DHT) FindValue(targetKeyID string) (string, []*KNode, error) {
	value, err := d.GetFromStorage(targetKeyID)

	if err != nil {
		return "", nil, err
	}

	// Found value
	if value != "" {
		return value, nil, nil
	}

	// If the value is not found in the closest nodes, perform a FindNode RPC
	nodes, err := d.RoutingTable.GetClosestNodes(targetKeyID)
	if err != nil {
		return "", nil, err
	}
	if len(nodes) == 0 {
		return "", nil, nil
	}

	return "", nodes, nil
}

func (d *DHT) FindNode(targetID string) ([]*KNode, error) {
	nodes, err := d.RoutingTable.GetClosestNodes(targetID)
	util.Log().Print("FindNode nodes count: ", len(nodes))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, nil
	}

	return nodes, nil
}

func (d *DHT) IterativeFindNode(targetID string) ([]*KNode, error) {
	rt := d.RoutingTable
	shortlist, err := rt.GetClosestNodes(targetID)
	if err != nil {
		return nil, err
	}
	closestNodeDistance, _ := XOR(shortlist[0].ID, targetID)
	//lastClosestNode := shortlist[0]
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

			byte32Key, err := message.HexStringToByte32(targetID)
			if err != nil {
				util.Log().Errorf("Error converting key to byte32: %v", err)
				return nil, err
			}
			msg := message.NewDHTFindNodeMessage(byte32Key)
			rpcMessage, serializationErr := msg.Serialize()

			if serializationErr != nil { //TODO handle error
				util.Log().Errorf("Error creating/serializing message: %v", serializationErr)
				return nil, serializationErr
			}

			//TODO wait for msgResponse
			msgResponse, msgResponseErr := d.Network.SendMessage(node.IP, node.Port, rpcMessage)
			if msgResponseErr != nil { //TODO handle error
				util.Log().Errorf("Error sending message: %v", msgResponseErr)
				return nil, msgResponseErr
			}

			deserializedResponse, deserializationErr := message.DeserializeMessage(msgResponse)
			// Deserialize the response to check if it contains a value or nodes
			if deserializationErr != nil {
				util.Log().Errorf("Error deserializing response: %v", deserializationErr)
				return nil, deserializationErr
			}

			//Check if the deserialized response has value or nodes
			switch response := deserializedResponse.(type) {
			case *message.DHTSuccessMessage:
				serializedSuccessResponse := response.Value
				smr, deserializationErr := Deserialize(serializedSuccessResponse)

				if deserializationErr != nil {
					util.Log().Errorf("Error deserializing SuccessMessageResponse: %v", deserializationErr)
					return nil, deserializationErr
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
				util.Log().Info("Failure message received")
			default:
				// TODO handle unknown message
				util.Log().Info("Unknown message type received")
			}
		}

		xorResult, _ := XOR(shortlist[0].ID, targetID)
		//break if the closest node is not closer than the last closest node
		if xorResult >= closestNodeDistance {
			break
		}
	}

	return shortlist[:min(len(shortlist), K)], nil
}

func (d *DHT) IterativeFindValue(targetID string) (string, []*KNode, error) {
	rt := d.RoutingTable
	shortlist, err := rt.GetClosestNodes(targetID)
	if err != nil {
		return "", nil, err
	}
	closestNodeDistance, _ := XOR(shortlist[0].ID, targetID)
	//lastClosestNode := shortlist[0]
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

			byte32Key, err := message.HexStringToByte32(targetID)
			if err != nil {
				util.Log().Errorf("Error converting key to byte32: %v", err)
				return "", nil, err
			}
			msg := message.NewDHTFindValueMessage(byte32Key)
			rpcMessage, serializationErr := msg.Serialize()

			if serializationErr != nil { //TODO handle error
				util.Log().Errorf("Error serializing message: %v", serializationErr)
				return "", nil, serializationErr
			}

			//TODO wait for msgResponse
			msgResponse, msgResponseErr := d.Network.SendMessage(node.IP, node.Port, rpcMessage)
			if msgResponseErr != nil { //TODO handle error
				util.Log().Errorf("Error sending message: %v", msgResponseErr)
				return "", nil, msgResponseErr
			}

			deserializedResponse, deserializationErr := message.DeserializeMessage(msgResponse)
			// Deserialize the response to check if it contains a value or nodes
			if deserializationErr != nil {
				util.Log().Errorf("Error deserializing response: %v", deserializationErr)
				return "", nil, deserializationErr
			}

			//Check if the deserialized response has value or nodes
			switch response := deserializedResponse.(type) {
			case *message.DHTSuccessMessage:
				serializedSuccessResponse := response.Value
				smr, deserializationErr := Deserialize(serializedSuccessResponse)

				if deserializationErr != nil {
					util.Log().Errorf("Error deserializing SuccessMessageResponse: %v", deserializationErr)
					return "", nil, deserializationErr
				}

				if smr.Value != "" {
					return smr.Value, nil, nil
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
				util.Log().Info("Failure message received")
			default:
				// TODO handle unknown message
				util.Log().Info("Unknown message type received")
			}
		}

		xorResult, _ := XOR(shortlist[0].ID, targetID)
		//break if the closest node is not closer than the last closest node
		if xorResult >= closestNodeDistance {
			break
		}
	}

	return "", shortlist[:min(len(shortlist), K)], nil
}

// IterativeStore stores a value in the DHT.
func (d *DHT) IterativeStore(key, value string) error {
	nodesToPublishData, err := d.IterativeFindNode(key)
	if err != nil {
		return err
	}
	if nodesToPublishData == nil {
		return errors.New("no nodes found")
	}
	failedNodes := make(map[string]*KNode)
	for _, node := range nodesToPublishData {

		response, err := d.SendStoreMessage(key, value, *node)

		if err != nil {
			util.Log().Errorf("Error sending retry message: %v", err)
			return err
		}

		if response.GetType() == message.DHT_SUCCESS {
			delete(failedNodes, node.ID)
		}
	}

	for i := 0; i < RetryLimit; i++ { // Retry failed nodes
		for _, node := range failedNodes {
			response, err := d.SendStoreMessage(key, value, *node)

			if err != nil {
				util.Log().Errorf("Error sending retry message: %v", err)
				return err
			}

			if response.GetType() == message.DHT_SUCCESS {
				delete(failedNodes, node.ID)
			}
		}
	}

	if len(failedNodes) > 0 {
		util.Log().Infof("Failed nodes: %v", failedNodes)
	}

	return nil
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
		util.Log().Println("DHT node stopped gracefully.")
	})
	return nil
}
