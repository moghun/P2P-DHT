package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
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
	util.SetupLogging("node.log")

	// Create a new node instance
	nodeInstance := node.NewNode(config, time.Duration(config.TTL))

	// Bootstrap the node to join the network
	err := nodeInstance.Bootstrap()
	if err != nil {
		log.Fatalf("Failed to bootstrap node: %v", err)
	}

	// Start the API server for the node
	go func() {
		err := api.StartServer(config.P2PAddress, nodeInstance)
		if err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	// Start listening for incoming messages from other nodes
	go func() {
		err := nodeInstance.Network.StartListening()
		if err != nil {
			log.Fatalf("Failed to start node listening: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigChan
	fmt.Printf("Received signal %s, shutting down...\n", sig)

	// Gracefully shut down the node
	nodeInstance.Shutdown()

	// Confirm shutdown
	fmt.Println("Node shut down successfully.")
}