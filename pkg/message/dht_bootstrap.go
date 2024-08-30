package message

import (
	"bytes"
)

type DHTBootstrapMessage struct {
	BaseMessage
	Address []byte
}

func NewDHTBootstrapMessage(address string) *DHTBootstrapMessage {
	addressBytes := []byte(address)
	size := uint16(4 + len(addressBytes))
	return &DHTBootstrapMessage{
		BaseMessage: BaseMessage{
			Size: size,
			Type: DHT_BOOTSTRAP,
		},
		Address: addressBytes,
	}
}

func (m *DHTBootstrapMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Address); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTBootstrapMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	m.Address = data[4:]
	return m, nil
}

func (m *DHTBootstrapMessage) GetType() int {
	return m.Type
}