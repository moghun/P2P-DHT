package dht

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/bits"
	"sort"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

type Node struct {
	ID       string
	IP       string
	Port     int
	Ping     bool
	Storage  *storage.Storage
	KBuckets []*KBucket
	IsDown   bool
	mu       sync.Mutex
	CertFile string
	KeyFile  string
}

func NewNode(ip string, port int, ping bool, key []byte) *Node {
	ttl := 24 * time.Hour

	node := &Node{
		ID:       GenerateNodeID(ip, port),
		IP:       ip,
		Port:     port,
		Ping:     ping,
		Storage:  storage.NewStorage(ttl, key),
		IsDown:   false,
		KBuckets: make([]*KBucket, 160),
		CertFile: "",
		KeyFile:  "",
	}


	for i := range node.KBuckets {
		node.KBuckets[i] = NewKBucket()
	}

	err := node.SetupTLS()
	if err != nil {
		fmt.Printf("Error setting up TLS: %v\n", err)
		return nil
	}
	return node
}

func GenerateNodeID(ip string, port int) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d", ip, port)))
	return hex.EncodeToString(h.Sum(nil))
}

func (n *Node) SetupTLS() error {
	certFile, keyFile, err := util.GenerateCertificates(n.IP, n.Port)
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}

	n.CertFile = certFile
	n.KeyFile = keyFile

	return nil
}

func (n *Node) AddPeer(nodeID, ip string, port int) {
	distance := calculateDistance(n.ID, nodeID)
	bucketIndex := getBucketIndex(distance)
	peerNode := &Node{ID: nodeID, IP: ip, Port: port}

	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.KBuckets[bucketIndex].Contains(peerNode) {
		n.KBuckets[bucketIndex].AddNode(peerNode)
	}
}

func (n *Node) RemovePeer(ip string, port int) {
	nodeID := GenerateNodeID(ip, port)
	distance := calculateDistance(n.ID, nodeID)
	bucketIndex := getBucketIndex(distance)

	n.mu.Lock()
	defer n.mu.Unlock()

	n.KBuckets[bucketIndex].RemoveNode(nodeID)
}

func (n *Node) GetAllPeers() []*Node {
	n.mu.Lock()
	defer n.mu.Unlock()

	var peers []*Node
	for _, bucket := range n.KBuckets {
		peers = append(peers, bucket.GetNodes()...)
	}
	return peers
}

func (n *Node) Put(key, value string, ttl int) error {
	return n.Storage.Put(key, value, ttl)
}

func (n *Node) Get(key string) (string, error) {
	value, err := n.Storage.Get(key)
	if err != nil {
		return "", err
	}
	return value, nil
}

func calculateDistance(id1, id2 string) uint64 {
	hash1, _ := hex.DecodeString(id1)
	hash2, _ := hex.DecodeString(id2)

	var distance uint64
	for i := range hash1 {
		distance += uint64(hash1[i] ^ hash2[i])
	}
	return distance
}

func getBucketIndex(distance uint64) int {
	if distance == 0 {
		return 0
	}
	return 159 - bits.LeadingZeros64(distance)
}

func (n *Node) GetClosestNodesToCurrNode(targetID string, k int) []*Node {
	n.mu.Lock()
	defer n.mu.Unlock()

	var allNodes []*Node
	for _, bucket := range n.KBuckets {
		allNodes = append(allNodes, bucket.GetNodes()...)
	}

	sort.Slice(allNodes, func(i, j int) bool {
		return calculateDistance(targetID, allNodes[i].ID) < calculateDistance(targetID, allNodes[j].ID)
	})

	if len(allNodes) > k {
		return allNodes[:k]
	}
	return allNodes
}

func (n *Node) JoinNetwork(dhtInstance *DHT) {
	dhtInstance.JoinNetwork(n)
}

func (n *Node) LeaveNetwork(dhtInstance *DHT) error {
	return dhtInstance.LeaveNetwork(n)
}