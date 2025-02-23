package dht

import (
	"bytes"
	"encoding/gob"
	"sync"
)

// KBucket represents a bucket in the Kademlia routing table.
type KBucket struct {
	Nodes []*KNode
	mutex sync.RWMutex
}

// KNode represents a node in the DHT.
type KNode struct {
	ID   string
	IP   string
	Port int
}

// Serialize serializes the KNode struct to a byte array.
func (k *KNode) Serialize() []byte { //TODO change name
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(k)
	if err != nil {
		return nil
	}
	return buffer.Bytes()
}

// Deserialize deserializes a byte array into the KNode struct.
func (k *KNode) Deserialize(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(k)
	if err != nil {
		return err
	}
	return nil
}

// NewKBucket creates a new KBucket.
func NewKBucket() *KBucket {
	return &KBucket{
		Nodes: []*KNode{},
	}
}

// AddNode adds a node to the KBucket.
func (kb *KBucket) AddNode(node *KNode) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()

	for i := 0; i < len(kb.Nodes); i++ {
		if kb.Nodes[i].ID == node.ID {
			// Node already exists in the bucket - move it to the end
			kb.Nodes = append(kb.Nodes[:i], kb.Nodes[i+1:]...)
			kb.Nodes = append(kb.Nodes, node)
			return
		}
	}

	if len(kb.Nodes) < 20 {
		kb.Nodes = append(kb.Nodes, node)
	} else {
		// Replace the first node in the bucket
		kb.Nodes = append(kb.Nodes[1:], node)
		//TODO - Liveness check to the oldest node
	}
}

// RemoveNode removes a node from the KBucket by ID.
func (kb *KBucket) RemoveNode(nodeID string) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()

	for i := 0; i < len(kb.Nodes); i++ {
		if kb.Nodes[i].ID == nodeID {
			kb.Nodes = append(kb.Nodes[:i], kb.Nodes[i+1:]...)
			return
		}
	}
}

// GetNodes returns all nodes in the KBucket.
func (kb *KBucket) GetNodes() []*KNode {
	return kb.Nodes
}
