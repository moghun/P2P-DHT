package message

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Message codes for DHT operations
const (
	DHT_PUT = iota + 650
	DHT_GET
	DHT_SUCCESS
	DHT_FAILURE
	DHT_PING
	DHT_PONG
	DHT_FIND_NODE
	DHT_FIND_VALUE
	DHT_STORE
	DHT_BOOTSTRAP
	DHT_BOOTSTRAP_REPLY
)

// Message interface defines the methods that all message types must implement
type Message interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) (Message, error)
	GetType() int
}

// BaseMessage provides a common structure for all message types
type BaseMessage struct {
	Size uint16
	Type int
}

// SerializeHeader serializes the header (Size and Type) of a message
func (m *BaseMessage) SerializeHeader(buf *bytes.Buffer) error {
	if err := binary.Write(buf, binary.BigEndian, m.Size); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, int16(m.Type)); err != nil {
		return err
	}
	return nil
}

// DeserializeHeader deserializes the header (Size and Type) of a message
func (m *BaseMessage) DeserializeHeader(data []byte) error {
	if len(data) < 4 {
		return errors.New("data too short to contain a valid message header")
	}
	m.Size = binary.BigEndian.Uint16(data[:2])
	m.Type = int(binary.BigEndian.Uint16(data[2:4]))
	return nil
}

// Serialize serializes the entire BaseMessage, including its header.
func (m *BaseMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CreateMessage is a factory function for creating new messages when you know the type
func CreateMessage(msgType int, data []byte) (Message, error) {
	switch msgType {
	case DHT_PUT:
		return NewDHTPutMessage(0, 0, [32]byte{}, data), nil // Placeholder for initialization
	case DHT_GET:
		return NewDHTGetMessage([32]byte{}), nil // Placeholder for initialization
	case DHT_SUCCESS:
		return NewDHTSuccessMessage([32]byte{}, data), nil // Placeholder for initialization
	case DHT_FAILURE:
		return NewDHTFailureMessage([32]byte{}), nil // Placeholder for initialization
	case DHT_PING:
		return NewDHTPingMessage(), nil
	case DHT_PONG:
		return NewDHTPongMessage(), nil
	case DHT_FIND_NODE:
		return NewDHTFindNodeMessage([32]byte{}), nil // Placeholder for initialization
	case DHT_FIND_VALUE:
		return NewDHTFindValueMessage([32]byte{}), nil // Placeholder for initialization
	case DHT_STORE:
		return NewDHTStoreMessage(0, 0, [32]byte{}, data), nil
	case DHT_BOOTSTRAP:
		return NewDHTBootstrapMessage(string(data)), nil
	case DHT_BOOTSTRAP_REPLY:
		return NewDHTBootstrapReplyMessage(data), nil
	default:
		return nil, errors.New("unknown message type")
	}
}

// DeserializeMessage is a factory function for deserializing a message when you don't know the type
func DeserializeMessage(data []byte) (Message, error) {
	if len(data) < 4 { // Ensuring there's enough data for the header
		return nil, errors.New("data too short to contain a valid message header")
	}

	// Extract the message type from the header
	msgType := int(binary.BigEndian.Uint16(data[2:4]))
	// Create the appropriate message based on the type
	msg, err := CreateMessage(msgType, nil)
	if err != nil {
		return nil, err
	}

	// Deserialize the full message (including the header)
	return msg.Deserialize(data)
}

// Convert string id to [32]byte
func StringToByte32(id string) [32]byte {
	var byteID [32]byte
	copy(byteID[:], id)
	return byteID
}

// Convert [32]byte to string
func Byte32ToString(id [32]byte) string {
	return string(id[:])
}
