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
	Type int
	Data []byte
}

func NewMessage(t int, data []byte) *Message {
	return &Message{Type: t, Data: data}
}

func (m *Message) Serialize() ([]byte, error) {
	if encryptionKey == nil {
		return nil, errors.New("encryption key not set")
	}
	
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, int16(m.Type)); err != nil {
		return nil, err
	}
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

	buf := bytes.NewReader(decryptedData)
	var msgType int16
	if err := binary.Read(buf, binary.BigEndian, &msgType); err != nil {
		return nil, err
	}

	var dataLen int16
	if err := binary.Read(buf, binary.BigEndian, &dataLen); err != nil {
		return nil, err
	}

	msgData := make([]byte, dataLen)
	if _, err := buf.Read(msgData); err != nil {
		return nil, err
	}

	return &Message{Type: int(msgType), Data: msgData}, nil
}

func SetEncryptionKey(key []byte) {
	encryptionKey = key
}