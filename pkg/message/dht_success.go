package message

import (
	"bytes"
	"encoding/binary"
)

type DHTSuccessMessage struct {
	BaseMessage
	Key   [32]byte
	Value []byte
}

func NewDHTSuccessMessage(key [32]byte, value []byte) *DHTSuccessMessage {
	size := uint16(36 + len(value))
	return &DHTSuccessMessage{
		BaseMessage: BaseMessage{
			Size: size,
			Type: DHT_SUCCESS,
		},
		Key:   key,
		Value: value,
	}
}

func (m *DHTSuccessMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Key[:]); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTSuccessMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data[4:])
	if err := binary.Read(reader, binary.BigEndian, &m.Key); err != nil {
		return nil, err
	}
	m.Value = make([]byte, m.Size-36)
	if _, err := reader.Read(m.Value); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DHTSuccessMessage) GetType() int {
	return m.Type
}