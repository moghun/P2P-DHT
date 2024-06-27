package message

import (
	"bytes"
	"encoding/binary"
	"errors"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
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
	DHT_NODE_REPLY
	DHT_FIND_VALUE
)

type Message struct {
	Type int
	Data []byte
}

// Key used for encryption and decryption
var encryptionKey []byte

// SetEncryptionKey sets the encryption key for all messages
func SetEncryptionKey(key []byte) {
	encryptionKey = key
}

func NewMessage(t int, data []byte) *Message {
	return &Message{Type: t, Data: data}
}

func (m *Message) Serialize() ([]byte, error) {
	if encryptionKey == nil {
		return nil, errors.New("encryption key not set")
	}

	encryptedData, err := util.Encrypt(m.Data, encryptionKey)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, int16(m.Type)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, int16(len(encryptedData))); err != nil {
		return nil, err
	}
	if _, err := buf.Write([]byte(encryptedData)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DeserializeMessage(data []byte) (*Message, error) {
	if encryptionKey == nil {
		return nil, errors.New("encryption key not set")
	}

	buf := bytes.NewReader(data)
	var msgType int16
	if err := binary.Read(buf, binary.BigEndian, &msgType); err != nil {
		return nil, err
	}

	var dataLen int16
	if err := binary.Read(buf, binary.BigEndian, &dataLen); err != nil {
		return nil, err
	}

	encryptedData := make([]byte, dataLen)
	if _, err := buf.Read(encryptedData); err != nil {
		return nil, err
	}

	decryptedData, err := util.Decrypt(string(encryptedData), encryptionKey)
	if err != nil {
		return nil, err
	}

	return &Message{Type: int(msgType), Data: decryptedData}, nil
}
