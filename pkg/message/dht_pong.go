package message

import (
	"bytes"
	"encoding/binary"
	"time"
)

type DHTPongMessage struct {
	BaseMessage
	Timestamp uint64
}

func NewDHTPongMessage() *DHTPongMessage {
	return &DHTPongMessage{
		BaseMessage: BaseMessage{
			Size: 12,
			Type: DHT_PONG,
		},
		Timestamp: uint64(time.Now().UnixNano()),
	}
}

func (m *DHTPongMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, m.Timestamp); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTPongMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data[4:])
	if err := binary.Read(reader, binary.BigEndian, &m.Timestamp); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DHTPongMessage) GetType() int {
	return m.Type
}