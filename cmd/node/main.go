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
	util.SetupLogging("node.log")

	// Create a new node instance
	nodeInstance := node.NewNode(config)

	// Bootstrap the node to join the network
	err := nodeInstance.Bootstrap()
	if err != nil {
		log.Fatalf("Failed to bootstrap node: %v", err)
	}

	// Start the network server for the node
	go func() {
		err := nodeInstance.Network.StartServer(config.DHT.P2PAddress, config.DHT.GetP2PIPPort())
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
	nodeInstance.Network.StopServer()
}
