package message

import (
	"bytes"
	"encoding/binary"
)

type DHTFindNodeMessage struct {
	BaseMessage
	Key [32]byte
}

func NewDHTFindNodeMessage(key [32]byte) *DHTFindNodeMessage {
	return &DHTFindNodeMessage{
		BaseMessage: BaseMessage{
			Size: 36,
			Type: DHT_FIND_NODE,
		},
		Key: key,
	}
}

func (m *DHTFindNodeMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Key[:]); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTFindNodeMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data[4:])
	if err := binary.Read(reader, binary.BigEndian, &m.Key); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DHTFindNodeMessage) GetType() int {
	return m.Type
}
