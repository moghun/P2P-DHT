package tests

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func TestStoreToStorage(t *testing.T) {
	//key := [32]byte{}
	keyStr := "key"
	hashKey := dht.EnsureKeyHashed(keyStr)
	value := []byte("value")

	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT:     dht,
	}

	// Store the value in the DHT
	err := realNode.DHT.StoreToStorage(hashKey, string(value), 3600)
	assert.NoError(t, err, "StoreToStorage should not return an error")

	// Check if the value exists in the storage
	retrievedValue, err := realNode.DHT.GetFromStorage(hashKey)
	assert.NoError(t, err, "GetFromStorage should not return an error")
	assert.Equal(t, string(value), retrievedValue, "The value retrieved should match the value stored")
}

func TestGetFromStorage(t *testing.T) {
	//key := [32]byte{}
	keyStr := "key"
	hashKey := dht.EnsureKeyHashed(keyStr)
	value := []byte("value")

	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT:     dht,
	}

	// Store the value in the DHT
	err := realNode.DHT.StoreToStorage(hashKey, string(value), 3600)
	assert.NoError(t, err, "StoreToStorage should not return an error")

	// Check if the value exists in the storage
	retrievedValue, err := realNode.DHT.GetFromStorage(hashKey)
	assert.NoError(t, err, "GetFromStorage should not return an error")
	assert.Equal(t, string(value), retrievedValue, "The value retrieved should match the value stored")
}

func TestSendStoreMessage(t *testing.T) {
	// Set up the receiver node and its network
	receiverPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	receiverDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", receiverPort)
	receiverStore := storage.NewStorage(86400, []byte("1234567890abcdef"))
	receiverNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    receiverPort,
		Storage: receiverStore,
		DHT:     receiverDht,
	}

	// Set up the sender node and its network
	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")
	senderDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", senderPort)
	senderStore := storage.NewStorage(86400, []byte("1234567890abcdef"))
	senderNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    senderPort,
		Storage: senderStore,
		DHT:     senderDht,
	}

	time.Sleep(2 * time.Second)
	config := &util.Config{
		EncryptionKey:    []byte("1234567890123456"),
		RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty:       4,
	}
	api.InitRateLimiter(config)

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
	hashKey := dht.EnsureKeyHashed(key)
	value := "testvalue"

	receiverKNode := dht.KNode{
		ID:   receiverNode.ID,
		IP:   receiverNode.IP,
		Port: receiverPort,
	}

	storeMessageResponse, err := senderDht.SendStoreMessage(hashKey, value, receiverKNode)
	if err != nil {
		log.Print("Send store message error")
	}

	//cast storeMessageResponse to DHTSuccessMessage
	log.Print("Store message response msg typeXXXXXX:", storeMessageResponse.GetType())
	assert.Equal(t, message.DHT_SUCCESS, storeMessageResponse.GetType(), "The response message should be a success message")

	// Check if the value exists in the storage
	retrievedValue, _, err := receiverNode.DHT.GET(hashKey)
	log.Print("Retrieved value:", retrievedValue)
	assert.NoError(t, err, "GET should not return an error")
	assert.Equal(t, value, retrievedValue, "The value retrieved should match the value stored")
}

func TestCreateStoreMessage(t *testing.T) {
	key := "testkey"
	hashKey := dht.EnsureKeyHashed(key)
	value := "testvalue"

	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT:     dht,
	}

	// Create a store message
	storeMsg, err := realNode.DHT.CreateStoreMessage(hashKey, value)

	assert.NoError(t, err)
	assert.NotNil(t, storeMsg)

	// Deserialize the store message
	deserializedMsg, err := message.DeserializeMessage(storeMsg)
	assert.NoError(t, err)
	assert.NotNil(t, deserializedMsg)
	assert.IsType(t, &message.DHTStoreMessage{}, deserializedMsg)

	// Convert deserialized key to a string for comparison
	storeMessage := deserializedMsg.(*message.DHTStoreMessage)
	deserializedKey := message.Byte32ToHexEncode(storeMessage.Key)

	// Compare both keys as strings
	assert.Equal(t, hashKey, deserializedKey)
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
		{"id1", "f3436f50b2f7f1613ad142dbce1d24801d9daaab"},
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
		distance, err := dht.XOR(test.id1, test.id2)
		if err != nil {
			t.Errorf("XOR(%q, %q) returned an error: %v", test.id1, test.id2, err)
		}
		if distance != test.distance {
			t.Errorf("XOR(%q, %q) = %d; want %d", test.id1, test.id2, distance, test.distance)
		}
	}
}

