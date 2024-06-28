package message

import (
	"bytes"
	"encoding/binary"
	"errors"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

var encryptionKey []byte

// Message codes for DHT operations
const (
	DHT_PUT = iota + 650
	DHT_GET
	DHT_SUCCESS
	DHT_FAILURE
	DHT_PING
	DHT_PONG
	DHT_FIND_NODE
	DHT_NODE_REPLY
	DHT_FIND_VALUE
)

type Message struct {
	Size uint16
	Type int
	Data []byte
}

func NewMessage(size uint16, t int, data []byte) *Message {
	return &Message{Size: size, Type: t, Data: data}
}

func (m *Message) Serialize() ([]byte, error) {
	if encryptionKey == nil {
		return nil, errors.New("encryption key not set")
	}

	buf := new(bytes.Buffer)
	// Write the type of the message
	if err := binary.Write(buf, binary.BigEndian, int16(m.Type)); err != nil {
		return nil, err
	}

	// Write the actual data of the message
	if err := binary.Write(buf, binary.BigEndian, int16(len(m.Data))); err != nil {
		return nil, err
	}

	if _, err := buf.Write(m.Data); err != nil {
		return nil, err
	}

	serializedData := buf.Bytes()

	encryptedData, err := util.Encrypt(serializedData, encryptionKey)
	if err != nil {
		return nil, err
	}

	return []byte(encryptedData), nil
}

func DeserializeMessage(data []byte) (*Message, error) {
	if encryptionKey == nil {
		return nil, errors.New("encryption key not set")
	}


	decryptedData, err := util.Decrypt(string(data), encryptionKey)
	if err != nil {
		return nil, err
	}

	if len(decryptedData) < 4 { // 2 bytes for type, 2 bytes for data length
		return nil, errors.New("decrypted data too short")
	}

	buf := bytes.NewReader(decryptedData)

	// Read the type of the message
	var msgType int16
	if err := binary.Read(buf, binary.BigEndian, &msgType); err != nil {
		return nil, err
	}

	// Read the length of the data
	var dataLength int16
	if err := binary.Read(buf, binary.BigEndian, &dataLength); err != nil {
		return nil, err
	}

	msgData := make([]byte, dataLength)
	if _, err := buf.Read(msgData); err != nil {
		return nil, err
	}

	return &Message{Size: uint16(len(msgData)+4),Type: int(msgType), Data: msgData}, nil
}


func SetEncryptionKey(key []byte) {
	encryptionKey = key
}