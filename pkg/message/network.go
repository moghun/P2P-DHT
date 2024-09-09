package message

import (
	"fmt"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

// NetworkInterface defines the interface for the Network struct
type NetworkInterface interface {
    SendMessage(targetIP string, targetPort int, data []byte) ([]byte, error)
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

// SendMessage sends a message to a target node specified by IP and port using TLS, and waits for a response.
func (n *Network) SendMessage(targetIP string, targetPort int, data []byte) ([]byte, error) {
    address := fmt.Sprintf("%s:%d", targetIP, targetPort)

    // Using TLS to dial the target node
    conn, err := security.DialTLS(n.ID, address)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to target node via TLS: %v", err)
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

    util.Log().Printf("Received response from %s: %x\n", address, buf[:response])
    return buf[:response], nil
}

// StartListening starts a TLS server to listen for incoming messages from other nodes.
func (n *Network) StartListening() error {
    address := fmt.Sprintf("%s:%d", n.IP, n.Port)

    // Start a TLS listener instead of a regular TCP listener
    listener, err := security.StartTLSListener(n.ID, address)
    if err != nil {
        return fmt.Errorf("failed to start TLS listener: %v", err)
    }
    defer listener.Close()

    util.Log().Printf("Node listening on %s (TLS)\n", address)

    for {
        conn, err := listener.Accept()
        if err != nil {
            util.Log().Errorf("Failed to accept connection: %v", err)
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
                util.Log().Errorf("Error reading from connection: %v", err)
            }
            if _, err := conn.Write([]byte("Failed")); err != nil {
                util.Log().Errorf("Failed to send response: %v", err)
                return
            }
        }

        util.Log().Printf("Received message from %s: %x\n", conn.RemoteAddr().String(), buf[:n])
        if _, err := conn.Write([]byte("Success")); err != nil {
            util.Log().Errorf("Failed to send response: %v", err)
            return
        }
    }
}
