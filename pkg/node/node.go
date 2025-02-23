package node

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

type NodeInterface interface {
	Put(key, value string, ttl int) error
	Get(key string) (string, error)
	FindNode(targetID string) ([]*dht.KNode, error)
	FindValue(targetKeyID string) (string, []*dht.KNode, error)
	AddPeer(nodeID, ip string, port int)
	GetAllPeers() []*dht.KNode
	GetID() string
}

type Node struct {
	ID      string
	IP      string
	Port    int
	Nonce   int
	Ping    bool
	DHT     *dht.DHT
	Storage *storage.Storage
	Network message.NetworkInterface
	Config  *util.Config
	IsDown  bool
	mu      sync.Mutex
}

func NewNode(config *util.Config, cleanup_interval time.Duration) *Node {

	ip, port, _ := util.ParseAddress(config.P2PAddress)

	// Generate a node ID using Proof of Work
	id, nonce := security.GenerateNodeIDWithPoW(ip, port, config.Difficulty)

	node := &Node{
		ID:      id,
		IP:      ip,
		Port:    port,
		Nonce:   nonce,
		Ping:    true,
		DHT:     dht.NewDHT(cleanup_interval, config.EncryptionKey, id, ip, port),
		Storage: storage.NewStorage(cleanup_interval, config.EncryptionKey),
		IsDown:  false,
		Config:  config, // Set the configuration
	}

	node.Network = message.NewNetwork(ip, id, port)

	// Start the DHT network join process
	go node.DHT.Join()

	return node
}

func (n *Node) Shutdown() {
	n.DHT.Stop()
	n.Storage.StopCleanup()
	n.IsDown = true
	
	util.Log().Infof("Node %s shut down successfully.", n.ID)
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

func (n *Node) FindNode(targetID string) ([]*dht.KNode, error) {
	nodes, err := n.DHT.FindNode(targetID)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (n *Node) FindValue(targetKeyID string) (string, []*dht.KNode, error) {
	return n.DHT.FindValue(targetKeyID)
}

func (n *Node) GetID() string {
	return n.ID
}

// AddPeer is a placeholder for adding a peer to the node's routing table (mocked for now).
func (n *Node) AddPeer(nodeID, ip string, port int) {
	// Validate the node ID using PoW before adding the peer
	if security.ValidateNodeIDWithPoW(ip, port, nodeID, n.Nonce, n.Config.Difficulty) {
		kNode := &dht.KNode{ID: nodeID, IP: ip, Port: port}
		n.DHT.RoutingTable.AddNode(kNode)
	}
}

// RemovePeer is a placeholder for removing a peer from the node's routing table (mocked for now).
func (n *Node) RemovePeer(id string) {
	n.DHT.RoutingTable.RemoveNode(id)
}

// GetAllPeers is a placeholder for retrieving all peers from the node's routing table (mocked for now).
func (n *Node) GetAllPeers() []*dht.KNode {
	allPeers, _ := n.DHT.RoutingTable.GetAllNodes()
	return allPeers
}

// GetClosestNodesToCurrNode is a placeholder for getting the closest nodes to the current node (mocked for now).
func (n *Node) GetClosestNodesToCurrNode(targetID string, k int) []*Node {
	return nil
}

/*FOR TEST PURPOSES*/
// GenerateNodeID generates a unique ID for the node based on its IP and port.
func GenerateNodeID(ip string, port int) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d", ip, port)))

	data := fmt.Sprintf("%s:%d", ip, port)
	hash := sha256.Sum256([]byte(data))

	// Truncate to the first 160 bits (20 bytes) of the SHA-256 hash
	truncatedHash := hash[:20]

	// Convert truncated hash to hexadecimal string
	hashStr := hex.EncodeToString(truncatedHash)

	return hashStr
}

/**************************/
