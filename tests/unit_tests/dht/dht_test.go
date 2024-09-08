package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
)

func TestStoreToStorage(t *testing.T) {
	key := [32]byte{}
	value := []byte("value")

	// Initialize a real storage and node for testing
	store := storage.NewStorage(24*time.Hour, []byte("1234567890abcdef"))
	dht := dht.NewDHT(24*time.Hour, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT:     dht,
	}

	// Store the value in the DHT
	err := realNode.DHT.StoreToStorage(string(key[:]), string(value), 3600)
	assert.NoError(t, err, "StoreToStorage should not return an error")

	// Check if the value exists in the storage
	retrievedValue, err := realNode.DHT.GetFromStorage(string(key[:]))
	assert.NoError(t, err, "GetFromStorage should not return an error")
	assert.Equal(t, string(value), retrievedValue, "The value retrieved should match the value stored")
}
func TestCreateStoreMessage(t *testing.T) {
	key := "testkey"
	value := "testvalue"

	// Initialize a real storage and node for testing
	store := storage.NewStorage(24*time.Hour, []byte("1234567890abcdef"))
	dht := dht.NewDHT(24*time.Hour, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT:     dht,
	}

	// Create a store message
	storeMsg, err := realNode.DHT.CreateStoreMessage(key, value)

	assert.NoError(t, err)
	assert.NotNil(t, storeMsg)

	// Deserialize the store message
	deserializedMsg, err := message.DeserializeMessage(storeMsg)
	assert.NoError(t, err)
	assert.NotNil(t, deserializedMsg)
	assert.IsType(t, &message.DHTStoreMessage{}, deserializedMsg)

	// Convert deserialized key to a string for comparison
	storeMessage := deserializedMsg.(*message.DHTStoreMessage)
	deserializedKey := string(storeMessage.Key[:]) // Convert byte32 array to string
	trimmedKey := strings.TrimRight(deserializedKey, "\x00")

	// Compare both keys as strings
	assert.Equal(t, key, trimmedKey)
}
