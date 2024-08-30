package tests

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

func TestBaseMessage_SerializeHeader(t *testing.T) {
	baseMsg := &message.BaseMessage{
		Size: 10,
		Type: message.DHT_PUT,
	}

	var buf bytes.Buffer
	err := baseMsg.SerializeHeader(&buf)
	assert.NoError(t, err)

	expected := []byte{0x00, 0x0A, 0x02, 0x8A} // Size: 10 (0x00, 0x0A), Type: DHT_PUT (0x02, 0x8A)
	assert.Equal(t, expected, buf.Bytes())
}

func TestBaseMessage_DeserializeHeader(t *testing.T) {
	data := []byte{0x00, 0x0A, 0x02, 0x8A} // Size: 10, Type: DHT_PUT

	baseMsg := &message.BaseMessage{}
	err := baseMsg.DeserializeHeader(data)
	assert.NoError(t, err)
	assert.Equal(t, uint16(10), baseMsg.Size)
	assert.Equal(t, message.DHT_PUT, baseMsg.Type)
}

func TestCreateMessage(t *testing.T) {
	msg, err := message.CreateMessage(message.DHT_PUT, []byte("value"))
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTPutMessage{}, msg)

	msg, err = message.CreateMessage(message.DHT_GET, nil)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTGetMessage{}, msg)

	msg, err = message.CreateMessage(message.DHT_PING, nil)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTPingMessage{}, msg)

	msg, err = message.CreateMessage(message.DHT_BOOTSTRAP, []byte("127.0.0.1:8080"))
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTBootstrapMessage{}, msg)

	_, err = message.CreateMessage(9999, nil) // Unknown type
	assert.Error(t, err)
}

func TestDeserializeMessage(t *testing.T) {
	// DHT_PING message
	data := []byte{0x00, 0x04, 0x02, 0x8E} // Size: 4, Type: DHT_PING
	msg, err := message.DeserializeMessage(data)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTPingMessage{}, msg)

	// DHT_PUT message with payload
	key := [32]byte{}
	payload := []byte("testdata")
	putMsg := message.NewDHTPutMessage(10000, 2, key, payload)
	serializedData, err := putMsg.Serialize()
	assert.NoError(t, err)

	msg, err = message.DeserializeMessage(serializedData)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTPutMessage{}, msg)
	assert.Equal(t, payload, msg.(*message.DHTPutMessage).Value)

	// Invalid message data (too short)
	_, err = message.DeserializeMessage([]byte{0x00, 0x01}) // Only 1 byte of data
	assert.Error(t, err)

	// Incomplete DHT_PUT message (header only, missing payload)
	incompleteData := []byte{0x00, 0x08, 0x02, 0x8A} // Size: 8, Type: DHT_PUT (but missing payload)
	_, err = message.DeserializeMessage(incompleteData)
	assert.Error(t, err)
}


func TestDHTPutMessage_SerializeDeserialize(t *testing.T) {
	key := [32]byte{}
	value := []byte("value")
	putMsg := message.NewDHTPutMessage(10000, 2, key, value)

	serializedMsg, err := putMsg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := putMsg.Deserialize(serializedMsg)
	assert.NoError(t, err)
	assert.Equal(t, putMsg.TTL, deserializedMsg.(*message.DHTPutMessage).TTL)
	assert.Equal(t, putMsg.Key, deserializedMsg.(*message.DHTPutMessage).Key)
	assert.Equal(t, putMsg.Value, deserializedMsg.(*message.DHTPutMessage).Value)
}