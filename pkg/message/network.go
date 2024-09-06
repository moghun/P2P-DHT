package message

import (
	"fmt"
	"log"
	"net"
)

// NetworkInterface defines the interface for the Network struct
type NetworkInterface interface {
	SendMessage(targetIP string, targetPort int, data []byte) error
	StartListening() error
}

type Network struct {
	IP   string
	ID   string
	Port int
}

func NewNetwork(ip string, id string, port int) *Network {
	return &Network{
		IP:   ip,
		ID:   id,
		Port: port,
	}
}

// SendMessage sends a message to a target node specified by IP and port and waits for a response.
func (n *Network) SendMessage(targetIP string, targetPort int, data []byte) ([]byte, error) {
	addr := fmt.Sprintf("%s:%d", targetIP, targetPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target node: %v", err)
	}
	defer conn.Close()

	// Send the data
	_, err = conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}

	// Now, wait for a response
	buf := make([]byte, 1024)
	response, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("Received response from %s: %x\n", addr, buf[:response])
	return buf[:response], nil
}

// StartListening starts a TCP server to listen for incoming messages from other nodes.
func (n *Network) StartListening() error {
	addr := fmt.Sprintf("%s:%d", n.IP, n.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start listening on %s: %v", addr, err)
	}
	defer listener.Close()

	fmt.Printf("Node listening on %s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go n.handleConnection(conn)
	}
}

// handleConnection handles an incoming connection and processes the received data.
func (n *Network) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		// Here you would add code to handle the message, such as deserializing it
		// and calling appropriate functions on the node instance.
		fmt.Printf("Received message from %s: %x\n", conn.RemoteAddr().String(), buf[:n])
	}
}
