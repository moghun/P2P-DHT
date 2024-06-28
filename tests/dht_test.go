package tests

import (
	"testing"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"github.com/stretchr/testify/assert"
)

func TestProcessMessage(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	message.SetEncryptionKey(key)
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance := dht.NewDHT(node)

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
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance := dht.NewDHT(node)

	peerNode := dht.NewNode("127.0.0.1", 8001, true, key)
	node.AddPeer(peerNode.ID, peerNode.IP, peerNode.Port)

	dhtInstance.StartPeriodicLivenessCheck(1 * time.Second)

	time.Sleep(2 * time.Second)

	peers := node.GetAllPeers()
	assert.Empty(t, peers)
}

func TestCheckLiveness(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance := dht.NewDHT(node)

	assert.False(t, dhtInstance.CheckLiveness("127.0.0.1", 9999, 1*time.Second))
}

func TestDHTHandlePing(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	message.SetEncryptionKey(key)
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance := dht.NewDHT(node)

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
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance := dht.NewDHT(node)

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
	node := dht.NewNode("127.0.0.1", 8000, true, key)
	dhtInstance := dht.NewDHT(node)

	data := []byte("get")
	response := dhtInstance.HandleGet(data)
	responseMsg, err := message.DeserializeMessage(response)
	assert.Nil(t, err)
	assert.Equal(t, message.DHT_SUCCESS, responseMsg.Type)
	assert.Equal(t, []byte("get works"), responseMsg.Data)
}