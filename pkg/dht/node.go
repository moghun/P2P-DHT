package dht

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/bits"
	"os"
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

func NewNode(ip string, port int, ping bool, key []byte) *Node {
	ttl := 24 * time.Hour

	node := &Node{
		ID:       GenerateNodeID(ip, port),
		IP:       ip,
		Port:     port,
		Ping:     ping,
		Storage:  storage.NewStorage(ttl, key),
		KBuckets: make([]*KBucket, 160),
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
	tlsDir := fmt.Sprintf("%s_%d", n.IP, n.Port)
	certsDir := filepath.Join("certificates", tlsDir)
	caDir := filepath.Join("certificates", "CA")

	// Ensure CA directory exists
	if _, err := os.Stat(caDir); os.IsNotExist(err) {
		os.MkdirAll(caDir, os.ModePerm)
	}

	// Ensure CA files exist
	if _, err := os.Stat(filepath.Join(caDir, "ca.pem")); os.IsNotExist(err) {
		return fmt.Errorf("CA certificate not found. Please generate the CA certificate and key.")
	}
	if _, err := os.Stat(filepath.Join(caDir, "ca.key")); os.IsNotExist(err) {
		return fmt.Errorf("CA key not found. Please generate the CA certificate and key.")
	}

	if _, err := os.Stat(certsDir); os.IsNotExist(err) {
		os.MkdirAll(certsDir, os.ModePerm)
		if err := n.GenerateCertificates(certsDir); err != nil {
			return fmt.Errorf("failed to generate certificates: %v", err)
		}
	}
	return nil
}

func (n *Node) GenerateCertificates(certsDir string) error {
	keyFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.key", n.IP, n.Port))
	csrFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.csr", n.IP, n.Port))
	certFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.crt", n.IP, n.Port))

	fmt.Printf("Generating key: %s\n", keyFile)
	err := exec.Command("openssl", "genpkey", "-algorithm", "RSA", "-out", keyFile).Run()
	if err != nil {
		return fmt.Errorf("error generating key: %v", err)
	}

	fmt.Printf("Generating CSR: %s\n", csrFile)
	err = exec.Command("openssl", "req", "-new", "-key", keyFile, "-out", csrFile, "-subj", fmt.Sprintf("/CN=%d", n.Port)).Run()
	if err != nil {
		return fmt.Errorf("error generating CSR: %v", err)
	}

	caDir := filepath.Join("certificates", "CA")
	fmt.Printf("Generating certificate: %s\n", certFile)
	cmd := exec.Command("openssl", "x509", "-req", "-in", csrFile, "-CA", filepath.Join(caDir, "ca.pem"), "-CAkey", filepath.Join(caDir, "ca.key"), "-CAcreateserial", "-out", certFile, "-days", "365")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error generating certificate: %v", err)
	}
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

func (n *Node) JoinNetwork(dhtInstance *DHT) {
	dhtInstance.JoinNetwork(n)
}

func (n *Node) LeaveNetwork(dhtInstance *DHT) error {
	return dhtInstance.LeaveNetwork(n)
}
