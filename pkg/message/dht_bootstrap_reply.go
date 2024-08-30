package message

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
	"time"
)

type DHTBootstrapReplyMessage struct {
	BaseMessage
	Timestamp uint64
	Nodes     []byte
}

func NewDHTBootstrapReplyMessage(nodes []byte) *DHTBootstrapReplyMessage {
	size := uint16(12 + len(nodes)) // Header + nodes
	return &DHTBootstrapReplyMessage{
		BaseMessage: BaseMessage{
			Size: size,
			Type: DHT_BOOTSTRAP_REPLY,
		},
		Timestamp: uint64(time.Now().UnixNano()),
		Nodes:     nodes,
	}
}

func (m *DHTBootstrapReplyMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.SerializeHeader(buf); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, m.Timestamp); err != nil {
		return nil, err
	}
	if _, err := buf.Write(m.Nodes); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *DHTBootstrapReplyMessage) Deserialize(data []byte) (Message, error) {
	if err := m.DeserializeHeader(data); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data[4:])
	if err := binary.Read(reader, binary.BigEndian, &m.Timestamp); err != nil {
		return nil, err
	}
	m.Nodes = data[12:] // Nodes start after the timestamp
	return m, nil
}

// ParseNodes parses the nodes from the bootstrap reply message into a list of structs
func (m *DHTBootstrapReplyMessage) ParseNodes() []struct{ IP string; Port int } {
	nodeStrings := strings.Split(string(m.Nodes), "\n")
	nodeList := []struct{ IP string; Port int }{}
	for _, line := range nodeStrings {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		port, _ := strconv.Atoi(parts[1])
		nodeList = append(nodeList, struct{ IP string; Port int }{IP: parts[0], Port: port})
	}
	return nodeList
}

func (m *DHTBootstrapReplyMessage) GetType() int {
	return m.Type
}