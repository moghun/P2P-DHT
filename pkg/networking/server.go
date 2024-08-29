package networking

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

// Server represents a simple TCP server that listens for incoming connections
type Server struct {
	node *node.Node
}

// NewServer creates a new Server instance
func NewServer(node *node.Node) *Server {
	return &Server{node: node}
}

// Start starts the TCP server
func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Server running at %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles an incoming connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		msgType, err := parseMessageType(buf[:n])
		if err != nil {
			log.Printf("Failed to parse message type: %v", err)
			return
		}

		var response []byte
		switch msgType {
		case message.DHT_PUT:
			response = s.handlePut(buf[:n])
		case message.DHT_GET:
			response = s.handleGet(buf[:n])
		default:
			log.Printf("Unknown message type: %d", msgType)
			return
		}

		if _, err := conn.Write(response); err != nil {
			log.Printf("Failed to send response: %v", err)
			return
		}
	}
}

// handlePut handles a DHT_PUT message
func (s *Server) handlePut(data []byte) []byte {
	var size uint16
	var msgType uint16
	var ttl, replication, reserved uint8
	var key [32]byte
	value := data[40:]

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
	if err := s.node.Put(string(key[:]), string(value), int(ttl)); err != nil {
		return createFailureResponse(key[:])
	}

	return createSuccessResponse(key[:], value)
}

// handleGet handles a DHT_GET message
func (s *Server) handleGet(data []byte) []byte {
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
	value, err := s.node.Get(string(key[:]))
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