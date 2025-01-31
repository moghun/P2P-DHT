package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

const relativeConfigPath = "config.ini"

func main() {
	// Parse command-line parameters for config path
	configPath := flag.String("c", "", "path to configuration file")
	flag.Parse()

	// If config path is not provided, fallback to relative config path
	if *configPath == "" {
		if _, err := os.Stat(relativeConfigPath); err == nil {
			*configPath = relativeConfigPath
		} else {
			util.Log().Fatal("No valid config file found. Please provide a valid config path.")
		}
	}

	// Load configuration
	config := util.LoadConfig(*configPath)

	// Set up logging
	util.SetupLogging("node.log")


	// Create a new node instance
	nodeInstance := node.NewNode(config, time.Duration(config.TTL))

	util.Log().Infof("Node (%s) is starting...", nodeInstance.ID)
	// Bootstrap the node to join the network
	err := nodeInstance.Bootstrap()
	if err != nil {
		util.Log().Fatalf("Failed to bootstrap node: %v", err)
	}

	// Start the API server for the node
	go func() {
		api.InitRateLimiter(config)
		err := api.StartServer(config.P2PAddress, config.APIAddress, nodeInstance)
		if err != nil {
			util.Log().Fatalf("Failed to start API server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigChan
	util.Log().Infof("Received signal %s, shutting down...", sig)

	// Gracefully shut down the node
	nodeInstance.Shutdown()

	// Confirm shutdown
	util.Log().Info("Node shut down successfully.")
}