func TestFindNode(t *testing.T) {
	// Initialize a real storage and node for testing
	nodeId := "id0"
	hashedNodeId := dht.EnsureKeyHashed(nodeId)
	newDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	newDht.RoutingTable.NodeID = hashedNodeId

	var hashList []string
	idList := []string{"id1", "id2", "id3", "id4", "id5", "id6", "id7", "id8", "id9", "id10", "id11", "id12", "id13", "id14", "id15", "id16", "id17", "id18", "id19", "id20",
		"id21", "id22", "id23", "id24", "id25", "id26", "id27", "id28", "id29", "id30", "id31", "id32", "id33", "id34", "id35", "id36", "id37", "id38", "id39", "id40"}
	for _, id := range idList {
		hashList = append(hashList, dht.EnsureKeyHashed(id))
	}

	for i, hash := range hashList {
		newKNode := &dht.KNode{
			ID:   hash,
			IP:   "1",
			Port: 8080,
		}
		log.Print("Adding node to routing table: ", i)
		newDht.RoutingTable.AddNode(newKNode)
	}

	strKey := "id45"
	strKey = dht.EnsureKeyHashed(strKey)

	log.Print("Finding node closest to key:", strKey)
	closestNodes, err := newDht.FindNode(strKey)
	if err != nil {
		assert.NoError(t, err, "FindNode should not return an error")
	}

	assert.NotNil(t, closestNodes, "FindNode should return a list of closest nodes")
	//assert.NotEmpty(t, closestNodes, "FindNode should return a non-empty list of closest nodes")

	for _, node := range closestNodes {
		log.Print("ID:", node.ID)
	}
}

func TestFindValue_NoValue(t *testing.T) {
	// Initialize a real storage and node for testing
	nodeId := "id0"
	hashedNodeId := dht.EnsureKeyHashed(nodeId)
	newDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	newDht.RoutingTable.NodeID = hashedNodeId

	var hashList []string
	idList := []string{"id1", "id2", "id3", "id4", "id5", "id6", "id7", "id8", "id9", "id10", "id11", "id12", "id13", "id14", "id15", "id16", "id17", "id18", "id19", "id20",
		"id21", "id22", "id23", "id24", "id25", "id26", "id27", "id28", "id29", "id31", "id32", "id33", "id34", "id35", "id36", "id37", "id38", "id39", "id40"}
	for _, id := range idList {
		hashList = append(hashList, dht.EnsureKeyHashed(id))
	}

	for i, hash := range hashList {
		newKNode := &dht.KNode{
			ID:   hash,
			IP:   "1",
			Port: 8080,
		}
		log.Print("Adding node to routing table: ", i)
		newDht.RoutingTable.AddNode(newKNode)
	}

	strKey := "id30"
	strKey = dht.EnsureKeyHashed(strKey)

	log.Print("Finding node closest to key:", strKey)
	val, closestNodes, err := newDht.FindValue(strKey)
	if err != nil {
		assert.NoError(t, err, "FindNode should not return an error")
	}

	if val != "" {
		log.Print("Value:", val)
	} else {
		log.Print("Value not found, nodes length: ", len(closestNodes))
		for _, node := range closestNodes {
			log.Print("ID:", node.ID)
		}
	}

	assert.Equal(t, "", val, "FindValue should return an empty value")
	assert.NotNil(t, closestNodes, "FindNode should return a list of closest nodes")
	assert.NotEmpty(t, closestNodes, "FindNode should return a non-empty list of closest nodes")
}

