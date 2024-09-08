package tests

import (
	"crypto/tls"
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

func TestPingPongIntegration(t *testing.T) {
	port, err := tests.GetFreePort()
	assert.NoError(t, err)

	config := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port),
		EncryptionKey: []byte("1234567890123456"),
	}

	mockNode := node.NewNode(config, 24*time.Hour)
	go api.StartServer(config.P2PAddress, mockNode)

	time.Sleep(1 * time.Second) // Give the server time to start

	conn := setupTLSConnection(t, config.P2PAddress)
	defer conn.Close()

	// Step 1: Send DHT_PING message
	pingMsg := message.NewDHTPingMessage()
	serializedPingMsg, err := pingMsg.Serialize()
	assert.NoError(t, err)

	_, err = conn.Write(serializedPingMsg)
	assert.NoError(t, err)

	// Step 2: Read and verify the PONG response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	assert.NoError(t, err)

	response, err := message.DeserializeMessage(buf[:n])
	assert.NoError(t, err)

	pongMsg := response.(*message.DHTPongMessage)
	assert.NotZero(t, pongMsg.Timestamp)
}

func TestPutMessageIntegration(t *testing.T) {
	port, err := tests.GetFreePort()
	assert.NoError(t, err)

	config := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port),
		EncryptionKey: []byte("1234567890123456"),
	}

	mockNode := node.NewNode(config, 24*time.Hour)
	go api.StartServer(config.P2PAddress, mockNode)

	time.Sleep(1 * time.Second) // Give the server time to start

	conn := setupTLSConnection(t, config.P2PAddress)
	defer conn.Close()

	// Create and send DHT_PUT message
	var key [32]byte
	copy(key[:], []byte("testkey"))
	value := []byte("testvalue")
	putMsg := message.NewDHTPutMessage(10000, 2, key, value)
	serializedPutMsg, err := putMsg.Serialize()
	assert.NoError(t, err)

	_, err = conn.Write(serializedPutMsg)
	assert.NoError(t, err)

	// Read and verify the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	assert.NoError(t, err)

	response, err := message.DeserializeMessage(buf[:n])
	assert.NoError(t, err)

	successMsg, ok := response.(*message.DHTSuccessMessage)
	assert.True(t, ok)
	assert.Equal(t, key, successMsg.Key)
	assert.Equal(t, value, successMsg.Value)
}

func TestGetMessageIntegration(t *testing.T) {
	port, err := tests.GetFreePort()
	assert.NoError(t, err)

	config := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port),
		EncryptionKey: []byte("1234567890123456"),
	}

	mockNode := node.NewNode(config, 24*time.Hour)
	go api.StartServer(config.P2PAddress, mockNode)

	time.Sleep(1 * time.Second) // Give the server time to start

	conn := setupTLSConnection(t, config.P2PAddress)
	defer conn.Close()

	// First, send a DHT_PUT message to store a value
	var key [32]byte
	copy(key[:], []byte("testkey"))
	value := []byte("testvalue")
	putMsg := message.NewDHTPutMessage(10000, 2, key, value)
	serializedPutMsg, err := putMsg.Serialize()
	assert.NoError(t, err)

	_, err = conn.Write(serializedPutMsg)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second) // Ensure the PUT is processed

	// Now, send a DHT_GET message
	getMsg := message.NewDHTGetMessage(key)
	serializedGetMsg, err := getMsg.Serialize()
	assert.NoError(t, err)

	_, err = conn.Write(serializedGetMsg)
	assert.NoError(t, err)

	// Read and verify the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	assert.NoError(t, err)

	response, err := message.DeserializeMessage(buf[:n])
	assert.NoError(t, err)

	successMsg, ok := response.(*message.DHTSuccessMessage)
	assert.True(t, ok)
	assert.Equal(t, key, successMsg.Key)
	assert.Equal(t, value, successMsg.Value)
}

