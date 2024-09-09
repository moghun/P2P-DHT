package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func TestBootstrapScenario(t *testing.T) {
	// Step 1: Set up the Bootstrap Node
	bootstrapPort, err := tests.GetFreePort()
	assert.NoError(t, err)

	bootstrapConfig := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", bootstrapPort),
		EncryptionKey: []byte("12345678901234567890123456789012"),
		RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty: 4,
		MaxBootstrapRetries: 2,
	}
	api.InitRateLimiter(bootstrapConfig)
	bootstrapNode := node.NewBootstrapNode(bootstrapConfig, 86400)
	util.Log().Printf("***  Node-bootstrap (%s) running at: %s  ***", bootstrapNode.ID, bootstrapConfig.P2PAddress)
	go api.StartServer(bootstrapConfig.P2PAddress, "", bootstrapNode)

	time.Sleep(1 * time.Second) // Ensure the bootstrap_node is up

	// Step 2: First peer node bootstraps with the bootstrap_node
	peer1Port, err := tests.GetFreePort()
	assert.NoError(t, err)

	peer1Config := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", peer1Port),
		EncryptionKey: []byte("12345678901234567890123456789012"),
		RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty: 4,
		MaxBootstrapRetries: 2,
		BootstrapNodes: []util.BootstrapNode{
			{IP: "127.0.0.1", Port: bootstrapPort},
		},
	}

	api.InitRateLimiter(peer1Config)
	peer1Node := node.NewNode(peer1Config, 86400)
	util.Log().Printf("***  Node-1 (%s) running at: %s  ***", peer1Node.ID, peer1Config.P2PAddress)
	go api.StartServer(peer1Config.P2PAddress, "", peer1Node)

	// Bootstrap peer1 with the bootstrap_node
	err = peer1Node.Bootstrap()
	assert.NoError(t, err)

	// Verify that the first node is added to bootstrap_node's peer list
	assert.Contains(t, bootstrapNode.GetKnownPeers(), fmt.Sprintf("127.0.0.1:%d", peer1Port))

	// Step 3: Second peer node bootstraps with the bootstrap_node
	peer2Port, err := tests.GetFreePort()
	assert.NoError(t, err)

	peer2Config := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", peer2Port),
		EncryptionKey: []byte("12345678901234567890123456789012"),
		RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty: 4,
		MaxBootstrapRetries: 2,
		BootstrapNodes: []util.BootstrapNode{
			{IP: "127.0.0.1", Port: bootstrapPort},
		},
	}

	api.InitRateLimiter(peer2Config)
	peer2Node := node.NewNode(peer2Config, 86400)
	util.Log().Printf("***  Node-2 (%s) running at: %s  ***", peer2Node.ID, peer2Config.P2PAddress)
	go api.StartServer(peer2Config.P2PAddress, "",peer2Node)

	// Bootstrap peer2 with the bootstrap_node
	err = peer2Node.Bootstrap()
	assert.NoError(t, err)

	// Verify that the second node is added to bootstrap_node's peer list
	assert.Contains(t, bootstrapNode.GetKnownPeers(), fmt.Sprintf("127.0.0.1:%d", peer2Port))

	// Verify that bootstrap_node returned the first node's address to the second node
	assert.Contains(t, peer2Node.GetAllPeers(), &dht.KNode{IP: "127.0.0.1", Port: peer1Port})

	// Step 4: Second node sends a PUT request to the first node
	key := "testkey"
	value := []byte("testvalue")
	hashKey := dht.EnsureKeyHashed(key)
	byteKey, err := message.HexStringToByte32(hashKey)
	assert.NoError(t, err)

	putMsg := message.NewDHTPutMessage(10000, 2, byteKey, value)
	serializedPutMsg, err := putMsg.Serialize()
	assert.NoError(t, err)

	// Send the PUT message from peer2 to peer1
	response, err := peer2Node.DHT.Network.SendMessage("127.0.0.1", peer1Port, serializedPutMsg)
	assert.NoError(t, err)

	// Verify that the PUT was successful
	successMsg, err := message.DeserializeMessage(response)
	assert.NoError(t, err)
	successPutMsg, ok := successMsg.(*message.DHTSuccessMessage)
	assert.True(t, ok)
	assert.Equal(t, byteKey, successPutMsg.Key)
	assert.Equal(t, value, successPutMsg.Value)

	// Step 5: Second node sends a GET request to the first node to retrieve the value
	getMsg := message.NewDHTGetMessage(byteKey)
	serializedGetMsg, err := getMsg.Serialize()
	assert.NoError(t, err)

	// Send the GET message from peer2 to peer1
	response, err = peer2Node.DHT.Network.SendMessage("127.0.0.1", peer1Port, serializedGetMsg)
	assert.NoError(t, err)

	// Verify that the GET was successful and the value matches
	successMsg, err = message.DeserializeMessage(response)
	assert.NoError(t, err)

	successGetMsg, ok := successMsg.(*message.DHTSuccessMessage)
	deserializedSuccessMsg, _ := dht.Deserialize(successGetMsg.Value)
	assert.Equal(t, value, []byte(deserializedSuccessMsg.Value))
}