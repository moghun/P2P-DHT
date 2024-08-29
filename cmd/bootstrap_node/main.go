package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/networking"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func main() {
	// Parse command-line parameters
	configPath := flag.String("c", "config.ini", "path to configuration file")
	flag.Parse()

	// Load configuration
	config := util.LoadConfig(*configPath)

	// Set up logging
	util.SetupLogging("bootstrap_node.log")

	// Extract IP and port from the config
	ip, port := config.DHT.GetP2PIPPort()

	// Create the bootstrap node
	bootstrapNode := NewBootstrapNode()

	// Start the network server for the bootstrap node
	network := networking.NewNetwork(bootstrapNode)
	go func() {
		err := network.StartServer(ip, port)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("Received signal %s, shutting down...\n", sig)

	// Graceful shutdown logic
	network.StopServer()
}

// BootstrapNode represents a simplified node used as a bootstrap server.
type BootstrapNode struct {
	KnownNodes map[string]util.BootstrapNode
	mu         sync.Mutex
}

// NewBootstrapNode creates a new bootstrap node with an empty list of known nodes.
func NewBootstrapNode() *BootstrapNode {
	return &BootstrapNode{
		KnownNodes: make(map[string]util.BootstrapNode),
	}
}

// HandleMessage processes incoming messages, registers new nodes, and responds with known nodes.
func (b *BootstrapNode) HandleMessage(msg *message.Message) (*message.Message, error) {
	if msg.Type == message.DHT_FIND_NODE {
		// Serialize the list of known nodes to send back
		b.mu.Lock()
		data := serializeNodes(b.KnownNodes)
		b.mu.Unlock()

		responseMsg := message.NewMessage(uint16(len(data)+4), message.DHT_NODE_REPLY, data)
		return responseMsg, nil
	} else if msg.Type == message.DHT_PING {
		// Deserialize the node info from the message and register it
		nodeInfo := deserializeNodeInfo(msg.Data)
		b.registerNode(nodeInfo)

		// Optionally send a response back acknowledging the registration
		responseMsg := message.NewMessage(0, message.DHT_PONG, nil)
		return responseMsg, nil
	}

	// If message type is not recognized, return nil (no response)
	return nil, fmt.Errorf("unrecognized message type: %d", msg.Type)
}

// registerNode adds a new node to the list of known nodes.
func (b *BootstrapNode) registerNode(node util.BootstrapNode) {
	b.mu.Lock()
	defer b.mu.Unlock()
	nodeKey := fmt.Sprintf("%s:%d", node.IP, node.Port)
	b.KnownNodes[nodeKey] = node
	fmt.Printf("Registered new node: %s\n", nodeKey)
}

// serializeNodes is a helper function to serialize known nodes into a byte slice.
func serializeNodes(nodes map[string]util.BootstrapNode) []byte {
	var data string
	for _, node := range nodes {
		data += fmt.Sprintf("%s:%d\n", node.IP, node.Port)
	}
	return []byte(data)
}

// deserializeNodeInfo is a helper function to convert byte data into a BootstrapNode.
func deserializeNodeInfo(data []byte) util.BootstrapNode {
	nodeString := string(data)
	parts := strings.Split(nodeString, ":")
	port, _ := strconv.Atoi(parts[1])
	return util.BootstrapNode{IP: parts[0], Port: port}
}