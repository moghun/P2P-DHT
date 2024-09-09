package node

import (
	"fmt"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

// BootstrapNode represents a special type of node that knows about other peers in the network.
type BootstrapNode struct {
	Node       // Embedding Node struct to inherit its fields and methods
	KnownPeers map[string]string // Dictionary to store known peers as "IP:Port" -> "NodeID"
	mu         sync.Mutex        // Mutex for thread-safe operations on KnownPeers
}

// NewBootstrapNode creates a new BootstrapNode with the given configuration and TTL.
func NewBootstrapNode(config *util.Config, cleanup_interval time.Duration) *BootstrapNode {
	return &BootstrapNode{
		Node:       *NewNode(config, cleanup_interval),
		KnownPeers: make(map[string]string),
	}
}

// Implement NodeInterface methods explicitly

// Put stores a key-value pair in the node's storage with a specified TTL.
func (bn *BootstrapNode) Put(key, value string, ttl int) error {
	return bn.Node.Put(key, value, ttl)
}

// Get retrieves a value from the node's storage based on the key.
func (bn *BootstrapNode) Get(key string) (string, error) {
	return bn.Node.Get(key)
}

// FindNode is a wrapper to call the embedded Node's FindNode method.
func (bn *BootstrapNode) FindNode(targetID string) ([]*dht.KNode, error) {
	return bn.Node.FindNode(targetID)
}

// FindValue is a wrapper to call the embedded Node's FindValue method.
func (bn *BootstrapNode) FindValue(targetKeyID string) (string, []*dht.KNode, error) {
	return bn.Node.FindValue(targetKeyID)
}

// AddPeer adds a peer to the bootstrap node.
func (bn *BootstrapNode) AddPeer(nodeID, ip string, port int) {
	bn.Node.AddPeer(nodeID, ip, port)
}

// GetAllPeers is a wrapper to call the embedded Node's GetAllPeers method.
func (bn *BootstrapNode) GetAllPeers() []*dht.KNode {
	return bn.Node.GetAllPeers()
}

// GetID returns the bootstrap node's ID.
func (bn *BootstrapNode) GetID() string {
	return bn.Node.GetID()
}

// Additional BootstrapNode-specific methods

// AddKnownPeer adds a peer to the known peers list.
func (bn *BootstrapNode) AddKnownPeer(nodeID, ip string, port int) {
	bn.mu.Lock()
	defer bn.mu.Unlock()
	address := fmt.Sprintf("%s:%d", ip, port)
	bn.KnownPeers[address] = nodeID
}

// RemoveKnownPeer removes a peer from the known peers list.
func (bn *BootstrapNode) RemoveKnownPeer(ip string, port int) {
	bn.mu.Lock()
	defer bn.mu.Unlock()
	address := fmt.Sprintf("%s:%d", ip, port)
	delete(bn.KnownPeers, address)
}

// GetKnownPeers returns the list of known peers.
func (bn *BootstrapNode) GetKnownPeers() map[string]string {
	bn.mu.Lock()
	defer bn.mu.Unlock()
	// Return a copy of the map to avoid external modification
	peersCopy := make(map[string]string)
	for k, v := range bn.KnownPeers {
		peersCopy[k] = v
	}
	return peersCopy
}

// Shutdown shuts down the BootstrapNode.
func (bn *BootstrapNode) Shutdown() {
	// Gracefully shut down the embedded Node
	bn.Node.Shutdown()

	// Additional shutdown tasks for BootstrapNode
	bn.mu.Lock()
	bn.KnownPeers = make(map[string]string) // Clear known peers
	bn.mu.Unlock()

	util.Log().Println("BootstrapNode shut down successfully.")
}