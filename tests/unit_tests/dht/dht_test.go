package tests

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
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

func TestGetFromStorage(t *testing.T) {
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

func TestSendStoreMessage(t *testing.T) {
	// Set up the receiver node and its network
	receiverPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	receiverDht := dht.NewDHT(24*time.Hour, []byte("1234567890abcdef"), "2", "127.0.0.1", receiverPort)
	receiverStore := storage.NewStorage(24*time.Hour, []byte("1234567890abcdef"))
	receiverNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    receiverPort,
		Storage: receiverStore,
		DHT:     receiverDht,
	}

	// Set up the sender node and its network
	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")
	senderDht := dht.NewDHT(24*time.Hour, []byte("1234567890abcdef"), "2", "127.0.0.1", senderPort)
	senderStore := storage.NewStorage(24*time.Hour, []byte("1234567890abcdef"))
	senderNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    senderPort,
		Storage: senderStore,
		DHT:     senderDht,
	}

	time.Sleep(2 * time.Second)

	go func() {
		err := api.StartServer(receiverNode.IP+":"+fmt.Sprint(receiverPort), receiverNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	time.Sleep(2 * time.Second)

	go func() {
		err := api.StartServer(senderNode.IP+":"+fmt.Sprint(senderPort), senderNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	time.Sleep(2 * time.Second)

	key := "testkey"
	value := "testvalue"

	receiverKNode := dht.KNode{
		ID:   receiverNode.ID,
		IP:   receiverNode.IP,
		Port: receiverPort,
	}

	storeMessageResponse, err := senderDht.SendStoreMessage(key, value, receiverKNode)
	if err != nil {
		log.Print("Send store message error")
	}

	//cast storeMessageResponse to DHTSuccessMessage
	log.Print("Store message response msg typeXXXXXX:", storeMessageResponse.GetType())
	assert.Equal(t, message.DHT_SUCCESS, storeMessageResponse.GetType(), "The response message should be a success message")

	// Check if the value exists in the storage
	retrievedValue, _, err := receiverNode.DHT.GET(key)
	log.Print("Retrieved value:", retrievedValue)
	assert.NoError(t, err, "GET should not return an error")
	assert.Equal(t, value, retrievedValue, "The value retrieved should match the value stored")
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

func TestHashKey(t *testing.T) {
	// Define test cases
	tests := []struct {
		input    string
		expected string
	}{
		{"exampleKey", "c8f61512ae6831f3419eda280b68cee2113f63c5"}, // Expected hash for "exampleKey"
		{"anotherKey", "86f64b4f2882f39601f9c73fb3fab414fd0bbe71"}, // Expected hash for "anotherKey"
	}

	for _, test := range tests {
		result := dht.HashKey(test.input)
		log.Print("Result:", result)
		if result != test.expected {
			t.Errorf("HashKey(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestIsHashedKey(t *testing.T) {
	// Define test cases
	tests := []struct {
		input    string
		expected bool
	}{
		{"b1a1f0e8b915efb0911e7fc92f941bfc23dfd7d8", true}, // Valid 160-bit hex string
		{"de2054c8e5d885c63b545583fe065ab4aa8b7f16", true}, // Valid 160-bit hex string
		{"notAHexString", false},                           // Not a hex string
		{"b1a1f0e8b915efb0911e7fc92f941bfc23dfd7", false},  // Invalid length
	}

	for _, test := range tests {
		result := dht.IsHashedKey(test.input)
		if result != test.expected {
			t.Errorf("IsHashedKey(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestEnsureKeyHashed(t *testing.T) {
	// Define test cases
	tests := []struct {
		input    string
		expected string
	}{
		{"exampleKey", "c8f61512ae6831f3419eda280b68cee2113f63c5"},                               // Hash of "exampleKey"
		{"b1a1f0e8b915efb0911e7fc92f941bfc23dfd7d8", "b1a1f0e8b915efb0911e7fc92f941bfc23dfd7d8"}, // Already hashed
	}

	for _, test := range tests {
		result := dht.EnsureKeyHashed(test.input)
		if result != test.expected {
			t.Errorf("EnsureKeyHashed(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestRun(t *testing.T) {
	key1 := "simplekey1"
	key2 := "simplekey2"
	key3 := "simplekey3"
	key4 := "simplekey4"
	key5 := "simplekey5"
	key6 := "simplekey6"

	// Hash keys
	hashedKey1 := dht.EnsureKeyHashed(key1)
	hashedKey2 := dht.EnsureKeyHashed(key2)
	hashedKey3 := dht.EnsureKeyHashed(key3)
	hashedKey4 := dht.EnsureKeyHashed(key4)
	hashedKey5 := dht.EnsureKeyHashed(key5)
	hashedKey6 := dht.EnsureKeyHashed(key6)

	log.Print("Hashed key1:", hashedKey1)
	log.Print("Hashed key2:", hashedKey2)
	log.Print("Hashed key3:", hashedKey3)
	log.Print("Hashed key4:", hashedKey4)
	log.Print("Hashed key5:", hashedKey5)
	log.Print("Hashed key6:", hashedKey6)

	distance, err := dht.XOR(hashedKey1, hashedKey2)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("XOR distance between key1 and key2:", distance)

	distance, err = dht.XOR(hashedKey1, hashedKey3)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("XOR distance between key1 and key3:", distance)
	//assert.Equal(t, "b1a1f0e8b915efb0911e7fc92f941bfc23dfd7d8", hashedKey1)
}

// TestXORDistance tests the XORDistance function.
func TestXORDistance(t *testing.T) {
	tests := []struct {
		id1      string
		id2      string
		distance int
	}{
		// Example test cases with precomputed distances
		{"0000000000000000000000000000000000000000", "ffffffffffffffffffffffffffffffffffffffff", 160}, // Max distance (all bits differ)
		{"fe5b9e737884dd552689492b3af815401bb9e03d", "db4d37c234ed01182591d1a1794ae2f06444cb1f", 77},
		{"fe5b9e737884dd552689492b3af815401bb9e03d", "10090e73a9868c30fa72b9fab4b1bc6df95320c3", 76},
	}

	for _, test := range tests {
		result, err := dht.XOR(test.id1, test.id2)
		if err != nil {
			t.Errorf("XORDistance(%q, %q) returned an error: %v", test.id1, test.id2, err)
			continue
		}
		if result.Int64() != int64(test.distance) {
			t.Errorf("XORDistance(%q, %q) = %d; want %d", test.id1, test.id2, result, test.distance)
		}
	}
}

// Test for the XOR distance function
func TestXORDistance1(t *testing.T) {
	// Test cases with precomputed distances
	tests := []struct {
		id1      string
		id2      string
		distance int
	}{
		{"0000000000000000000000000000000000000000", "ffffffffffffffffffffffffffffffffffffffff", 160},
		{"fe5b9e737884dd552689492b3af815401bb9e03d", "db4d37c234ed01182591d1a1794ae2f06444cb1f", 77},
		{"fe5b9e737884dd552689492b3af815401bb9e03d", "10090e73a9868c30fa72b9fab4b1bc6df95320c3", 76},
	}

	for _, tt := range tests {
		t.Run(tt.id1+"_"+tt.id2, func(t *testing.T) {
			// Calculate XOR distance
			xorResult, err := dht.XOR(tt.id1, tt.id2)
			if err != nil {
				t.Fatalf("XOR failed with error: %v", err)
			}

			// Calculate the distance by finding the most significant bit set
			calculatedDistance := xorResult.BitLen()

			if calculatedDistance != tt.distance {
				t.Errorf("Expected distance %d, but got %d", tt.distance, calculatedDistance)
			}
		})
	}
}

// // TestCountBits tests the countBits function.
// func TestCountBits(t *testing.T) {
// 	tests := []struct {
// 		input    byte
// 		expected int
// 	}{
// 		{0x00, 0}, // No bits set
// 		{0x01, 1}, // Single bit set
// 		{0x0F, 4}, // All 4 lower bits set
// 		{0xFF, 8}, // All bits set in a byte
// 		{0xAA, 4}, // 10101010 (alternating bits)
// 	}

// 	for _, test := range tests {
// 		result := dht.CountBits(test.input)
// 		if result != test.expected {
// 			t.Errorf("countBits(%02x) = %d; want %d", test.input, result, test.expected)
// 		}
// 	}
// }
