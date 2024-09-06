package message

import (
	"bytes"
	"encoding/binary"
)

type DHTFindValueMessage struct {
	BaseMessage
	Key [32]byte
}

func NewDHTFindValueMessage(key [32]byte) *DHTFindValueMessage {
	return &DHTFindValueMessage{
		BaseMessage: BaseMessage{
			Size: 36,
			Type: DHT_FIND_VALUE,
		},
		Key: key,
	}
}

func (m *DHTFindValueMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Key[:]); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTFindValueMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data[4:])
	if err := binary.Read(reader, binary.BigEndian, &m.Key); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DHTFindValueMessage) GetType() int {
	return m.Type
}
