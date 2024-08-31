package api

import (
	"fmt"
	"log"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

func StartServer(address string, nodeInstance node.NodeInterface) error {
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

		go WithMiddleware(func(c net.Conn) {
			HandleConnection(c, nodeInstance)
		})(conn)
	}
}

func HandleConnection(conn net.Conn, nodeInstance node.NodeInterface) {
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

		log.Printf("Received %d bytes: %x", n, buf[:n])

		msg, err := message.DeserializeMessage(buf[:n])
		if err != nil {
			log.Printf("Failed to deserialize message: %v", err)
			return
		}

		var response []byte

		switch msg.GetType() {
		case message.DHT_PUT:
			response = HandlePut(msg, nodeInstance)
		case message.DHT_GET:
			response = HandleGet(msg, nodeInstance)
		case message.DHT_PING:
			response = HandlePing(msg, nodeInstance)
		case message.DHT_FIND_NODE:
			response = HandleFindNode(msg, nodeInstance)
		case message.DHT_FIND_VALUE:
			response = HandleFindValue(msg, nodeInstance)
		case message.DHT_BOOTSTRAP:
			response = HandleBootstrap(msg, nodeInstance)
		case message.DHT_BOOTSTRAP_REPLY:
			response = HandleBootstrapReply(msg, nodeInstance)
		default:
			log.Printf("Unknown message type: %d", msg.GetType())
			return
		}

		if _, err := conn.Write(response); err != nil {
			log.Printf("Failed to send response: %v", err)
			return
		}

		log.Printf("Sent response: %x", response)
	}
}