func TestFindValue_WithValue(t *testing.T) {
	// Initialize a real storage and node for testing
	nodeId := "id0"
	hashedNodeId := dht.EnsureKeyHashed(nodeId)
	newDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	newDht.RoutingTable.NodeID = hashedNodeId

	var hashList []string
	idList := []string{"id1", "id2", "id3", "id4", "id5", "id6", "id7", "id8", "id9", "id10", "id11", "id12", "id13", "id14", "id15", "id16", "id17", "id18", "id19", "id20",
		"id21", "id22", "id23", "id24", "id25", "id26", "id27", "id28", "id29", "id31", "id32", "id33", "id34", "id35", "id36", "id37", "id38", "id39", "id40"}
	for _, id := range idList {
		hashList = append(hashList, dht.EnsureKeyHashed(id))
	}

	for i, hash := range hashList {
		newKNode := &dht.KNode{
			ID:   hash,
			IP:   "1",
			Port: 8080,
		}
		log.Print("Adding node to routing table: ", i)
		newDht.RoutingTable.AddNode(newKNode)
	}

	strKey := "id30"
	strKey = dht.EnsureKeyHashed(strKey)

	newDht.StoreToStorage(strKey, "value", 3600)

	val, closestNodes, err := newDht.FindValue(strKey)
	if err != nil {
		assert.NoError(t, err, "FindNode should not return an error")
	}

	if val != "" {
		log.Print("Value:", val)
	} else {
		log.Print("Value not found, nodes length: ", len(closestNodes))
		for _, node := range closestNodes {
			log.Print("ID:", node.ID)
		}
	}

	assert.NotEqual(t, "", val, "FindValue should return a non-empty value")
	assert.Equal(t, "value", val, "FindValue should return the correct value")
}

func TestGet_WithValue(t *testing.T) {
	nodeId := "id0"
	hashedNodeId := dht.EnsureKeyHashed(nodeId)

	newDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	newDht.RoutingTable.NodeID = hashedNodeId

	var hashList []string
	idList := []string{"id1", "id2", "id3", "id4", "id5", "id6", "id7", "id8", "id9", "id10", "id11", "id12", "id13", "id14", "id15", "id16", "id17", "id18", "id19", "id20",
		"id21", "id22", "id23", "id24", "id25", "id26", "id27", "id28", "id29", "id31", "id32", "id33", "id34", "id35", "id36", "id37", "id38", "id39", "id40"}
	for _, id := range idList {
		hashList = append(hashList, dht.EnsureKeyHashed(id))
	}

	for i, hash := range hashList {
		newKNode := &dht.KNode{
			ID:   hash,
			IP:   "1",
			Port: 8080,
		}
		log.Print("Adding node to routing table: ", i)
		newDht.RoutingTable.AddNode(newKNode)
	}

	strKey := "id30"
	strKey = dht.EnsureKeyHashed(strKey)

	newDht.StoreToStorage(strKey, "value", 3600)

	val, closestNodes, err := newDht.GET(strKey)
	if err != nil {
		assert.NoError(t, err, "Get should not return an error")
	}

	if val != "" {
		log.Print("Value:", val)
	} else {
		log.Print("Value not found, nodes length: ", len(closestNodes))
		for _, node := range closestNodes {
			log.Print("ID:", node.ID)
		}
	}

	assert.NotEqual(t, "", val, "Get should return a non-empty value")
	assert.Equal(t, "value", val, "Get should return the correct value")
}

