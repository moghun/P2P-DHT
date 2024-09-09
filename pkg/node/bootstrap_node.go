package node

import (
	"fmt"
	"sync"
	"time"

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

func (bn *BootstrapNode) Shutdown() {
	// Gracefully shut down the embedded Node
	bn.Node.Shutdown()

	// Any additional shutdown tasks for BootstrapNode can be added here
	bn.mu.Lock()
	bn.KnownPeers = make(map[string]string) // Clear known peers
	bn.mu.Unlock()

	util.Log().Println("BootstrapNode shut down successfully.")
}
