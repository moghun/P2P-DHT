package dht

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/bits"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
)

type Node struct {
	ID       string
	IP       string
	Port     int
	Ping     bool
	Storage  *storage.Storage
	KBuckets []*KBucket
	mu       sync.Mutex
}

func NewNode(ip string, port int, ping bool) *Node {
	node := &Node{
		ID:       generateNodeID(ip, port),
		IP:       ip,
		Port:     port,
		Ping:     ping,
		Storage:  storage.NewStorage(24 * time.Hour), // TTL
		KBuckets: make([]*KBucket, 160),
	}

	for i := range node.KBuckets {
		node.KBuckets[i] = NewKBucket()
	}

	return node
}

func generateNodeID(ip string, port int) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d", ip, port)))
	return hex.EncodeToString(h.Sum(nil))
}



func (n *Node) generateCertificates(certsDir string) {
	keyFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.key", n.IP, n.Port))
	csrFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.csr", n.IP, n.Port))
	certFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.crt", n.IP, n.Port))

	exec.Command("openssl", "genpkey", "-algorithm", "RSA", "-out", keyFile).Run()
	exec.Command("openssl", "req", "-new", "-key", keyFile, "-out", csrFile, "-subj", fmt.Sprintf("/CN=%d", n.Port)).Run()
	caDir := filepath.Join("certificates", "CA")
	exec.Command("openssl", "x509", "-req", "-in", csrFile, "-CA", filepath.Join(caDir, "ca.pem"), "-CAkey", filepath.Join(caDir, "ca.key"), "-CAcreateserial", "-out", certFile, "-days", "365").Run()
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
	nodeID := generateNodeID(ip, port)
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

func (n *Node) Put(key, value string, ttl int) {
	n.Storage.Put(key, value, ttl)
}

func (n *Node) Get(key string) string {
	return n.Storage.Get(key)
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

func (n *Node) GetClosestNodes(targetID string, k int) []*Node {
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