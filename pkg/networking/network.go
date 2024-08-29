package networking

import (
	"fmt"
	"log"
	"net"
	"sync"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

type Network struct {
	nodeInstance *node.Node
	listener     net.Listener
	mu           sync.Mutex
}

func NewNetwork(nodeInstance *node.Node) *Network {
	return &Network{nodeInstance: nodeInstance}
}

func (n *Network) StartServer(ip string, port int) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	n.mu.Lock()
	n.listener = ln
	n.mu.Unlock()

	fmt.Printf("Server started at %s\n", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			n.mu.Lock()
			if n.listener == nil {
				n.mu.Unlock()
				return nil
			}
			n.mu.Unlock()
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go n.handleConnection(conn)
	}
}

func (n *Network) StopServer() {
	n.mu.Lock()
	if n.listener != nil {
		n.listener.Close()
		n.listener = nil
	}
	n.mu.Unlock()
	log.Println("Server stopped")
}

func (n *Network) GetListeningPort() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.listener == nil {
		return 0
	}
	return n.listener.Addr().(*net.TCPAddr).Port
}

func (n *Network) handleConnection(conn net.Conn) {
	defer conn.Close()

	data := make([]byte, 1024)
	for {
		length, err := conn.Read(data)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		log.Printf("Received data - connection handler: %x", data[:length])

		msg, err := message.DeserializeMessage(data[:length])
		if err != nil {
			log.Printf("Error deserializing message: %v", err)
			return
		}

		// Process the message at the node level
		responseData, err := n.nodeInstance.HandleMessage(msg)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			return
		}

		if responseData != nil {
			// Serialize the response message before sending
			responseMsg := message.NewMessage(uint16(len(responseData.Data)+4), responseData.Type, responseData.Data)
			serializedResponse, err := responseMsg.Serialize()
			if err != nil {
				log.Printf("Error serializing response message: %v", err)
				return
			}

			_, err = conn.Write(serializedResponse)
			if err != nil {
				log.Printf("Error writing response: %v", err)
				return
			}
			log.Printf("Sent response - connection handler: %x", serializedResponse)
		}
	}
}

func (n *Network) SendMessage(targetIP string, targetPort int, message []byte) error {
	addr := fmt.Sprintf("%s:%d", targetIP, targetPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(message)
	if err != nil {
		return err
	}

	fmt.Printf("Sent message to %s: %x\n", addr, message)

	return nil
}