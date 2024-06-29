package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

func TestJoinNetwork(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node1 := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node1)

	node2 := dht.NewNode("127.0.0.1", 8001, true, key)
	dhtInstance.JoinNetwork(node2)

	// Check that both nodes recognize each other
	node1Peers := node1.GetAllPeers()
	node2Peers := node2.GetAllPeers()

	assert.Equal(t, 1, len(node1Peers), "Node1 should have 1 peer")
	assert.Equal(t, node2.ID, node1Peers[0].ID, "Node1 should recognize Node2 as a peer")

	assert.Equal(t, 1, len(node2Peers), "Node2 should have 1 peer")
	assert.Equal(t, node1.ID, node2Peers[0].ID, "Node2 should recognize Node1 as a peer")
}

func TestLeaveNetwork(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node1 := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node1)

	node2 := dht.NewNode("127.0.0.1", 8001, true, key)
	dhtInstance.JoinNetwork(node2)

	err := dhtInstance.LeaveNetwork(node2)
	assert.Nil(t, err, "Node2 should leave the network without error")

	node1Peers := node1.GetAllPeers()
	assert.Equal(t, 0, len(node1Peers), "Node1 should have no peers after Node2 leaves")
}

func TestProcessMessage(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	message.SetEncryptionKey(key)
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node)

	tests := []struct {
		size     uint16
		msgType  int
		data     []byte
		expected []byte
		errMsg   string
	}{
		{
			size:     uint16(4 + len([]byte("ping"))),
			msgType:  message.DHT_PING,
			data:     []byte("ping"),
			expected: []byte("ping"),
		},
		{
			size:     uint16(4 + len([]byte("put:testKey:testValue"))),
			msgType:  message.DHT_PUT,
			data:     []byte("testKey:testValue"),
			expected: []byte("put works"),
		},
		{
			size:     uint16(4 + len([]byte("get:testKey"))),
			msgType:  message.DHT_GET,
			data:     []byte("testKey"),
			expected: []byte("testValue"),
		},
		{
			size:   3, // Less than the minimum required length
			data:   []byte("abc"),
			errMsg: "data too short to process",
		},
		{
			size:    uint16(8), // Mismatch between size and data length
			data:    []byte("abcd"),
			errMsg:  "invalid request type",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			response, err := dhtInstance.ProcessMessage(test.size, test.msgType, test.data)
			if test.errMsg != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.errMsg)
			} else {
				if response != nil {
					responseMsg, err := message.DeserializeMessage(response)
					assert.Nil(t, err)
					assert.Equal(t, test.msgType, responseMsg.Type)
					assert.Equal(t, test.expected, responseMsg.Data)
				}
			}
		})
	}
}

func TestCheckLiveness(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node)

	assert.False(t, dhtInstance.CheckLiveness("127.0.0.1", 9999, 1*time.Second))
}

func TestDHTHandlePing(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	message.SetEncryptionKey(key)
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node)

	data := []byte("ping")
	response := dhtInstance.HandlePing(data)
	responseMsg, err := message.DeserializeMessage(response)
	assert.Nil(t, err)
	assert.Equal(t, message.DHT_PING, responseMsg.Type)
	assert.Equal(t, data, responseMsg.Data)
}

func TestDHTHandlePut(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	message.SetEncryptionKey(key)
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node)

	data := []byte("testKey:testValue")
	response := dhtInstance.HandlePut(data)
	responseMsg, err := message.DeserializeMessage(response)
	assert.Nil(t, err)
	assert.Equal(t, message.DHT_SUCCESS, responseMsg.Type)
	assert.Equal(t, []byte("put works"), responseMsg.Data)
}

func TestDHTHandleGet(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	message.SetEncryptionKey(key)
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node)

	// Store a value first
	err := node.Put("testKey", "testValue", 3600)
	assert.Nil(t, err, "Put operation should not return an error")

	data := []byte("testKey")
	response := dhtInstance.HandleGet(data)
	responseMsg, err := message.DeserializeMessage(response)
	assert.Nil(t, err)
	assert.Equal(t, message.DHT_SUCCESS, responseMsg.Type)
	assert.Equal(t, []byte("testValue"), responseMsg.Data)
}

func TestDhtPut(t *testing.T) {
    key := []byte("12345678901234567890123456789012")
    dhtInstance := dht.NewDHT()

    node1 := dht.NewNode("127.0.0.1", 8000, true, key)
    dhtInstance.JoinNetwork(node1)

    err := dhtInstance.DhtPut("testKey", "testValue", 3600)
    assert.Nil(t, err, "dhtPut should store the value without error")

    value, err := node1.Get("testKey")
    assert.Nil(t, err, "Get operation should not return an error")
    assert.Equal(t, "testValue", value, "Retrieved value should match the stored value")
}

func TestDhtGet(t *testing.T) {
    key := []byte("12345678901234567890123456789012")
    dhtInstance := dht.NewDHT()

    node1 := dht.NewNode("127.0.0.1", 8000, true, key)
    dhtInstance.JoinNetwork(node1)

    err := node1.Put("testKey", "testValue", 3600)
    assert.Nil(t, err, "Put operation should not return an error")

    value, err := dhtInstance.DhtGet("testKey")
    assert.Nil(t, err, "dhtGet should retrieve the value without error")
    assert.Equal(t, "testValue", value, "Retrieved value should match the stored value")
}

func TestGetClosestNodes(t *testing.T) {
    key := []byte("12345678901234567890123456789012")
    dhtInstance := dht.NewDHT()

    node1 := dht.NewNode("127.0.0.1", 8000, true, key)
    node2 := dht.NewNode("127.0.0.1", 8001, true, key)
    dhtInstance.JoinNetwork(node1)
    dhtInstance.JoinNetwork(node2)

    targetID := dht.GenerateNodeID("127.0.0.1", 8002)
    closestNodes := dhtInstance.GetClosestNodes(targetID, 1)

    assert.Equal(t, 1, len(closestNodes), "There should be 1 closest node")
    assert.Contains(t, []*dht.Node{node1, node2}, closestNodes[0], "The closest node should be one of the joined nodes")
}
