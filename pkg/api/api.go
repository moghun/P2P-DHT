package api

import (
	"fmt"
	"log"
	"net"
	"sync"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

func StartServer(address string, nodeInstance *node.Node) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("API TCP server running at %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn, nodeInstance)
	}
}

func handleConnection(conn net.Conn, nodeInstance *node.Node) {
	defer conn.Close()

	for {
		// Read the message
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		// Handle the message based on type
		msgType, err := parseMessageType(buf[:n])
		if err != nil {
			log.Printf("Failed to parse message type: %v", err)
			return
		}

		var response []byte
		switch msgType {
		case node.DHT_PUT:
			response = handlePut(buf[:n], nodeInstance)
		case node.DHT_GET:
			response = handleGet(buf[:n], nodeInstance)
		default:
			log.Printf("Unknown message type: %d", msgType)
			return
		}

		// Send the response
		if _, err := conn.Write(response); err != nil {
			log.Printf("Failed to send response: %v", err)
			return
		}
	}
}