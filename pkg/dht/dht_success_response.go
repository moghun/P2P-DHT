package dht

import (
	"bytes"
	"encoding/binary"
)

// This is used for find node and find value reponses
type SuccessMessageResponse struct {
	Value string
	Nodes []*KNode
}

// Serialize SuccessMessageResponse to byte array
func (smr *SuccessMessageResponse) Serialize() []byte {
	var buf bytes.Buffer

	// Serialize Value field with null-termination
	buf.WriteString(smr.Value)
	buf.WriteByte(0) // Null terminator for Value

	// Serialize each node
	for _, node := range smr.Nodes {
		buf.WriteString(node.ID)
		buf.WriteByte(0) // Null terminator for ID

		buf.WriteString(node.IP)
		buf.WriteByte(0) // Null terminator for IP

		// Serialize Port as 4 bytes (BigEndian format)
		portBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(portBytes, uint32(node.Port))
		buf.Write(portBytes)
	}

	return buf.Bytes()
}

// Deserialize byte array back to SuccessMessageResponse
func Deserialize(data []byte) (*SuccessMessageResponse, error) {
	smr := &SuccessMessageResponse{}
	buf := bytes.NewBuffer(data)

	// Deserialize Value (read until null terminator)
	value, err := buf.ReadString(0)
	if err != nil {
		return nil, err
	}
	smr.Value = value[:len(value)-1] // Remove the null terminator

	// Deserialize Nodes (continue reading until the buffer is empty)
	var nodes []*KNode
	for buf.Len() > 0 {
		node := &KNode{}

		// Deserialize ID
		id, err := buf.ReadString(0)
		if err != nil {
			return nil, err
		}
		node.ID = id[:len(id)-1] // Remove the null terminator

		// Deserialize IP
		ip, err := buf.ReadString(0)
		if err != nil {
			return nil, err
		}
		node.IP = ip[:len(ip)-1] // Remove the null terminator

		// Deserialize Port (read 4 bytes for Port)
		portBytes := make([]byte, 4)
		if _, err := buf.Read(portBytes); err != nil {
			return nil, err
		}
		node.Port = int(binary.BigEndian.Uint32(portBytes))

		nodes = append(nodes, node)
	}

	smr.Nodes = nodes
	return smr, nil
}
