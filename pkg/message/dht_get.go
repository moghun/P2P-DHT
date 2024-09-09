package message

import (
	"bytes"
	"encoding/binary"
)

type DHTGetMessage struct {
	BaseMessage
	Key  [32]byte
	Data []byte
}

func NewDHTGetMessage(key [32]byte, data []byte) *DHTGetMessage {
	return &DHTGetMessage{
		BaseMessage: BaseMessage{
			Size: 36,
			Type: DHT_GET,
		},
		Key:  key,
		Data: data,
	}
}

func (m *DHTGetMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Key[:]); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Data[:]); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTGetMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data[4:])
	if err := binary.Read(reader, binary.BigEndian, &m.Key); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &m.Data); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DHTGetMessage) GetType() int {
	return m.Type
}
