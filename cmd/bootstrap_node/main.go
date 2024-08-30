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
	util.SetupLogging("bootstrap_node.log")

	// Create a new node instance specifically for the bootstrap node
	bootstrapNodeInstance := node.NewNode(config, 720 * time.Hour)

	// Start the API server to handle bootstrap requests from other nodes
	go func() {
		err := api.StartServer(config.P2PAddress, bootstrapNodeInstance)
		if err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	// Start listening for incoming messages from other nodes
	go func() {
		err := bootstrapNodeInstance.Network.StartListening()
		if err != nil {
			log.Fatalf("Failed to start node listening: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("Received signal %s, shutting down...\n", sig)

	// Graceful shutdown logic can be added here if needed
}
