package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

func TestDHTSuccessMessage(t *testing.T) {
	key := [32]byte{}
	value := []byte("success")
	msg := message.NewDHTSuccessMessage(key, value)

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTSuccessMessage{}, deserializedMsg)

	assert.Equal(t, msg.Key, deserializedMsg.(*message.DHTSuccessMessage).Key)
	assert.Equal(t, msg.Value, deserializedMsg.(*message.DHTSuccessMessage).Value)
}

func TestDHTPutMessage(t *testing.T) {
	key := [32]byte{}
	value := []byte("putvalue")
	msg := message.NewDHTPutMessage(10000, 2, key, value)

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTPutMessage{}, deserializedMsg)

	assert.Equal(t, msg.Key, deserializedMsg.(*message.DHTPutMessage).Key)
	assert.Equal(t, msg.Value, deserializedMsg.(*message.DHTPutMessage).Value)
	assert.Equal(t, msg.TTL, deserializedMsg.(*message.DHTPutMessage).TTL)
	assert.Equal(t, msg.Replication, deserializedMsg.(*message.DHTPutMessage).Replication)
}

func TestDHTPongMessage(t *testing.T) {
	msg := message.NewDHTPongMessage()

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTPongMessage{}, deserializedMsg)

	assert.Equal(t, msg.Timestamp, deserializedMsg.(*message.DHTPongMessage).Timestamp)
}

func TestDHTPingMessage(t *testing.T) {
	msg := message.NewDHTPingMessage()

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTPingMessage{}, deserializedMsg)
}

func TestDHTGetMessage(t *testing.T) {
	key := [32]byte{}
	msg := message.NewDHTGetMessage(key)

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTGetMessage{}, deserializedMsg)

	assert.Equal(t, msg.Key, deserializedMsg.(*message.DHTGetMessage).Key)
}

func TestDHTFindValueMessage(t *testing.T) {
	key := [32]byte{}
	msg := message.NewDHTFindValueMessage(key)

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTFindValueMessage{}, deserializedMsg)

	assert.Equal(t, msg.Key, deserializedMsg.(*message.DHTFindValueMessage).Key)
}

func TestDHTFindNodeMessage(t *testing.T) {
	key := [32]byte{}
	msg := message.NewDHTFindNodeMessage(key)

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTFindNodeMessage{}, deserializedMsg)

	assert.Equal(t, msg.Key, deserializedMsg.(*message.DHTFindNodeMessage).Key)
}

func TestDHTFailureMessage(t *testing.T) {
	key := [32]byte{}
	msg := message.NewDHTFailureMessage(key)

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTFailureMessage{}, deserializedMsg)

	assert.Equal(t, msg.Key, deserializedMsg.(*message.DHTFailureMessage).Key)
}

func TestDHTBootstrapMessage(t *testing.T) {
	address := "127.0.0.1:8080"
	msg := message.NewDHTBootstrapMessage(address)

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTBootstrapMessage{}, deserializedMsg)

	assert.Equal(t, msg.Address, deserializedMsg.(*message.DHTBootstrapMessage).Address)
}

func TestDHTBootstrapReplyMessage(t *testing.T) {
	nodes := []byte("127.0.0.1:8081\n127.0.0.2:8082")
	msg := message.NewDHTBootstrapReplyMessage(nodes)

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTBootstrapReplyMessage{}, deserializedMsg)

	assert.Equal(t, msg.Nodes, deserializedMsg.(*message.DHTBootstrapReplyMessage).Nodes)
	assert.Equal(t, msg.Timestamp, deserializedMsg.(*message.DHTBootstrapReplyMessage).Timestamp)

	parsedNodes := deserializedMsg.(*message.DHTBootstrapReplyMessage).ParseNodes()
	assert.Len(t, parsedNodes, 2)
	assert.Equal(t, "127.0.0.1", parsedNodes[0].IP)
	assert.Equal(t, 8081, parsedNodes[0].Port)
	assert.Equal(t, "127.0.0.2", parsedNodes[1].IP)
	assert.Equal(t, 8082, parsedNodes[1].Port)
}