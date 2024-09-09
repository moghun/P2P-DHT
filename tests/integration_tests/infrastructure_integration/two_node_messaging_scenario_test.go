package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

// TestTwoNodePingPong refactored to use SendMessage
func TestTwoNodePingPong(t *testing.T) {
    // Set up Node 1
    port1, err := tests.GetFreePort()
    assert.NoError(t, err)

    config1 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port1),
        EncryptionKey: []byte("1234567890123456"),
        RateLimiterRate:  10,
		RateLimiterBurst: 20,
    }
    api.InitRateLimiter(config1)
    node1 := node.NewNode(config1, 24*time.Hour)
    go api.StartServer(config1.P2PAddress, node1)

    // Set up Node 2
    port2, err := tests.GetFreePort()
    assert.NoError(t, err)

    config2 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port2),
        EncryptionKey: []byte("1234567890123456"),
        RateLimiterRate:  10,
		RateLimiterBurst: 20,
    }

    node2 := node.NewNode(config2, 24*time.Hour)
    go api.StartServer(config2.P2PAddress, node2)

    time.Sleep(1 * time.Second) // Ensure both servers are up

    // Create the ping message
    pingMsg := message.NewDHTPingMessage()
    serializedPingMsg, err := pingMsg.Serialize()
    assert.NoError(t, err)

    // Use SendMessage from node1 to send the ping to node2
    response, err := node1.DHT.Network.SendMessage("127.0.0.1", port2, serializedPingMsg)
    assert.NoError(t, err)

    // Verify the response
    pongMsg, err := message.DeserializeMessage(response)
    assert.NoError(t, err)

    dhtPong, ok := pongMsg.(*message.DHTPongMessage)
    assert.True(t, ok)
    assert.NotZero(t, dhtPong.Timestamp)
}

// TestTwoNodePut refactored to use SendMessage
func TestTwoNodePut(t *testing.T) {
    // Set up Node 1
    port1, err := tests.GetFreePort()
    assert.NoError(t, err)

    config1 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port1),
        EncryptionKey: []byte("1234567890123456"),
        RateLimiterRate:  10,
		RateLimiterBurst: 20,
    }
    api.InitRateLimiter(config1)

    node1 := node.NewNode(config1, 24*time.Hour)
    go api.StartServer(config1.P2PAddress, node1)

    // Set up Node 2
    port2, err := tests.GetFreePort()
    assert.NoError(t, err)

    config2 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port2),
        EncryptionKey: []byte("1234567890123456"),
    }

    node2 := node.NewNode(config2, 24*time.Hour)
    go api.StartServer(config2.P2PAddress, node2)

    time.Sleep(1 * time.Second) // Ensure both servers are up

    // Create the put message
    var key [32]byte
    copy(key[:], []byte("testkey"))
    value := []byte("testvalue")
    putMsg := message.NewDHTPutMessage(10000, 2, key, value)
    serializedPutMsg, err := putMsg.Serialize()
    assert.NoError(t, err)

    // Use SendMessage from node1 to send the PUT to node2
    response, err := node1.DHT.Network.SendMessage("127.0.0.1", port2, serializedPutMsg)
    assert.NoError(t, err)

    // Verify the response
    successMsg, err := message.DeserializeMessage(response)
    assert.NoError(t, err)

    successPutMsg, ok := successMsg.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successPutMsg.Key)
    assert.Equal(t, value, successPutMsg.Value)
}

