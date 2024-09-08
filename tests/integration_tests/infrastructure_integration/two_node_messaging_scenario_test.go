package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func TestTwoNodePingPong(t *testing.T) {
    // Set up Node 1
    port1, err := tests.GetFreePort()
    assert.NoError(t, err)

    config1 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port1),
        EncryptionKey: []byte("1234567890123456"),
    }

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

    // Connect to Node 2 from Node 1 and send ping
    conn1, err := security.DialTLS("node1", config2.P2PAddress)
    assert.NoError(t, err)
    defer conn1.Close()

    pingMsg := message.NewDHTPingMessage()
    serializedPingMsg, err := pingMsg.Serialize()
    assert.NoError(t, err)

    _, err = conn1.Write(serializedPingMsg)
    assert.NoError(t, err)

    // Read DHT_PONG response
    buf := make([]byte, 1024)
    n, err := conn1.Read(buf)
    assert.NoError(t, err)

    response, err := message.DeserializeMessage(buf[:n])
    assert.NoError(t, err)

    pongMsg, ok := response.(*message.DHTPongMessage)
    assert.True(t, ok)
    assert.NotZero(t, pongMsg.Timestamp)
}

func TestTwoNodePut(t *testing.T) {
    // Set up Node 1
    port1, err := tests.GetFreePort()
    assert.NoError(t, err)

    config1 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port1),
        EncryptionKey: []byte("1234567890123456"),
    }

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

    // Connect to Node 2 from Node 1 and send a PUT message
    conn1 := setupTLSConnection(t, config2.P2PAddress)
    defer conn1.Close()

    // Create and send DHT_PUT message
    var key [32]byte
    copy(key[:], []byte("testkey"))
    value := []byte("testvalue")
    putMsg := message.NewDHTPutMessage(10000, 2, key, value)
    serializedPutMsg, err := putMsg.Serialize()
    assert.NoError(t, err)

    _, err = conn1.Write(serializedPutMsg)
    assert.NoError(t, err)

    // Read and verify the response
    buf := make([]byte, 1024)
    n, err := conn1.Read(buf)
    assert.NoError(t, err)

    response, err := message.DeserializeMessage(buf[:n])
    assert.NoError(t, err)

    successMsg, ok := response.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successMsg.Key)
    assert.Equal(t, value, successMsg.Value)
}

func TestTwoNodeGet(t *testing.T) {
    // Set up Node 1
    port1, err := tests.GetFreePort()
    assert.NoError(t, err)

    config1 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port1),
        EncryptionKey: []byte("1234567890123456"),
    }

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
    conn1 := setupTLSConnection(t, config2.P2PAddress)
    defer conn1.Close()

    var key [32]byte
    copy(key[:], []byte("testkey"))
    value := []byte("testvalue")
    putMsg := message.NewDHTPutMessage(10000, 2, key, value)
    serializedPutMsg, err := putMsg.Serialize()
    assert.NoError(t, err)

    _, err = conn1.Write(serializedPutMsg)
    assert.NoError(t, err)

    time.Sleep(1 * time.Second) // Ensure the PUT is processed

    // Now, send a DHT_GET message
    getMsg := message.NewDHTGetMessage(key)
    serializedGetMsg, err := getMsg.Serialize()
    assert.NoError(t, err)

    _, err = conn1.Write(serializedGetMsg)
    assert.NoError(t, err)

    // Read and verify the response
    buf := make([]byte, 1024)
    n, err := conn1.Read(buf)
    assert.NoError(t, err)

    response, err := message.DeserializeMessage(buf[:n])
    assert.NoError(t, err)

    successMsg, ok := response.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successMsg.Key)
    assert.Equal(t, value, successMsg.Value)
}


func TestTwoNodePutGet(t *testing.T) {
    // Set up Node 1
    port1, err := tests.GetFreePort()
    assert.NoError(t, err)

    config1 := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port1),
        EncryptionKey: []byte("1234567890123456"),
    }

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
    conn1 := setupTLSConnection(t, config2.P2PAddress)
    defer conn1.Close()

    var key [32]byte
    copy(key[:], []byte("testkey"))
    value := []byte("testvalue")
    putMsg := message.NewDHTPutMessage(10000, 2, key, value)
    serializedPutMsg, err := putMsg.Serialize()
    assert.NoError(t, err)

    _, err = conn1.Write(serializedPutMsg)
    assert.NoError(t, err)

    // 2. Read the DHT_SUCCESS response from the server
    buf := make([]byte, 1024)
    n, err := conn1.Read(buf)
    assert.NoError(t, err)

    response, err := message.DeserializeMessage(buf[:n])
    assert.NoError(t, err)
    successMsg, ok := response.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successMsg.Key)
    assert.Equal(t, value, successMsg.Value)

    // 3. Send DHT_GET message from Node 1 to Node 2
    getMsg := message.NewDHTGetMessage(key)
    serializedGetMsg, err := getMsg.Serialize()
    assert.NoError(t, err)

    _, err = conn1.Write(serializedGetMsg)
    assert.NoError(t, err)

    // 4. Read the DHT_SUCCESS response from the server
    n, err = conn1.Read(buf)
    assert.NoError(t, err)

    response, err = message.DeserializeMessage(buf[:n])
    assert.NoError(t, err)
    successMsg, ok = response.(*message.DHTSuccessMessage)
    assert.True(t, ok)
    assert.Equal(t, key, successMsg.Key)
    assert.Equal(t, value, successMsg.Value)
}

