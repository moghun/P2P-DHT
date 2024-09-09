package api

import (
	"fmt"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func StartServer(tlsAddress, nonTLSAddress string, nodeInstance node.NodeInterface) error {
	// Start a TLS listener for peer connections
	tlsListener, err := security.StartTLSListener(nodeInstance.GetID(), tlsAddress)
	if err != nil {
		return fmt.Errorf("failed to start TLS server: %v", err)
	}
	defer tlsListener.Close()

	util.Log().Infof("Peer TLS server running at %s", tlsAddress)

	// Start a separate goroutine to handle TLS (Peer) connections
	go func() {
		for {
			conn, err := tlsListener.Accept()
			if err != nil {
				util.Log().Errorf("Failed to accept TLS connection: %v", err)
				continue
			}
			go WithMiddleware(func(c net.Conn) {
				HandlePeerConnection(c, nodeInstance) // Handle peer connections (TLS)
			})(conn)
		}
	}()

	// Conditionally start non-TLS listener for client connections
	if nonTLSAddress != "" {
		nonTLSListener, err := net.Listen("tcp", nonTLSAddress)
		if err != nil {
			return fmt.Errorf("failed to start non-TLS server on %s: %v", nonTLSAddress, err)
		}
		defer nonTLSListener.Close()

		util.Log().Infof("Client non-TLS server running at %s", nonTLSAddress)

		// Handle non-TLS (Client) connections
		go func() {
			for {
				conn, err := nonTLSListener.Accept()
				if err != nil {
					util.Log().Errorf("Failed to accept non-TLS connection: %v", err)
					continue
				}
				go WithMiddleware(func(c net.Conn) {
					HandleClientConnection(c, nodeInstance) // Handle client connections (non-TLS)
				})(conn)
			}
		}()
	}

	// Block forever (or until an error occurs)
	select {}
}
// HandlePeerConnection processes incoming TLS connections from peers
func HandlePeerConnection(conn net.Conn, nodeInstance node.NodeInterface) {
	defer conn.Close()

	// Read the 32-byte node ID from the connection
	idBuf := make([]byte, 32)
	_, err := conn.Read(idBuf)
	if err != nil {
		util.Log().Errorf("Failed to read node ID: %v", err)
		return
	}

	peerNodeID := string(idBuf) // The peer's node ID (could be padded with zeros)
	peerAddr := conn.RemoteAddr().String() // Get the IP and Port of the peer connection

	util.Log().Infof("Connection established from peer with ID: %s at %s", peerNodeID, peerAddr)

	// Now handle the actual message
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				util.Log().Errorf("Error reading from peer (%s) at %s: %v", peerNodeID, peerAddr, err)
			}
			return
		}

		util.Log().Infof("(%s) received %d bytes from peer (%s) at %s: %x", nodeInstance.GetID(), n, peerNodeID, peerAddr, buf[:n])

		msg, err := message.DeserializeMessage(buf[:n])
		if err != nil {
			util.Log().Errorf("Failed to deserialize message from peer (%s) at %s: %v", peerNodeID, peerAddr, err)
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
		case message.DHT_STORE:
			response = HandleStore(msg, nodeInstance)
		case message.DHT_BOOTSTRAP:
			response = HandleBootstrap(msg, nodeInstance)
		default:
			util.Log().Errorf("Unknown message type from peer (%s) at %s: %d", peerNodeID, peerAddr, msg.GetType())
			return
		}

		if _, err := conn.Write(response); err != nil {
			util.Log().Errorf("Failed to send response to peer (%s) at %s: %v", peerNodeID, peerAddr, err)
			return
		}

		util.Log().Infof("(%s) sent response to peer (%s) at %s: %x", nodeInstance.GetID(), peerNodeID, peerAddr, response)
	}
}

// HandleClientConnection processes incoming non-TLS client connections
func HandleClientConnection(conn net.Conn, nodeInstance node.NodeInterface) {
	defer conn.Close()

	// Log the client address (IP and Port)
	clientAddr := conn.RemoteAddr().String()
	util.Log().Infof("Client connection established from %s", clientAddr)

	// Handle the client's messages
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				util.Log().Errorf("Error reading from client at %s: %v", clientAddr, err)
			}
			return
		}

		util.Log().Infof("(%s) received %d bytes from client at %s: %x", nodeInstance.GetID(), n, clientAddr, buf[:n])

		msg, err := message.DeserializeMessage(buf[:n])
		if err != nil {
			util.Log().Errorf("Failed to deserialize message from client at %s: %v", clientAddr, err)
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
		case message.DHT_STORE:
			response = HandleStore(msg, nodeInstance)
		case message.DHT_BOOTSTRAP:
			response = HandleBootstrap(msg, nodeInstance)
		default:
			util.Log().Errorf("Unknown message type from client at %s: %d", clientAddr, msg.GetType())
			return
		}

		if _, err := conn.Write(response); err != nil {
			util.Log().Errorf("Failed to send response to client at %s: %v", clientAddr, err)
			return
		}

		util.Log().Infof("(%s) sent response to client at %s: %x", nodeInstance.GetID(), clientAddr, response)
	}
}