func TestGet_NoValue(t *testing.T) {
	// Initialize a real storage and node for testing
	nodeId := "id0"
	hashedNodeId := dht.EnsureKeyHashed(nodeId)
	newDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	newDht.RoutingTable.NodeID = hashedNodeId

	var hashList []string
	idList := []string{"id1", "id2", "id3", "id4", "id5", "id6", "id7", "id8", "id9", "id10", "id11", "id12", "id13", "id14", "id15", "id16", "id17", "id18", "id19", "id20",
		"id21", "id22", "id23", "id24", "id25", "id26", "id27", "id28", "id29", "id31", "id32", "id33", "id34", "id35", "id36", "id37", "id38", "id39", "id40"}
	for _, id := range idList {
		hashList = append(hashList, dht.EnsureKeyHashed(id))
	}

	for i, hash := range hashList {
		newKNode := &dht.KNode{
			ID:   hash,
			IP:   "1",
			Port: 8080,
		}
		log.Print("Adding node to routing table: ", i)
		newDht.RoutingTable.AddNode(newKNode)
	}

	strKey := "id30"
	strKey = dht.EnsureKeyHashed(strKey)

	log.Print("Finding node closest to key:", strKey)
	val, closestNodes, err := newDht.GET(strKey)
	if err != nil {
		assert.NoError(t, err, "Get should not return an error")
	}

	if val != "" {
		log.Print("Value:", val)
	} else {
		log.Print("Value not found, nodes length: ", len(closestNodes))
		for _, node := range closestNodes {
			log.Print("ID:", node.ID)
		}
	}

	assert.Equal(t, "", val, "Geet should return an empty value")
	assert.NotNil(t, closestNodes, "Get should return a list of closest nodes")
	assert.NotEmpty(t, closestNodes, "Get should return a non-empty list of closest nodes")
}

