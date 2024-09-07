package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"

)

// TestNodeGenerateNodeID tests the GenerateNodeID function.
func TestNodeGenerateNodeID(t *testing.T) {
	nodeID := node.GenerateNodeID("127.0.0.1", 8080)
	assert.NotEmpty(t, nodeID)
}

// TestNewNodeInitialization tests the initialization of a new Node.
func TestNewNodeInitialization(t *testing.T) {
    port, err := tests.GetFreePort()
    assert.NoError(t, err, "Failed to get a free port")

    config := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port),
        EncryptionKey: []byte("1234567890123456"),
    }
    nodeInstance := node.NewNode(config, 24 * time.Hour)

    assert.NotNil(t, nodeInstance)
    assert.Equal(t, "127.0.0.1", nodeInstance.IP)
    assert.Equal(t, port, nodeInstance.Port)
    assert.NotNil(t, nodeInstance.Storage)
    assert.IsType(t, &storage.Storage{}, nodeInstance.Storage)
}


// TestNodePutAndGet tests the Put and Get methods of the Node.
func TestNodePutAndGet(t *testing.T) {
	config := &util.Config{
		P2PAddress:    "127.0.0.1:8080",
		EncryptionKey: []byte("1234567890123456"),
	}
	nodeInstance := node.NewNode(config, 24 * time.Hour)

	err := nodeInstance.Put("key1", "value1", 60)
	assert.NoError(t, err)

	value, err := nodeInstance.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
}

func TestNodeStorageTTL(t *testing.T) {
	config := &util.Config{
		P2PAddress:    "127.0.0.1:8080",
		EncryptionKey: []byte("1234567890123456"),
	}
	nodeInstance := node.NewNode(config, 24 * time.Hour)

	err := nodeInstance.Put("key1", "value1", 1)
	assert.NoError(t, err)

	// Immediately check that the value is there
	value, err := nodeInstance.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	time.Sleep(2 * time.Second) // Wait for the TTL to expire

	// After sleep, the value should be gone
	value, err = nodeInstance.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "", value) // Value should be empty after TTL expiration
}