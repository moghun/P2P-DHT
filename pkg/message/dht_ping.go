package message

import (
	"bytes"
)

type DHTPingMessage struct {
	BaseMessage
}

func NewDHTPingMessage() *DHTPingMessage {
	return &DHTPingMessage{
		BaseMessage: BaseMessage{
			Size: 4,
			Type: DHT_PING,
		},
	}
}

func (m *DHTPingMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTPingMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DHTPingMessage) GetType() int {
	return m.Type
}