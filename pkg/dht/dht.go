package dht

import (
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

// DHT represents a Distributed Hash Table.
type DHT struct {
	Node         *node.Node
	RoutingTable *RoutingTable
}

// NewDHT creates a new instance of DHT.
func NewDHT(nodeInstance *node.Node) *DHT {
	return &DHT{
		Node:         nodeInstance,
		RoutingTable: NewRoutingTable(),
	}
}

// PUT stores a value in the DHT.
func (d *DHT) PUT(key, value string, ttl int) error {
	// Mock implementation
	return nil
}

// GET retrieves a value from the DHT.
func (d *DHT) GET(key string) (string, error) {
	// Mock implementation
	return "", nil
}

// Join allows the node to join the DHT network.
func (d *DHT) Join() {
	// Mock implementation
}

// Leave gracefully leaves the DHT network.
func (d *DHT) Leave() error {
	// Mock implementation
	return nil
}