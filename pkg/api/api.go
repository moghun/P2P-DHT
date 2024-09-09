package api

import (
	"fmt"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

// StartServer starts a TLS API server for the peer node
func StartServer(address string, nodeInstance node.NodeInterface) error {
	// Start a TLS listener instead of a regular TCP listener
	listener, err := security.StartTLSListener(nodeInstance.GetID(), address)
	if err != nil {
		return fmt.Errorf("failed to start TLS server: %v", err)
	}
	defer listener.Close()

	util.Log().Infof("API TLS server running at %s", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			util.Log().Errorf("Failed to accept connection: %v", err)
			continue
		}

		go WithMiddleware(func(c net.Conn) {
			HandleConnection(c, nodeInstance)
		})(conn)
	}
}

// HandleConnection processes incoming TLS connections and dispatches messages
func HandleConnection(conn net.Conn, nodeInstance node.NodeInterface) {
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				util.Log().Errorf("Error reading from connection: %v", err)
			}
			return
		}

		util.Log().Infof("(%s) received %d bytes: %x", nodeInstance.GetID(), n, buf[:n])

		msg, err := message.DeserializeMessage(buf[:n])
		if err != nil {
			util.Log().Errorf("Failed to deserialize message: %v", err)
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
		case message.DHT_BOOTSTRAP_REPLY:
			response = HandleBootstrapReply(msg, nodeInstance)
		default:
			util.Log().Errorf("Unknown message type: %d", msg.GetType())
			return
		}

		if _, err := conn.Write(response); err != nil {
			util.Log().Errorf("Failed to send response: %v", err)
			return
		}

		util.Log().Infof("(%s) sent response: %x", nodeInstance.GetID(), response)
	}
}