func TestPutGetIntegration(t *testing.T) {
	port, err := tests.GetFreePort()
	assert.NoError(t, err)

	config := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port),
		EncryptionKey: []byte("1234567890123456"),
	}

	mockNode := node.NewNode(config, 24*time.Hour)
	go api.StartServer(config.P2PAddress, mockNode)

	time.Sleep(1 * time.Second) // Give the server time to start

	conn := setupTLSConnection(t, config.P2PAddress)
	defer conn.Close()

	// 1. Send DHT_PUT message
	key := "testkey"
	hashedKey := dht.EnsureKeyHashed(key)
	value := []byte("testvalue")

	byte32Key, err := message.HexStringToByte32(hashedKey)
	assert.NoError(t, err)
	putMsg := message.NewDHTPutMessage(10000, 2, byte32Key, value)
	serializedPutMsg, err := putMsg.Serialize()
	assert.NoError(t, err)

	_, err = conn.Write(serializedPutMsg)
	assert.NoError(t, err)

	// 2. Read the DHT_SUCCESS response from the server
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	assert.NoError(t, err)

	response, err := message.DeserializeMessage(buf[:n])
	assert.NoError(t, err)
	assert.Equal(t, message.DHT_SUCCESS, response.GetType(), "Expected a SUCCESS response")

	// Since the response is a DHTSuccessMessage, cast it
	successMsg, ok := response.(*message.DHTSuccessMessage)
	assert.True(t, ok)
	assert.Equal(t, hashedKey, message.Byte32ToHexEncode(successMsg.Key))
	assert.Equal(t, value, successMsg.Value)

	// 3. Send DHT_GET message
	getMsg := message.NewDHTGetMessage(byte32Key)
	serializedGetMsg, err := getMsg.Serialize()
	assert.NoError(t, err)

	_, err = conn.Write(serializedGetMsg)
	assert.NoError(t, err)

	// 4. Read the DHT_SUCCESS response from the server
	n, err = conn.Read(buf)
	assert.NoError(t, err)

	response, err = message.DeserializeMessage(buf[:n])
	assert.NoError(t, err)
	assert.Equal(t, message.DHT_SUCCESS, response.GetType(), "Expected a SUCCESS response")

	// Since the response is a DHTSuccessMessage, cast it
	successMsg, ok = response.(*message.DHTSuccessMessage)
	assert.True(t, ok)
	assert.Equal(t, hashedKey, message.Byte32ToHexEncode(successMsg.Key))

	unpackedSuccessMsg, err := dht.Deserialize(successMsg.Value)
	assert.NoError(t, err)
	assert.Equal(t, value, []byte(unpackedSuccessMsg.Value))
}

func TestGetNonExistentKeyIntegration(t *testing.T) {
	port, err := tests.GetFreePort()
	assert.NoError(t, err)

	config := &util.Config{
		P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port),
		EncryptionKey: []byte("1234567890123456"),
	}

	mockNode := node.NewNode(config, 24*time.Hour)
	go api.StartServer(config.P2PAddress, mockNode)

	time.Sleep(1 * time.Second) // Give the server time to start

	conn := setupTLSConnection(t, config.P2PAddress)
	defer conn.Close()

	// 1. Send DHT_GET message for a non-existent key
	var key [32]byte
	copy(key[:], []byte("nonexistentkey"))
	getMsg := message.NewDHTGetMessage(key)
	serializedGetMsg, err := getMsg.Serialize()
	assert.NoError(t, err)

	_, err = conn.Write(serializedGetMsg)
	assert.NoError(t, err)

	// 2. Read the DHT_FAILURE response from the server
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	assert.NoError(t, err)

	response, err := message.DeserializeMessage(buf[:n])
	assert.NoError(t, err)
	assert.Equal(t, message.DHT_FAILURE, response.GetType(), "Expected a FAILURE response")

	// Since the response is a DHTFailureMessage, cast it
	failureMsg, ok := response.(*message.DHTFailureMessage)
	assert.True(t, ok)
	assert.Equal(t, key, failureMsg.Key)
}

func setupTLSConnection(t *testing.T, address string) *tls.Conn {
	// Set up the TLS config for the client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Establish a TLS connection to the server
	conn, err := tls.Dial("tcp", address, tlsConfig)
	assert.NoError(t, err)
	return conn
}
