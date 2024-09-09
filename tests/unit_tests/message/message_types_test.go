package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
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

func TestDHTStoreMessage(t *testing.T) {

	key := "testkey"
	hashKey := dht.EnsureKeyHashed(key)
	value := "testvalue"

	byte32Key, err := message.HexStringToByte32(hashKey)
	assert.NoError(t, err)
	msg := message.NewDHTStoreMessage(10000, 2, byte32Key, []byte(value))

	serialized, err := msg.Serialize()
	assert.NoError(t, err)

	deserializedMsg, err := message.DeserializeMessage(serialized)
	//deserializedMsg, err := msg.Deserialize(serialized)
	assert.NoError(t, err)
	assert.IsType(t, &message.DHTStoreMessage{}, deserializedMsg)

	assert.Equal(t, msg.Key, deserializedMsg.(*message.DHTStoreMessage).Key)
	assert.Equal(t, msg.Value, deserializedMsg.(*message.DHTStoreMessage).Value)
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
