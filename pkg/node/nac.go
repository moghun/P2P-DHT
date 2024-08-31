package node

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
)

type NetworkInterface interface {
	SendMessage(targetIP string, targetPort int, data []byte) error
	StartListening() error
}

type Network struct {
	nodeInstance *Node
	Tlsm   *security.TLSManager
}

func NewNetwork(nodeInstance *Node) *Network {
	// Initialize the Tlsmanager within the Network struct, pointing to the .certificates directory
	Tlsm := security.NewTLSManager(".certificates")

	// Ensure certificate and key are available
	err := Tlsm.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize TLS manager: %v", err)
	}

	return &Network{
		nodeInstance: nodeInstance,
		Tlsm:   Tlsm,
	}
}

// SendMessage sends a message to a target node specified by IP and port using TLS.
func (n *Network) SendMessage(targetIP string, targetPort int, data []byte) error {
	addr := fmt.Sprintf("%s:%d", targetIP, targetPort)

	// Use TLS for secure communication
	tlsConfig, err := n.Tlsm.LoadTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to load TLS config: %v", err)
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to target node: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	fmt.Printf("Sent message to %s: %x\n", addr, data)
	return nil
}

// StartListening starts a TLS server to listen for incoming messages from other nodes.
func (n *Network) StartListening() error {
	addr := fmt.Sprintf("%s:%d", n.nodeInstance.IP, n.nodeInstance.Port)

	// Use TLS for secure communication
	tlsConfig, err := n.Tlsm.LoadTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to load TLS config: %v", err)
	}

	listener, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to start listening on %s: %v", addr, err)
	}
	defer listener.Close()

	fmt.Printf("Node listening on %s with TLS\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go n.handleConnection(conn)
	}
}

// handleConnection handles an incoming TLS connection and processes the received data.
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

		// Handle the message
		fmt.Printf("Received message from %s: %x\n", conn.RemoteAddr().String(), buf[:n])
	}
}