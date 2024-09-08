package dht

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

// DHT represents a Distributed Hash Table.
type DHT struct {
	RoutingTable *RoutingTable
	Storage      *DHTStorage
	Network      *message.Network
}

// NewDHT creates a new instance of DHT.
func NewDHT(ttl time.Duration, encryptionKey []byte, id string, ip string, port int) *DHT {
	return &DHT{
		RoutingTable: NewRoutingTable(id),
		Storage:      NewDHTStorage(ttl, encryptionKey),
		Network:      message.NewNetwork(id, ip, port),
	}
}

// PUT stores a value in the DHT.
func (d *DHT) PUT(key, value string, ttl int) error {
	nodesToStore, err := d.FindNode(key)
	if err != nil {
		return err
	}

	failedNodes := make(map[string]*KNode)

	for _, node := range nodesToStore {
		response, err := d.SendStoreMessage(key, value, *node)

		log.Printf("Response: %v", response)

		if err != nil {
			log.Printf("Error sending message: %v", err)
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
				log.Printf("Error sending retry message: %v", err)
				return err
			}

			if response.GetType() == message.DHT_SUCCESS {
				delete(failedNodes, node.ID)
			}
		}
	}

	if len(failedNodes) > 0 {
		log.Printf("Failed nodes: %v", failedNodes)
	}

	return nil
}

func (d *DHT) CreateStoreMessage(key, value string) ([]byte, error) {
	msg := message.NewDHTStoreMessage(10000, 2, message.StringToByte32(key), []byte(value))
	rpcMessage, serializationErr := msg.Serialize()

	if serializationErr != nil { //TODO handle error
		log.Printf("Error serializing message: %v", serializationErr)
		return nil, serializationErr
	}

	return rpcMessage, nil
}

func (d *DHT) SendStoreMessage(key, value string, targetNode KNode) (message.Message, error) {
	rpcMessage, err := d.CreateStoreMessage(key, value)

	if err != nil {
		log.Printf("Error creating/serializing message: %v", err)
		return nil, err
	}

	responseChan := make(chan []byte)
	errorChan := make(chan error)

	// Send message asynchronously
	go func() {
		log.Print("sending this from dht.sendStoreMessage: ", rpcMessage)
		msgResponse, err := d.Network.SendMessage(targetNode.IP, targetNode.Port, rpcMessage)
		log.Print("dht.sendStoreMessage response: ", msgResponse)
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
			log.Printf("Error deserializing response: %v", deserializationErr)
			return nil, deserializationErr
		}
		return deserializedResponse, nil

	case msgResponseErr := <-errorChan:
		log.Printf("Error sending message: %v", msgResponseErr)
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
	if err != nil {
		return "", err
	}

	// Found value
	if value != "" {
		return value, nil
	}

	return "", errors.New("unexpected return path")
}

func (d *DHT) StoreToStorage(key, value string, ttl int) error {
	log.Print("Storing to storage")
	return d.Storage.Put(key, value, ttl)
}

func (d *DHT) FindValue(targetKeyID string) (string, []*KNode, error) {
	value, err := d.GetFromStorage(message.StringTo32ByteString(targetKeyID))

	if err != nil {
		return "", nil, err
	}

	// Found value
	if value != "" {
		return value, nil, nil
	}

	// If the value is not found in the closest nodes, perform a FindNode RPC
	nodes := d.RoutingTable.GetClosestNodes(targetKeyID)
	if len(nodes) == 0 {
		return "", nil, errors.New("no nodes found")
	}

	return "", nodes, nil
}

func (d *DHT) FindNode(targetID string) ([]*KNode, error) {
	nodes := d.RoutingTable.GetClosestNodes(targetID)
	if len(nodes) == 0 {
		return nil, errors.New("no nodes found")
	}

	return nodes, nil
}

func (d *DHT) IterativeFindNode(targetID string) []*KNode {
	rt := d.RoutingTable
	shortlist := rt.GetClosestNodes(targetID)
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

			msg := message.NewDHTFindNodeMessage(message.StringToByte32(targetID))
			rpcMessage, serializationErr := msg.Serialize()

			if serializationErr != nil { //TODO handle error
				log.Printf("Error creating/serializing message: %v", serializationErr)
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
	shortlist := rt.GetClosestNodes(targetID)
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

			msg := message.NewDHTFindValueMessage(message.StringToByte32(targetID))
			rpcMessage, serializationErr := msg.Serialize()

			if serializationErr != nil { //TODO handle error
				log.Printf("Error serializing message: %v", serializationErr)
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

// IterativeStore stores a value in the DHT.
func (d *DHT) IterativeStore(key, value string) error {
	nodesToPublishData := d.IterativeFindNode(key)
	if nodesToPublishData == nil {
		return errors.New("no nodes found")
	}
	failedNodes := make(map[string]*KNode)
	for _, node := range nodesToPublishData {

		response, err := d.SendStoreMessage(key, value, *node)

		if err != nil {
			log.Printf("Error sending retry message: %v", err)
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
				log.Printf("Error sending retry message: %v", err)
				return err
			}

			if response.GetType() == message.DHT_SUCCESS {
				delete(failedNodes, node.ID)
			}
		}
	}

	if len(failedNodes) > 0 {
		log.Printf("Failed nodes: %v", failedNodes)
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

// DHTStorage is a simple in-memory key-value store with TTL and encryption.
type DHTStorage struct {
	data map[string]*storageItem
	mu   sync.Mutex
	ttl  time.Duration
	key  []byte // Encryption key
}

type storageItem struct {
	value  string
	expiry time.Time
	hash   string
}

// NewDHTStorage initializes a new DHTStorage with a given TTL and encryption key.
func NewDHTStorage(ttl time.Duration, key []byte) *DHTStorage {
	storage := &DHTStorage{
		data: make(map[string]*storageItem),
		ttl:  ttl,
		key:  key,
	}
	storage.StartCleanup(ttl)
	return storage
}

// Put stores a value with the specified TTL.
func (s *DHTStorage) Put(key, value string, ttl int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encryptedValue, err := util.Encrypt([]byte(value), s.key)
	if err != nil {
		log.Printf("Error encrypting value: %v", err)
		return err
	}

	hasher := sha1.New()
	hasher.Write([]byte(encryptedValue))
	hash := hex.EncodeToString(hasher.Sum(nil))

	expiry := time.Now().Add(time.Duration(ttl) * time.Second)
	s.data[key] = &storageItem{
		value:  encryptedValue,
		expiry: expiry,
		hash:   hash,
	}
	return nil
}

// Get retrieves the value associated with the key if it exists and hasn't expired.
func (s *DHTStorage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return "", nil
	}

	if time.Now().After(item.expiry) {
		delete(s.data, key)
		return "", nil
	}

	hasher := sha1.New()
	hasher.Write([]byte(item.value))
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != item.hash {
		return "", errors.New("data integrity check failed")
	}

	decryptedValue, err := util.Decrypt(item.value, s.key)
	if err != nil {
		return "", err
	}

	return string(decryptedValue), nil
}

// StartCleanup starts a routine that periodically cleans up expired items.
func (s *DHTStorage) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.CleanupExpired()
		}
	}()
}

// CleanupExpired removes expired items from the storage.
func (s *DHTStorage) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.data {
		if time.Now().After(item.expiry) {
			delete(s.data, key)
		}
	}
}
