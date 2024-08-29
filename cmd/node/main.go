package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func main() {
	// Parse command-line parameters
	configPath := flag.String("c", "config.ini", "path to configuration file")
	flag.Parse()

	// Load configuration
	config := util.LoadConfig(*configPath)

	// Set up logging
	util.SetupLogging("kademlia.log")

	// Extract IP and port from the config
	ip, port := config.DHT.GetP2PIPPort()

	// Create the node
	nodeInstance := node.NewNode(ip, port, true, []byte("encryption-key-placeholder"))

	// Bootstrap the node with known peers (this could be from the config or hardcoded for now)
	bootstrapNodes := []*node.Node{
		// Example bootstrap nodes; in a real-world scenario, these would come from a config or discovery mechanism
		node.NewNode("127.0.0.1", 8001, false, []byte("bootstrap-key1")),
		node.NewNode("127.0.0.1", 8002, false, []byte("bootstrap-key2")),
	}

	// Attempt to bootstrap the node into the network
	err := nodeInstance.BootstrapNode(bootstrapNodes)
	if err != nil {
		log.Fatalf("Failed to bootstrap node: %v", err)
	}

	// Start the node's network server
	go func() {
		err := nodeInstance.Network.StartServer(ip, port)
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
	nodeInstance.LeaveNetwork()
}