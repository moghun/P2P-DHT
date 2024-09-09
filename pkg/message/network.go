package message

import (
	"encoding/binary"
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

    // Extract the message type from the data and log it
    if len(data) >= 4 { // Assuming message type is stored in the first 4 bytes of the data
        messageType := int(binary.BigEndian.Uint16(data[2:4])) // Adjust this if message type is at a different position
        messageTypeName := getMessageTypeName(messageType)
        util.Log().Printf("Node (%s) sent message of type: %s to %s", n.ID, messageTypeName, address)
    } else {
        util.Log().Printf("Node (%s) sent message to %s, but couldn't determine message type (data too short)", n.ID, address)
    }

    // Now, wait for a response
    buf := make([]byte, 1024)
    response, err := conn.Read(buf)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %v", err)
    }

    util.Log().Printf("Node (%s) received response from %s: %x", n.ID, address, buf[:response])
    return buf[:response], nil
}

// getMessageTypeName returns the message type name for a given message type code.
func getMessageTypeName(messageType int) string {
    switch messageType {
    case DHT_PUT:
        return "DHT_PUT"
    case DHT_GET:
        return "DHT_GET"
    case DHT_SUCCESS:
        return "DHT_SUCCESS"
    case DHT_FAILURE:
        return "DHT_FAILURE"
    case DHT_PING:
        return "DHT_PING"
    case DHT_PONG:
        return "DHT_PONG"
    case DHT_FIND_NODE:
        return "DHT_FIND_NODE"
    case DHT_FIND_VALUE:
        return "DHT_FIND_VALUE"
    case DHT_STORE:
        return "DHT_STORE"
    case DHT_BOOTSTRAP:
        return "DHT_BOOTSTRAP"
    case DHT_BOOTSTRAP_REPLY:
        return "DHT_BOOTSTRAP_REPLY"
    default:
        return fmt.Sprintf("Unknown message type: %d", messageType)
    }
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

    util.Log().Printf("Node (%s) listening on %s (TLS)", n.ID, address)

    for {
        conn, err := listener.Accept()
        if err != nil {
            util.Log().Errorf("Node (%s) failed to accept connection: %v", n.ID, err)
            continue
        }
        go n.handleConnection(conn)
    }
}

// handleConnection handles an incoming TLS connection and processes the received data.
func (n *Network) handleConnection(conn net.Conn) {
    defer conn.Close()
    node_id := n.ID
    buf := make([]byte, 1024)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            if err.Error() != "EOF" {
                util.Log().Errorf("Error reading from connection: %v", err)
            }
            if _, err := conn.Write([]byte("Failed")); err != nil {
                util.Log().Errorf("Node (%s) failed to send response: %v", node_id, err)
                return
            }
        }

        util.Log().Printf("Node (%s) received message from %s: %x", node_id, conn.RemoteAddr().String(), buf[:n])
        if _, err := conn.Write([]byte("Success")); err != nil {
            util.Log().Errorf("Failed to send response: %v", err)
            return
        }
    }
}
