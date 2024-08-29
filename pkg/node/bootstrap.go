package node

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

const bootstrapRetryInterval = 5 * time.Second

// Bootstrap attempts to connect to a bootstrap node to join the network.
func (n *Node) Bootstrap() error {
	for {
		// Try to bootstrap using the known bootstrap nodes from the config
		success := n.tryBootstrap()
		if success {
			fmt.Println("Successfully bootstrapped to the network.")
			return nil
		}

		fmt.Printf("Failed to bootstrap. Retrying in %v...\n", bootstrapRetryInterval)
		time.Sleep(bootstrapRetryInterval)
	}
}

func (n *Node) tryBootstrap() bool {
	bootstrapNodes := n.Config.BootstrapNodes

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(bootstrapNodes), func(i, j int) { bootstrapNodes[i], bootstrapNodes[j] = bootstrapNodes[j], bootstrapNodes[i] })

	for _, bootstrapNode := range bootstrapNodes {
		fmt.Printf("Attempting to bootstrap with node %s:%d...\n", bootstrapNode.IP, bootstrapNode.Port)

		// Create a DHT_PING message to send to the bootstrap node
		pingMessage := message.NewMessage(0, message.DHT_PING, []byte(fmt.Sprintf("%s:%d", n.IP, n.Port)))
		serializedMessage, err := pingMessage.Serialize()
		if err != nil {
			fmt.Printf("Failed to serialize DHT_PING message: %v\n", err)
			continue
		}

		// Send the DHT_PING message to the bootstrap node
		err = n.Network.SendMessage(bootstrapNode.IP, bootstrapNode.Port, serializedMessage)
		if err != nil {
			fmt.Printf("Failed to send DHT_PING message to %s:%d: %v\n", bootstrapNode.IP, bootstrapNode.Port, err)
			continue
		}

		// Wait for a response from the bootstrap node
		response, err := n.Network.ReceiveMessage()
		if err != nil {
			fmt.Printf("Failed to receive response from %s:%d: %v\n", bootstrapNode.IP, bootstrapNode.Port, err)
			continue
		}

		// Process the response to see if it contains the closest nodes
		if response.Type == message.DHT_NODE_REPLY {
			// Assuming the response contains a list of nodes, add them to the routing table
			n.handleNodeReply(response.Data)
			return true
		} else {
			fmt.Printf("Unexpected message type %d received from %s:%d\n", response.Type, bootstrapNode.IP, bootstrapNode.Port)
		}
	}
	return false
}

// handleNodeReply processes the DHT_NODE_REPLY message and adds the nodes to the routing table.
func (n *Node) handleNodeReply(data []byte) {
	// Deserialize the list of nodes from the response
	nodes := deserializeNodes(data)
	for _, nodeInfo := range nodes {
		nodeID := GenerateNodeID(nodeInfo.IP, nodeInfo.Port)
		n.AddPeer(nodeID, nodeInfo.IP, nodeInfo.Port)
		fmt.Printf("Added node %s:%d to routing table.\n", nodeInfo.IP, nodeInfo.Port)
	}
}

// deserializeNodes is a helper function to convert the byte array into a list of nodes.
func deserializeNodes(data []byte) []struct{ IP string; Port int } {
	// Simplified deserialization logic
	nodeStrings := string(data)
	nodeList := []struct{ IP string; Port int }{}
	for _, line := range strings.Split(nodeStrings, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		port, _ := strconv.Atoi(parts[1])
		nodeList = append(nodeList, struct{ IP string; Port int }{IP: parts[0], Port: port})
	}
	return nodeList
}