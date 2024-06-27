package tests

import (
	"testing"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func TestSerializeAndDeserializeMessage(t *testing.T) {
	key, _ := util.GenerateRandomKey()
	message.SetEncryptionKey(key)
	originalMessage := message.NewMessage(message.DHT_PUT, []byte("test data"))

	serialized, err := originalMessage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize message: %v", err)
	}

	deserializedMessage, err := message.DeserializeMessage(serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize message: %v", err)
	}

	if originalMessage.Type != deserializedMessage.Type {
		t.Errorf("Expected message type %d, got %d", originalMessage.Type, deserializedMessage.Type)
	}

	if string(originalMessage.Data) != string(deserializedMessage.Data) {
		t.Errorf("Expected message data %s, got %s", originalMessage.Data, deserializedMessage.Data)
	}
}

func TestSerializeAndDeserializeEncryptedMessage(t *testing.T) {
	key, _ := util.GenerateRandomKey()
	message.SetEncryptionKey(key)
	originalMessage := message.NewMessage(message.DHT_PUT, []byte("test data"))

	serialized, err := originalMessage.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize message: %v", err)
	}

	deserializedMessage, err := message.DeserializeMessage(serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize message: %v", err)
	}

	if originalMessage.Type != deserializedMessage.Type {
		t.Errorf("Expected message type %d, got %d", originalMessage.Type, deserializedMessage.Type)
	}

	if string(originalMessage.Data) != string(deserializedMessage.Data) {
		t.Errorf("Expected message data %s, got %s", originalMessage.Data, deserializedMessage.Data)
	}
}
