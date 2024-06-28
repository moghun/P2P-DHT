package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

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
			size:     uint16(4 + len([]byte("put"))),
			msgType:  message.DHT_PUT,
			data:     []byte("put"),
			expected: []byte("put works"),
		},
		{
			size:     uint16(4 + len([]byte("get"))),
			msgType:  message.DHT_GET,
			data:     []byte("get"),
			expected: []byte("get works"),
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

func TestStartPeriodicLivenessCheck(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node)

	peerNode := dht.NewNode("127.0.0.1", 8001, true, key)
	dhtInstance.JoinNetwork(peerNode)

	dhtInstance.StartPeriodicLivenessCheck(1 * time.Second)

	time.Sleep(2 * time.Second)

	peers := dhtInstance.GetAllPeers()
	assert.Len(t, peers, 1) // Only one peer (node) should remain
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

	data := []byte("put")
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

	data := []byte("get")
	response := dhtInstance.HandleGet(data)
	responseMsg, err := message.DeserializeMessage(response)
	assert.Nil(t, err)
	assert.Equal(t, message.DHT_SUCCESS, responseMsg.Type)
	assert.Equal(t, []byte("get works"), responseMsg.Data)
}

func TestJoinAndLeaveNetwork(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node1 := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node1)

	node2 := dht.NewNode("127.0.0.1", 8001, true, key)
	dhtInstance.JoinNetwork(node2)

	peers := dhtInstance.GetAllPeers()
	assert.Len(t, peers, 2)

	err := dhtInstance.LeaveNetwork(node2)
	assert.Nil(t, err)

	peers = dhtInstance.GetAllPeers()
	assert.Len(t, peers, 1)
}

func TestRemoveNode(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node1 := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.JoinNetwork(node1)

	node2 := dht.NewNode("127.0.0.1", 8001, true, key)
	dhtInstance.JoinNetwork(node2)

	dhtInstance.RemoveNode(node2)

	peers := dhtInstance.GetAllPeers()
	assert.Len(t, peers, 1)
}

func TestAddAndRemoveNodeToBuckets(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance.AddNodeToBuckets(node)

	// Verify the node is added to the buckets
	found := false
	for _, bucket := range dhtInstance.GetKBuckets() {
		if bucket.Contains(node) {
			found = true
			break
		}
	}
	assert.True(t, found, "Node should be found in one of the buckets")

	dhtInstance.RemoveNodeFromBuckets(node)

	// Verify the node is removed from the buckets
	found = false
	for _, bucket := range dhtInstance.GetKBuckets() {
		if bucket.Contains(node) {
			found = true
			break
		}
	}
	assert.False(t, found, "Node should not be found in any bucket")
}