// TestTwoNodeGet refactored to use SendMessage
func TestTwoNodeGet(t *testing.T) {
    // Set up Node 1
    port1, err := tests.GetFreePort()
    assert.NoError(t, err)

    config1 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port1),
        EncryptionKey: []byte("1234567890123456"),
        RateLimiterRate:  10,
		RateLimiterBurst: 20,
    }
    api.InitRateLimiter(config1)

    node1 := node.NewNode(config1, 24*time.Hour)
    go api.StartServer(config1.P2PAddress, node1)

    // Set up Node 2
    port2, err := tests.GetFreePort()
    assert.NoError(t, err)

    config2 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port2),
        EncryptionKey: []byte("1234567890123456"),
    }

    node2 := node.NewNode(config2, 24*time.Hour)
    go api.StartServer(config2.P2PAddress, node2)

    time.Sleep(1 * time.Second) // Ensure both servers are up

    // First, send a DHT_PUT message to store a value
    var key [32]byte
    copy(key[:], []byte("testkey"))
    value := []byte("testvalue")
    putMsg := message.NewDHTPutMessage(10000, 2, key, value)
    serializedPutMsg, err := putMsg.Serialize()
    assert.NoError(t, err)

    // Use SendMessage from node1 to send the PUT to node2
    response, err := node1.DHT.Network.SendMessage("127.0.0.1", port2, serializedPutMsg)
    assert.NoError(t, err)

    // Verify the DHT_SUCCESS response for PUT
    successMsg, err := message.DeserializeMessage(response)
    assert.NoError(t, err)

    successPutMsg, ok := successMsg.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successPutMsg.Key)
    assert.Equal(t, value, successPutMsg.Value)

    // Now, send a DHT_GET message to retrieve the value
    getMsg := message.NewDHTGetMessage(key)
    serializedGetMsg, err := getMsg.Serialize()
    assert.NoError(t, err)

    // Use SendMessage from node1 to send the GET to node2
    response, err = node1.DHT.Network.SendMessage("127.0.0.1", port2, serializedGetMsg)
    assert.NoError(t, err)

    // Verify the DHT_SUCCESS response for GET
    successMsg, err = message.DeserializeMessage(response)
    assert.NoError(t, err)

    successGetMsg, ok := successMsg.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successGetMsg.Key)
    assert.Equal(t, value, successGetMsg.Value)
}

// TestTwoNodePutGet refactored to use SendMessage
func TestTwoNodePutGet(t *testing.T) {
    // Set up Node 1
    port1, err := tests.GetFreePort()
    assert.NoError(t, err)

    config1 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port1),
        EncryptionKey: []byte("1234567890123456"),
        RateLimiterRate:  10,
		RateLimiterBurst: 20,
    }
    api.InitRateLimiter(config1)

    node1 := node.NewNode(config1, 24*time.Hour)
    go api.StartServer(config1.P2PAddress, node1)

    // Set up Node 2
    port2, err := tests.GetFreePort()
    assert.NoError(t, err)

    config2 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port2),
        EncryptionKey: []byte("1234567890123456"),
    }

    node2 := node.NewNode(config2, 24*time.Hour)
    go api.StartServer(config2.P2PAddress, node2)

    time.Sleep(1 * time.Second) // Ensure both servers are up

    // 1. Send DHT_PUT message from Node 1 to Node 2
    var key [32]byte
    copy(key[:], []byte("testkey"))
    value := []byte("testvalue")
    putMsg := message.NewDHTPutMessage(10000, 2, key, value)
    serializedPutMsg, err := putMsg.Serialize()
    assert.NoError(t, err)

    // Use SendMessage from node1 to send the PUT to node2
    response, err := node1.DHT.Network.SendMessage("127.0.0.1", port2, serializedPutMsg)
    assert.NoError(t, err)

    // Verify the DHT_SUCCESS response for PUT
    successMsg, err := message.DeserializeMessage(response)
    assert.NoError(t, err)

    successPutMsg, ok := successMsg.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successPutMsg.Key)
    assert.Equal(t, value, successPutMsg.Value)

    // 2. Send DHT_GET message from Node 1 to Node 2
    getMsg := message.NewDHTGetMessage(key)
    serializedGetMsg, err := getMsg.Serialize()
    assert.NoError(t, err)

    // Use SendMessage from node1 to send the GET to node2
    response, err = node1.DHT.Network.SendMessage("127.0.0.1", port2, serializedGetMsg)
    assert.NoError(t, err)

    // Verify the DHT_SUCCESS response for GET
    successMsg, err = message.DeserializeMessage(response)
    assert.NoError(t, err)

    successGetMsg, ok := successMsg.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successGetMsg.Key)
    assert.Equal(t, value, successGetMsg.Value)
}