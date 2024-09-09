package tests

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
)

func TestHandlePut(t *testing.T) {
	key := [32]byte{}
	value := []byte("value")

	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"),"1","127.0.0.1",8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT: dht,
	}

	putMsg := message.NewDHTPutMessage(10000, 2, key, value)
	serializedMsg, err := putMsg.Serialize()
	log.Print(serializedMsg)
	assert.NoError(t, err)

	response := api.HandlePut(putMsg, realNode)
	log.Print(response)
	assert.NotNil(t, response)

	// Verify that the value was actually stored in the node's storage
	storedValue, err := realNode.DHT.GET(string(key[:]))
	log.Print(storedValue)
	assert.NoError(t, err)
	assert.Equal(t, string(value), storedValue)
}

func TestHandleGet(t *testing.T) {
	key := [32]byte{}
	value := "value"

	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"),"1","127.0.0.1",8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT: dht,
	}

	// Pre-store the value in the node's storage
	err := realNode.DHT.PUT(string(key[:]), value, 10000)
	assert.NoError(t, err)

	getMsg := message.NewDHTGetMessage(key)
	serializedMsg, err := getMsg.Serialize()
	log.Print(serializedMsg)
	assert.NoError(t, err)

	response := api.HandleGet(getMsg, realNode)
	assert.NotNil(t, response)

	// Check that the response contains the correct value
	deserializedResponse, err := message.DeserializeMessage(response)
	assert.NoError(t, err)
	successMsg, ok := deserializedResponse.(*message.DHTSuccessMessage)
	assert.True(t, ok)
	assert.Equal(t, value, string(successMsg.Value))
}

func TestHandlePing(t *testing.T) {
	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"),"1","127.0.0.1",8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT: dht,
	}

	pingMsg := message.NewDHTPingMessage()
	_, err := pingMsg.Serialize()
	assert.NoError(t, err)

	response := api.HandlePing(pingMsg, realNode)
	assert.NotNil(t, response)

	// Check that the response is a PONG message
	deserializedResponse, err := message.DeserializeMessage(response)
	assert.NoError(t, err)
	assert.Equal(t, message.DHT_PONG, deserializedResponse.GetType())
}

func TestHandleFindNode(t *testing.T) {
	key := [32]byte{}

	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"),"1","127.0.0.1",8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT: dht,
	}

	findNodeMsg := message.NewDHTFindNodeMessage(key)
	_, err := findNodeMsg.Serialize()
	assert.NoError(t, err)

	response := api.HandleFindNode(findNodeMsg, realNode)
	assert.NotNil(t, response)

	// Check that the response contains the correct value
	deserializedResponse, err := message.DeserializeMessage(response)
	assert.NoError(t, err)
	_, ok := deserializedResponse.(*message.DHTSuccessMessage)
	assert.True(t, ok)
}

func TestHandleFindValue(t *testing.T) {
	key := [32]byte{123}

	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"),"1","127.0.0.1",8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT: dht,
	}

	dht.PUT(string(key[:]), "received", 100)
	
	findValueMsg := message.NewDHTFindValueMessage(key)
	_, err := findValueMsg.Serialize()
	assert.NoError(t, err)

	response := api.HandleFindValue(findValueMsg, realNode)
	assert.NotNil(t, response)

	// Check that the response contains the correct value
	deserializedResponse, err := message.DeserializeMessage(response)
	assert.NoError(t, err)
	_, ok := deserializedResponse.(*message.DHTSuccessMessage)
	assert.True(t, ok)
}

func TestHandleBootstrap(t *testing.T) {
	bootstrapData := fmt.Sprintf("127.0.0.1:%d", 8080)

	// Initialize a real storage and node for testing
	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	realNode := &node.BootstrapNode{
		Node: node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		},
		KnownPeers: map[string]string{},
	}

	bootstrapMsg := message.NewDHTBootstrapMessage(bootstrapData)
	_, err := bootstrapMsg.Serialize()
	assert.NoError(t, err)

	response := api.HandleBootstrap(bootstrapMsg, realNode)
	assert.NotNil(t, response)

	deserializedResponse, err := message.DeserializeMessage(response)
	assert.NoError(t, err)
	assert.Equal(t, message.DHT_BOOTSTRAP_REPLY, deserializedResponse.GetType())
}

func TestHandleBootstrapReply(t *testing.T) {
	bootstrapReplyData := "192.168.1.1:8081\n192.168.1.2:8082"

	store := storage.NewStorage(86400, []byte("1234567890abcdef"))
	dht := dht.NewDHT(86400, []byte("1234567890abcdef"),"1","127.0.0.1",8080)
	realNode := &node.Node{
		IP:      "127.0.0.1",
		Port:    8080,
		Storage: store,
		DHT: dht,
	}
	bootstrapReplyMsg := message.NewDHTBootstrapReplyMessage([]byte(bootstrapReplyData))
	_, err := bootstrapReplyMsg.Serialize()
	assert.NoError(t, err)

	response := api.HandleBootstrapReply(bootstrapReplyMsg, realNode)
	assert.Nil(t, response)
}