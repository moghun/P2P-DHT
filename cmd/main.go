package main

import (
	"flag"
	"log"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/networking"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func main() {
    // Parse command-line parameters
    configPath := flag.String("c", "config.ini", "path to configuration file")
    flag.Parse()

    // Initialize logging
    util.InitLogging()

    // Load configuration
    config := util.LoadConfig(*configPath)
    if config == nil {
        log.Fatal("Failed to load configuration")
    }

    // Initialize storage
    store := storage.NewStorage()

    // Initialize DHT with a placeholder for network (will be set later)
    dht := dht.NewDHT(nil, store)
    if dht == nil {
        log.Fatal("Failed to initialize DHT node")
    }

    // Initialize network with a message handler function
    network := networking.NewNetwork(config, dht.HandleMessage)
    if network == nil {
        log.Fatal("Failed to initialize network")
    }

    // Set the network in the DHT
    dht.SetNetwork(network)

    // Start the DHT node
    go dht.Run()

    // Keep the main function running
    select {}
}
