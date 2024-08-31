package message

import (
	"bytes"
	"encoding/binary"
)

type DHTPutMessage struct {
	BaseMessage
	TTL         uint16
	Replication uint8
	Reserved    uint8
	Key         [32]byte
	Value       []byte
}

func NewDHTPutMessage(ttl uint16, replication uint8, key [32]byte, value []byte) *DHTPutMessage {
	size := uint16(40 + len(value))
	return &DHTPutMessage{
		BaseMessage: BaseMessage{
			Size: size,
			Type: DHT_PUT,
		},
		TTL:         ttl,
		Replication: replication,
		Key:         key,
		Value:       value,
	}
}

func (m *DHTPutMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, m.TTL); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, m.Replication); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, m.Reserved); err != nil {
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

func (m *DHTPutMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data[4:]) // Skip header

	if err := binary.Read(reader, binary.BigEndian, &m.TTL); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &m.Replication); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &m.Reserved); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &m.Key); err != nil {
		return nil, err
	}
	m.Value = make([]byte, m.Size-40)
	if _, err := reader.Read(m.Value); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DHTPutMessage) GetType() int {
	return m.Type
}