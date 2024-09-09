package message

import (
	"bytes"
	"encoding/binary"
)

type DHTPingMessage struct {
	BaseMessage
	Data []byte
}

func NewDHTPingMessage(data []byte) *DHTPingMessage {
	return &DHTPingMessage{
		BaseMessage: BaseMessage{
			Size: 4,
			Type: DHT_PING,
		},
		Data: data,
	}
}

func (m *DHTPingMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Data[:]); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTPingMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data[4:])
	if err := binary.Read(reader, binary.BigEndian, &m.Data); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DHTPingMessage) GetType() int {
	return m.Type
}