func TestPut(t *testing.T) {
	// Set up the receiver node and its network
	receiverPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	receiverDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", receiverPort)
	receiverStore := storage.NewStorage(86400, []byte("1234567890abcdef"))
	nodeId := "id0"
	hashedNodeId := dht.EnsureKeyHashed(nodeId)
	receiverDht.RoutingTable.NodeID = hashedNodeId
	receiverNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    receiverPort,
		Storage: receiverStore,
		DHT:     receiverDht,
	}

	// Set up the sender node and its network
	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	senderDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", senderPort)
	senderStore := storage.NewStorage(86400, []byte("1234567890abcdef"))
	nodeId2 := "id0"
	hashedNodeId2 := dht.EnsureKeyHashed(nodeId2)
	senderDht.RoutingTable.NodeID = hashedNodeId2

	senderNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    senderPort,
		Storage: senderStore,
		DHT:     senderDht,
	}

	config := &util.Config{
		EncryptionKey:    []byte("1234567890123456"),
		RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty:       4,
	}
	api.InitRateLimiter(config)

	go func() {
		err := api.StartServer(receiverNode.IP+":"+fmt.Sprint(receiverPort), receiverNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	go func() {
		err := api.StartServer(senderNode.IP+":"+fmt.Sprint(senderPort), senderNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	time.Sleep(2 * time.Second)

	var hashList []string
	idList := []string{"id1", "id2", "id3", "id4", "id5", "id6", "id7", "id8", "id9", "id10", "id11", "id12", "id13", "id14", "id15", "id16", "id17", "id18", "id19", "id20",
		"id21", "id22", "id23", "id24", "id25", "id26", "id27", "id28", "id29", "id31", "id32", "id33", "id34", "id35", "id36", "id37", "id38", "id39", "id40"}

	for _, id := range idList {
		hashList = append(hashList, dht.EnsureKeyHashed(id))
	}

	for i, hash := range hashList {
		newKNode := &dht.KNode{
			ID:   hash,
			IP:   "127.0.0.1",
			Port: receiverPort,
		}
		log.Print("Adding node to routing table: ", i)
		receiverDht.RoutingTable.AddNode(newKNode)
		senderDht.RoutingTable.AddNode(newKNode)
	}

	strKey := "id30"
	strKey = dht.EnsureKeyHashed(strKey)

	err = senderDht.PUT(strKey, "value", 3600)

	log.Print("Waiting for 2 seconds for PUT to complete")
	time.Sleep(2 * time.Second)

	val, closestNodes, err := receiverDht.GET(strKey)
	if err != nil {
		assert.NoError(t, err, "GET should not return an error")
	}

	if val != "" {
		log.Print("Value:", val)
	} else {
		log.Print("Value not found, nodes length: ", len(closestNodes))
		for _, node := range closestNodes {
			log.Print("ID:", node.ID)
		}
	}

	assert.NotEqual(t, "", val, "GET should return a non-empty value")
	assert.Equal(t, "value", val, "GET should return the correct value")
}

func TestIterativeFindValue(t *testing.T) {
	// Set up the receiver node and its network
	receiverPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	receiverDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", receiverPort)
	receiverStore := storage.NewStorage(86400, []byte("1234567890abcdef"))
	nodeId := "id0"
	hashedNodeId := dht.EnsureKeyHashed(nodeId)
	receiverDht.RoutingTable.NodeID = hashedNodeId
	receiverNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    receiverPort,
		Storage: receiverStore,
		DHT:     receiverDht,
	}
	receiverNode.ID = hashedNodeId

	// Set up the sender node and its network
	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	senderDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", senderPort)
	senderStore := storage.NewStorage(86400, []byte("1234567890abcdef"))
	nodeId2 := "id20"
	hashedNodeId2 := dht.EnsureKeyHashed(nodeId2)
	senderDht.RoutingTable.NodeID = hashedNodeId2

	senderNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    senderPort,
		Storage: senderStore,
		DHT:     senderDht,
	}
	senderNode.ID = hashedNodeId2

	strKey := "targetKey"
	strKeyHashed := dht.EnsureKeyHashed(strKey)

	util.Log().Info("Starting intermediate nodes")
	iNode1 := CreateAndStartNodeWithGivenDistance(strKeyHashed, 120)
	senderDht.RoutingTable.AddNode(GetKNodeOfNode(iNode1))
	iNode2 := CreateAndStartNodeWithGivenDistance(strKeyHashed, 100)
	iNode1.DHT.RoutingTable.AddNode(GetKNodeOfNode(iNode2))
	iNode3 := CreateAndStartNodeWithGivenDistance(strKeyHashed, 80)
	iNode2.DHT.RoutingTable.AddNode(GetKNodeOfNode(iNode3))
	iNode3.DHT.RoutingTable.AddNode(GetKNodeOfNode(receiverNode))

	config := &util.Config{
		EncryptionKey:    []byte("1234567890123456"),
		RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty:       4,
	}
	api.InitRateLimiter(config)

	go func() {
		err := api.StartServer(receiverNode.IP+":"+fmt.Sprint(receiverPort), receiverNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	go func() {
		err := api.StartServer(senderNode.IP+":"+fmt.Sprint(senderPort), senderNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	time.Sleep(2 * time.Second)

	targetValue := "targetValue"

	err = receiverDht.StoreToStorage(strKeyHashed, targetValue, 3600) //Store target value to receiver
	assert.NoError(t, err, "Store should not return an error")

	log.Print("Waiting for 2 seconds for Store and node assignments to complete")
	time.Sleep(2 * time.Second)

	val, closestNodes, err := receiverDht.GET(strKeyHashed) //Check if targetvalue is successfully stored
	if err != nil {
		assert.NoError(t, err, "GET should not return an error")
	}

	if val != "" {
		log.Print("Value found in storage, control OK:", val)
	} else {
		log.Print("Value has not been stored, nodes length: ", len(closestNodes))
		for _, node := range closestNodes {
			log.Print("ID:", node.ID)
		}
	}

	assert.NotEqual(t, "", val, "GET should return a non-empty value")
	assert.Equal(t, targetValue, val, "GET should return the correct value")

	log.Print("Starting iterative find value")
	finalValue, finalNodes, err := senderDht.IterativeFindValue(strKeyHashed)

	assert.NoError(t, err, "IterativeFindValue should not return an error")
	util.Log().Info("Final nodes should be empty, nodes length:", len(finalNodes))
	assert.NotEqual(t, "", finalValue, "IterativeFindValue should return a non-empty value")
	assert.Equal(t, targetValue, finalValue, "IterativeFindValue should return the correct value")
}

func TestIterativeFindNode(t *testing.T) {
	// Set up the receiver node and its network
	receiverPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	receiverDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", receiverPort)
	receiverStore := storage.NewStorage(86400, []byte("1234567890abcdef"))
	nodeId := "id0"
	hashedNodeId := dht.EnsureKeyHashed(nodeId)
	receiverDht.RoutingTable.NodeID = hashedNodeId
	receiverNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    receiverPort,
		Storage: receiverStore,
		DHT:     receiverDht,
	}
	receiverNode.ID = hashedNodeId

	// Set up the sender node and its network
	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	senderDht := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", senderPort)
	senderStore := storage.NewStorage(86400, []byte("1234567890abcdef"))
	nodeId2 := "id20"
	hashedNodeId2 := dht.EnsureKeyHashed(nodeId2)
	senderDht.RoutingTable.NodeID = hashedNodeId2

	senderNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    senderPort,
		Storage: senderStore,
		DHT:     senderDht,
	}
	senderNode.ID = hashedNodeId2

	strKey := "targetKey"
	strKeyHashed := dht.EnsureKeyHashed(strKey)

	util.Log().Info("Starting intermediate nodes")
	iNode1 := CreateAndStartNodeWithGivenDistance(strKeyHashed, 120)
	senderDht.RoutingTable.AddNode(GetKNodeOfNode(iNode1))
	iNode2 := CreateAndStartNodeWithGivenDistance(strKeyHashed, 100)
	iNode1.DHT.RoutingTable.AddNode(GetKNodeOfNode(iNode2))
	iNode3 := CreateAndStartNodeWithGivenDistance(strKeyHashed, 80)
	iNode2.DHT.RoutingTable.AddNode(GetKNodeOfNode(iNode3))
	iNode3.DHT.RoutingTable.AddNode(GetKNodeOfNode(receiverNode))

	config := &util.Config{
		EncryptionKey:    []byte("1234567890123456"),
		RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty:       4,
	}
	api.InitRateLimiter(config)

	go func() {
		err := api.StartServer(receiverNode.IP+":"+fmt.Sprint(receiverPort), receiverNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	go func() {
		err := api.StartServer(senderNode.IP+":"+fmt.Sprint(senderPort), senderNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	time.Sleep(2 * time.Second)

	targetValue := "targetValue"

	var receiveKeyList []string
	//geneerate 5 hashedKeys that has distance 87 to targetKeyHashed
	for i := 0; i < 4; i++ {
		newKey := generateHammingDistanceString(receiverNode.ID, 87)
		receiveKeyHashed := dht.EnsureKeyHashed(newKey)
		receiveKeyList = append(receiveKeyList, receiveKeyHashed)
	}

	for _, key := range receiveKeyList {
		newKNode := &dht.KNode{
			ID:   key,
			IP:   "127.0.0.1",
			Port: receiverPort,
		}
		receiverDht.RoutingTable.AddNode(newKNode)

		//err = receiverDht.StoreToStorage(key, targetValue, 3600) //Store target value to receiver
		//assert.NoError(t, err, "Store should not return an error")
	}

	err = receiverDht.StoreToStorage(strKeyHashed, targetValue, 3600) //Store target value to receiver
	assert.NoError(t, err, "Store should not return an error")

	log.Print("Waiting for 2 seconds for Store and node assignments to complete")
	time.Sleep(2 * time.Second)

	val, closestNodes, err := receiverDht.GET(strKeyHashed) //Check if targetvalue is successfully stored
	if err != nil {
		assert.NoError(t, err, "GET should not return an error")
	}

	if val != "" {
		log.Print("Value found in storage, control OK:", val)
	} else {
		log.Print("Value has not been stored, nodes length: ", len(closestNodes))
		for _, node := range closestNodes {
			log.Print("ID:", node.ID)
		}
	}

	assert.NotEqual(t, "", val, "GET should return a non-empty value")
	assert.Equal(t, targetValue, val, "GET should return the correct value")

	log.Print("Starting iterative find value")
	finalNodes, err := senderDht.IterativeFindNode(strKeyHashed)

	assert.NoError(t, err, "IterativeFindValue should not return an error")
	util.Log().Info("Final nodes should not be empty, nodes length:", len(finalNodes))
	assert.NotEqual(t, 0, len(finalNodes), "IterativeFindValue should return a non-empty value")
}

func CreateAndStartNode(id string) *node.Node {
	intermediatePort, err := tests.GetFreePort()
	if err != nil {
		log.Print("Failed to get a free port")
		return nil
	}

	intermediateDht1 := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", intermediatePort)
	intermediateStore1 := storage.NewStorage(86400, []byte("1234567890abcdef"))
	intermediateNodeId1 := id
	intermediateHashedNodeId1 := dht.EnsureKeyHashed(intermediateNodeId1)
	intermediateDht1.RoutingTable.NodeID = intermediateHashedNodeId1

	intermediateNode1 := &node.Node{
		IP:      "127.0.0.1",
		Port:    intermediatePort,
		Storage: intermediateStore1,
		DHT:     intermediateDht1,
	}
	intermediateNode1.ID = intermediateHashedNodeId1

	go func() {
		_ = api.StartServer(intermediateNode1.IP+":"+fmt.Sprint(intermediateNode1.Port), intermediateNode1)
	}()

	time.Sleep(2 * time.Second)
	util.Log().Info("Intermediate node started, id: ", intermediateDht1.RoutingTable.NodeID)
	return intermediateNode1
}

func GetKNodeOfNode(node *node.Node) *dht.KNode {
	return &dht.KNode{
		ID:   node.ID,
		IP:   node.IP,
		Port: node.Port,
	}
}

// flipBitInHex flips a bit at the given bit position in a hex-encoded string
func flipBitInHex(hexString string, bitIndex int) string {
	// Convert the hex string back to a byte array
	bytes, _ := hex.DecodeString(hexString)

	// Calculate byte index and bit position within the byte
	byteIndex := bitIndex / 8
	bitPosition := bitIndex % 8

	// Flip the bit at the given position
	bytes[byteIndex] ^= (1 << (7 - bitPosition))

	// Convert back to hex string
	return hex.EncodeToString(bytes)
}

// generateHammingDistanceString generates a string with a specified Hamming distance from the input
func generateHammingDistanceString(input string, distance int) string {
	rand.Seed(time.Now().UnixNano())

	// The input is a hex string, so it has 160 bits (since it's 40 hex characters)
	totalBits := len(input) * 4 // 4 bits per hex character

	if distance > totalBits {
		panic("Hamming distance is greater than the total number of bits in the input string.")
	}

	// Track which bits have already been flipped to avoid flipping the same bit twice
	flippedBits := make(map[int]bool)

	// Mutable copy of the input string
	output := input

	// Randomly flip 'distance' bits
	for i := 0; i < distance; i++ {
		for {
			// Randomly choose a bit to flip
			bitIndex := rand.Intn(totalBits)

			// If the bit is already flipped, choose another
			if flippedBits[bitIndex] {
				continue
			}

			// Flip the bit in the output string
			output = flipBitInHex(output, bitIndex)

			// Mark the bit as flipped
			flippedBits[bitIndex] = true

			break
		}
	}

	return output
}

func CreateAndStartNodeWithGivenDistance(originID string, distance int) *node.Node {
	intermediatePort, err := tests.GetFreePort()
	if err != nil {
		log.Print("Failed to get a free port")
		return nil
	}

	intermediateDht1 := dht.NewDHT(86400, []byte("1234567890abcdef"), "2", "127.0.0.1", intermediatePort)
	intermediateStore1 := storage.NewStorage(86400, []byte("1234567890abcdef"))
	intermediateHashedNodeId1 := generateHammingDistanceString(originID, distance)
	intermediateDht1.RoutingTable.NodeID = intermediateHashedNodeId1

	intermediateNode1 := &node.Node{
		IP:      "127.0.0.1",
		Port:    intermediatePort,
		Storage: intermediateStore1,
		DHT:     intermediateDht1,
	}
	intermediateNode1.ID = intermediateHashedNodeId1

	go func() {
		_ = api.StartServer(intermediateNode1.IP+":"+fmt.Sprint(intermediateNode1.Port), intermediateNode1)
	}()

	time.Sleep(2 * time.Second)
	util.Log().Info("Intermediate node started, id: ", intermediateDht1.RoutingTable.NodeID)
	return intermediateNode1
}
