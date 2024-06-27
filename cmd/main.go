package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
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
	util.SetupLogging("kademlia.log")

	// Extract IP and port from the config
	ip, port := config.DHT.GetP2PIPPort()

	// Create a new node
	node := dht.NewNode(ip, port, false)

	// Create a new DHT instance
	dhtInstance := dht.NewDHT(node)

	// Create a new network instance
	network := networking.NewNetwork(dhtInstance)

	// Start the DHT server
	go func() {
		err := network.StartServer(ip, port)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start periodic liveness check
	dhtInstance.StartPeriodicLivenessCheck(10 * time.Second)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("Received signal %s, shutting down...\n", sig)

	// Additional shutdown logic if necessary
}