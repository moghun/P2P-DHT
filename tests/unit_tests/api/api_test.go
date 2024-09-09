package tests

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func setupTLSConnection(t *testing.T, address string) *tls.Conn {
	// Set up the TLS config for the client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Skip certificate verification for testing purposes
	}

	// Establish a TLS connection to the server
	conn, err := tls.Dial("tcp", address, tlsConfig)
	assert.NoError(t, err)
	return conn
}
func TestStartServer(t *testing.T) {
	// Dynamic port allocation
	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")
	mockNode := NewMockNode("127.0.0.1", port)

	config := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port),
		EncryptionKey: []byte("12345678901234567890123456789012"),
        RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty: 4,
    }

	go func() {
		api.InitRateLimiter(config)
		err := api.StartServer(fmt.Sprintf("127.0.0.1:%d", port), "", &mockNode.Node)
		assert.NoError(t, err, "Failed to start API server")
	}()

	// Give the server time to start
	time.Sleep(2 * time.Second)

	conn := setupTLSConnection(t, fmt.Sprintf("127.0.0.1:%d", port))
	defer conn.Close()

	// Prepare a 32-byte node ID
	nodeID := make([]byte, 32)
	copy(nodeID, []byte("my-node-id")) // Example node ID, padded with zeros

	// Send the 32-byte node ID first
	_, err = conn.Write(nodeID)
	assert.NoError(t, err, "Failed to send node ID to the server")

	// Send a PING message to test the server
	pingMsg := message.NewDHTPingMessage()
	serializedMsg, err := pingMsg.Serialize()
	assert.NoError(t, err, "Failed to serialize message")
	_, err = conn.Write(serializedMsg)
	assert.NoError(t, err, "Failed to send message to the server")

	// Read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	assert.NoError(t, err, "Failed to read response from the server")

	if n == 0 {
		t.Fatal("No data received from the server")
	}

	// Deserialize the response
	response, err := message.DeserializeMessage(buf[:n])
	assert.NoError(t, err, "Failed to deserialize response from the server")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, message.DHT_PONG, response.GetType(), "Expected a PONG response")
}

func TestHandleConnection(t *testing.T) {
	// Dynamic port allocation
	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")
	mockNode := NewMockNode("127.0.0.1", port)

	conn := &MockConn{
		readData: make([]byte, 1024),
	}

	t.Run("TestHandlePut", func(t *testing.T) {
		// Prepare a 32-byte node ID
		nodeID := make([]byte, 32)
		copy(nodeID, []byte("my-node-id"))

		// Serialize the PUT message
		key := [32]byte{}
		value := []byte("value")
		putMsg := message.NewDHTPutMessage(10000, 2, key, value)
		serializedMsg, err := putMsg.Serialize()
		assert.NoError(t, err, "Failed to serialize message")

		// Prepend the node ID to the serialized message
		conn.readData = append(nodeID, serializedMsg...)
		api.HandlePeerConnection(conn, &mockNode.Node)
		assert.Greater(t, len(conn.writeData), 0, "Expected writeData to have at least 1 byte, got 0 bytes")
	})

	t.Run("TestHandleGet", func(t *testing.T) {
		// Prepare a 32-byte node ID
		nodeID := make([]byte, 32)
		copy(nodeID, []byte("my-node-id"))

		// Serialize the GET message
		key := [32]byte{}
		getMsg := message.NewDHTGetMessage(key)
		serializedMsg, err := getMsg.Serialize()
		assert.NoError(t, err, "Failed to serialize message")

		// Prepend the node ID to the serialized message
		conn.readData = append(nodeID, serializedMsg...)
		api.HandlePeerConnection(conn, &mockNode.Node)
		assert.Greater(t, len(conn.writeData), 0, "Expected writeData to have at least 1 byte, got 0 bytes")
	})

	t.Run("TestHandlePing", func(t *testing.T) {
		// Prepare a 32-byte node ID
		nodeID := make([]byte, 32)
		copy(nodeID, []byte("my-node-id"))

		// Serialize the PING message
		pingMsg := message.NewDHTPingMessage()
		serializedMsg, err := pingMsg.Serialize()
		assert.NoError(t, err, "Failed to serialize message")

		// Prepend the node ID to the serialized message
		conn.readData = append(nodeID, serializedMsg...)
		api.HandlePeerConnection(conn, &mockNode.Node)
		assert.Greater(t, len(conn.writeData), 0, "Expected writeData to have at least 1 byte, got 0 bytes")
	})

	t.Run("TestHandleInvalidMessage", func(t *testing.T) {
		// Prepare a 32-byte node ID
		nodeID := make([]byte, 32)
		copy(nodeID, []byte("my-node-id"))

		// Prepend the node ID to invalid data (which is also a byte slice)
		invalidData := []byte{0x00, 0x01} // Invalid data

		// Concatenate the node ID and the invalid data
		conn.readData = append(nodeID, invalidData...)

		api.HandlePeerConnection(conn, &mockNode.Node)
		assert.NotEqual(t, 0, len(conn.writeData), "Expected writeData to have a response for an invalid message")
	})
}