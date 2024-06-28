package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
)

func TestNewNode(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode("127.0.0.1", 8000, true, key)

	assert.NotNil(t, node)
	assert.Equal(t, "127.0.0.1", node.IP)
	assert.Equal(t, 8000, node.Port)
	assert.True(t, node.Ping)
	assert.NotNil(t, node.Storage)
	assert.Len(t, node.KBuckets, 160)
}

func TestAddPeer(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	peerID := dht.GenerateNodeID("192.168.1.1", 9000)
	node.AddPeer(peerID, "192.168.1.1", 9000)

	peers := node.GetAllPeers()
	assert.Equal(t, 1, len(peers))
	assert.Equal(t, "192.168.1.1", peers[0].IP)
	assert.Equal(t, 9000, peers[0].Port)
}

func TestRemovePeer(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	peerID := dht.GenerateNodeID("192.168.1.1", 9000)
	node.AddPeer(peerID, "192.168.1.1", 9000)

	node.RemovePeer("192.168.1.1", 9000)
	peers := node.GetAllPeers()
	assert.Equal(t, 0, len(peers))
}

func TestPutGetStorage(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode("127.0.0.1", 8000, true, key)

	err := node.Put("key1", "value1", int(time.Hour.Seconds()))
	assert.Nil(t, err, "Put operation should not return an error")

	value, err := node.Get("key1")
	assert.Nil(t, err, "Get operation should not return an error")
	assert.Equal(t, "value1", value, "Retrieved value should match the stored value")

	t.Logf("Retrieved value for key1: %s", value)
}

func TestGetClosestNodesToCurrNode(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	peerID1 := dht.GenerateNodeID("192.168.1.1", 9000)
	peerID2 := dht.GenerateNodeID("192.168.1.2", 9001)
	node.AddPeer(peerID1, "192.168.1.1", 9000)
	node.AddPeer(peerID2, "192.168.1.2", 9001)

	targetID := dht.GenerateNodeID("192.168.1.3", 9002)
	closestNodes := node.GetClosestNodesToCurrNode(targetID, 2)

	assert.Equal(t, 2, len(closestNodes))
}

func TestSetupTLS(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode("127.0.0.1", 8000, true, key)

	tlsDir := fmt.Sprintf("%s_%d", node.IP, node.Port)
	certsDir := filepath.Join("certificates", tlsDir)

	_, err := os.Stat(certsDir)
	assert.False(t, os.IsNotExist(err))
}
func TestGenerateCertificates(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode("127.0.0.1", 8000, true, key)

	tlsDir := fmt.Sprintf("%s_%d", node.IP, node.Port)
	certsDir := filepath.Join("certificates", tlsDir)

	err := os.MkdirAll(certsDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create certsDir: %v", err)
	}

	node.GenerateCertificates(certsDir)

	keyFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.key", node.IP, node.Port))
	certFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.csr", node.IP, node.Port))

	_, err = os.Stat(keyFile)
	if os.IsNotExist(err) {
		t.Fatalf("Key file does not exist: %v", err)
	}

	_, err = os.Stat(certFile)
	if os.IsNotExist(err) {
		t.Fatalf("Cert file does not exist: %v", err)
	}

	t.Logf("Key file: %s", keyFile)
	t.Logf("Cert file: %s", certFile)

	assert.False(t, os.IsNotExist(err), "Cert file does not exist")
}

func TestNodeJoinNetwork(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	node.JoinNetwork(dhtInstance)

	peers := dhtInstance.GetNumNodes()
	assert.Equal(t, 1, len(peers))
	assert.Equal(t, node.ID, peers[0].ID)
}

func TestNodeLeaveNetwork(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	node.JoinNetwork(dhtInstance)

	err := node.LeaveNetwork(dhtInstance)
	assert.Nil(t, err)

	peers := dhtInstance.GetNumNodes()
	assert.Equal(t, 0, len(peers))
}