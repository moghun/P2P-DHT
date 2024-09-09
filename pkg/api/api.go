package api

import (
	"fmt"
	"net"
	"strconv"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func StartServer(tlsAddress, nonTLSAddress string, nodeInstance node.NodeInterface) error {
	// Start a TLS listener
	tlsListener, err := security.StartTLSListener(nodeInstance.GetID(), tlsAddress)
	if err != nil {
		return fmt.Errorf("failed to start TLS server: %v", err)
	}
	defer tlsListener.Close()

	util.Log().Infof("API TLS server running at %s", tlsAddress)

	// Start a separate goroutine to handle TLS connections
	go func() {
		for {
			conn, err := tlsListener.Accept()
			if err != nil {
				util.Log().Errorf("Failed to accept TLS connection: %v", err)
				continue
			}
			go WithMiddleware(func(c net.Conn) {
				HandleConnection(c, nodeInstance) // true means TLS connection
			})(conn)
		}
	}()

	// Conditionally start non-TLS listener if nonTLSAddress is provided
	if nonTLSAddress != "" {
		nonTLSListener, err := net.Listen("tcp", nonTLSAddress)
		if err != nil {
			return fmt.Errorf("failed to start non-TLS server on %s: %v", nonTLSAddress, err)
		}
		defer nonTLSListener.Close()

		util.Log().Infof("Non-TLS server running at %s", nonTLSAddress)

		// Handle non-TLS connections in the main routine
		go func() {
			for {
				conn, err := nonTLSListener.Accept()
				if err != nil {
					util.Log().Errorf("Failed to accept non-TLS connection: %v", err)
					continue
				}
				go WithMiddleware(func(c net.Conn) {
					HandleConnection(c, nodeInstance) // false means non-TLS connection
				})(conn)
			}
		}()
	}

	// Block forever (or until an error occurs)
	select {}
}

// HandleConnection processes incoming TLS connections and dispatches messages
func HandleConnection(conn net.Conn, nodeInstance node.NodeInterface) {
	defer conn.Close()

	// Extract the IP and Port of the client sending the message
	clientAddr := conn.RemoteAddr().String()
	util.Log().Infof("Connection established from %s", clientAddr)

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

		//split clientAddr to get IP and Port

		peerIp, peerPort, err := net.SplitHostPort(clientAddr)
		if err != nil {
			util.Log().Errorf("Failed to split client address: %v", err)
			return
		}
		peerPortInt, err := strconv.Atoi(peerPort)
		if err != nil {
			util.Log().Errorf("Failed to convert peer port to integer: %v", err)
			return
		}
		incomingNodeID := dht.EnsureKeyHashed(clientAddr)
		util.Log().Infof("Adding Peer with node ID: %s", incomingNodeID)
		nodeInstance.AddPeer(incomingNodeID, peerIp, peerPortInt)

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
