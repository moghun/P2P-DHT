package node

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

type NodeInterface interface {
	Put(key, value string, ttl int) error
	Get(key string) (string, error)
	AddPeer(nodeID, ip string, port int)
	GetAllPeers() []*Node
}


type Node struct {
	ID       string
	IP       string
	Port     int
	Ping     bool
	Storage  *storage.Storage
	Network  NetworkInterface
	Config   *util.Config
	IsDown   bool
	mu       sync.Mutex
}

func NewNode(config *util.Config, ttl time.Duration) *Node {

	ip, port, _ := util.ParseAddress(config.P2PAddress)

	node := &Node{
		ID:       GenerateNodeID(ip, port),
		IP:       ip,
		Port:     port,
		Ping:     true,
		Storage:  storage.NewStorage(ttl, config.EncryptionKey),
		IsDown:   false,
		Config:   config,  // Set the configuration
	}

	node.Network = NewNetwork(node)

	return node
}

// GenerateNodeID generates a unique ID for the node based on its IP and port.
func GenerateNodeID(ip string, port int) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d", ip, port)))
	return hex.EncodeToString(h.Sum(nil))
}

// Put stores a key-value pair in the node's storage with a specified TTL.
func (n *Node) Put(key, value string, ttl int) error {
	return n.Storage.Put(key, value, ttl)
}

// Get retrieves a value from the node's storage based on the key.
func (n *Node) Get(key string) (string, error) {
	value, err := n.Storage.Get(key)
	if err != nil {
		return "", err
	}
	return value, nil
}

// AddPeer is a placeholder for adding a peer to the node's routing table (mocked for now).
func (n *Node) AddPeer(nodeID, ip string, port int) {
}

// RemovePeer is a placeholder for removing a peer from the node's routing table (mocked for now).
func (n *Node) RemovePeer(ip string, port int) {
}

// GetAllPeers is a placeholder for retrieving all peers from the node's routing table (mocked for now).
func (n *Node) GetAllPeers() []*Node {
	return nil
}

// GetClosestNodesToCurrNode is a placeholder for getting the closest nodes to the current node (mocked for now).
func (n *Node) GetClosestNodesToCurrNode(targetID string, k int) []*Node {
	return nil
}
