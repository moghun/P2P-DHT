package api

import (
	"bytes"
	"encoding/binary"
	"log"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

func handlePut(data []byte, nodeInstance *node.Node) []byte {
	// Parse the PUT message
	var size uint16
	var msgType uint16
	var ttl, replication, reserved uint8
	var key [32]byte
	value := data[40:] // Assuming the value starts at byte 40

	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
		log.Printf("Failed to read PUT message size: %v", err)
		return createFailureResponse(key[:])
	}
	if err := binary.Read(reader, binary.BigEndian, &msgType); err != nil {
		log.Printf("Failed to read PUT message type: %v", err)
		return createFailureResponse(key[:])
	}
	if err := binary.Read(reader, binary.BigEndian, &ttl); err != nil {
		log.Printf("Failed to read PUT message TTL: %v", err)
		return createFailureResponse(key[:])
	}
	if err := binary.Read(reader, binary.BigEndian, &replication); err != nil {
		log.Printf("Failed to read PUT message replication: %v", err)
		return createFailureResponse(key[:])
	}
	if err := binary.Read(reader, binary.BigEndian, &reserved); err != nil {
		log.Printf("Failed to read PUT message reserved byte: %v", err)
		return createFailureResponse(key[:])
	}
	if err := binary.Read(reader, binary.BigEndian, &key); err != nil {
		log.Printf("Failed to read PUT message key: %v", err)
		return createFailureResponse(key[:])
	}

	// Store the value in the node's storage
	if err := nodeInstance.Put(string(key[:]), string(value), int(ttl)); err != nil {
		return createFailureResponse(key[:])
	}

	return createSuccessResponse(key[:], value)
}

func handleGet(data []byte, nodeInstance *node.Node) []byte {
	// Parse the GET message
	var size uint16
	var msgType uint16
	var key [32]byte

	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
		log.Printf("Failed to read GET message size: %v", err)
		return createFailureResponse(key[:])
	}
	if err := binary.Read(reader, binary.BigEndian, &msgType); err != nil {
		log.Printf("Failed to read GET message type: %v", err)
		return createFailureResponse(key[:])
	}
	if err := binary.Read(reader, binary.BigEndian, &key); err != nil {
		log.Printf("Failed to read GET message key: %v", err)
		return createFailureResponse(key[:])
	}

	// Retrieve the value from the node's storage
	value, err := nodeInstance.Get(string(key[:]))
	if err != nil || value == "" {
		return createFailureResponse(key[:])
	}

	return createSuccessResponse(key[:], []byte(value))
}

func parseMessageType(data []byte) (uint16, error) {
	var msgType uint16
	buf := bytes.NewReader(data[2:4]) // Assuming type is at byte 2-3
	err := binary.Read(buf, binary.BigEndian, &msgType)
	return msgType, err
}

func createSuccessResponse(key, value []byte) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint16(4+len(key)+len(value)))
	binary.Write(&buf, binary.BigEndian, uint16(message.DHT_SUCCESS))
	buf.Write(key)
	buf.Write(value)
	return buf.Bytes()
}

func createFailureResponse(key []byte) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint16(4+len(key)))
	binary.Write(&buf, binary.BigEndian, uint16(message.DHT_FAILURE))
	buf.Write(key)
	return buf.Bytes()
}
