package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

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
	util.SetupLogging("bootstrap_node.log")
	
	// Create a new BootstrapNode instance
	bootstrapNodeInstance := node.NewBootstrapNode(config, 86400)
	util.Log().Infof("Node (%s) is starting...", bootstrapNodeInstance.ID)

	// Start the API server to handle bootstrap requests from other nodes
	go func() {
		api.InitRateLimiter(config)
		err := api.StartServer(config.P2PAddress, config.APIAddress, bootstrapNodeInstance)
		if err != nil {
			util.Log().Fatalf("Failed to start API server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	util.Log().Infof("Received signal %s, shutting down...", sig)

	// Gracefully shut down the BootstrapNode
	bootstrapNodeInstance.Shutdown()

	// Confirm shutdown
	util.Log().Info("BootstrapNode shut down successfully.")
}