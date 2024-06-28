package tests

import (
	"net"
	"testing"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/networking"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func TestNodeInitializationAndConfig(t *testing.T) {
    config := util.LoadConfig("../config.ini")

    ip, port := config.DHT.GetP2PIPPort()
    node := dht.NewNode(ip, port, true, []byte("testkey1234567890"))
    if node == nil {
        t.Errorf("Failed to initialize node with config")
    }
}

func TestNetworkMessagePassing(t *testing.T) {
    key := []byte("12345678901234567890123456789012")
    node := dht.NewNode("127.0.0.1", 8000, true, key)
    dhtInstance := dht.NewDHT(node)
    network := networking.NewNetwork(dhtInstance)

    go func() {
        err := network.StartServer("127.0.0.1", 8000)
        if err != nil {
            t.Errorf("Failed to start server: %v", err)
        }
    }()
    time.Sleep(1 * time.Second) // Wait for server to start

    err := network.SendMessage("127.0.0.1", 8000, []byte("Hello"))
    if err != nil {
        t.Errorf("Failed to send message through network: %v", err)
    }
}

func TestDHTWithStorage(t *testing.T) {
    key := []byte("12345678901234567890123456789012")
    storageInstance := storage.NewStorage(24*time.Hour, key)
    node := dht.NewNode("127.0.0.1", 8000, true, key)
    node.Storage = storageInstance


    err := node.Put("testKey", "testValue", int(10*time.Minute.Seconds()))
    if err != nil {
        t.Fatalf("Failed to put value into storage: %v", err)
    }

    value, err := node.Get("testKey")
    if err != nil || value != "testValue" {
        t.Errorf("Expected 'testValue', got '%s', error: %v", value, err)
    }
}

func startNodeWithDynamicPort(networkKey []byte) (*dht.Node, *networking.Network, string, int, error) {
    listener, err := net.Listen("tcp", "127.0.0.1:8000") // Listen on a free port
    if err != nil {
        return nil, nil, "", 0, err
    }

    port := listener.Addr().(*net.TCPAddr).Port
    ip := "127.0.0.1"

    node := dht.NewNode(ip, port, true, networkKey)
	dhtInstance := dht.NewDHT(node)
    network := networking.NewNetwork(dhtInstance)
    go network.StartServer(ip, port)

    return node, network, ip, port, nil
}

func TestDynamicPortAssignment(t *testing.T) {
    key := []byte("12345678901234567890123456789012")
    _, network, ip, port, err := startNodeWithDynamicPort(key)
    if err != nil {
        t.Fatalf("Failed to start node: %v", err)
    }

    // Use the dynamically assigned address
    err = network.SendMessage(ip, port, []byte("Hello"))
    if err != nil {
        t.Errorf("Failed to send message to dynamically started server: %v", err)
    }
}


func TestCompleteIntegration(t *testing.T) {
	config := util.LoadConfig("../config.ini")
	ip, _ := config.DHT.GetP2PIPPort() // Get IP and ignore static port here

	key := []byte("12345678901234567890123456789012")
	node := dht.NewNode(ip, 0, true, key)  // Initialize node with a dynamic port
	dhtInstance := dht.NewDHT(node)
	network := networking.NewNetwork(dhtInstance)

	// Start the server on a dynamic port
	go func() {
		err := network.StartServer(ip, 0)
		if err != nil {
			t.Fatalf("Server failed to start: %v", err)
		}
	}()
	time.Sleep(1 * time.Second) // Ensure server starts

	// Retrieve the actual dynamic port from the network
	actualPort := network.GetListeningPort()
	if actualPort == 0 {
		t.Fatalf("Failed to retrieve the dynamic port")
	}

	// Test sending a network message
	err := network.SendMessage(ip, actualPort, []byte("Test message for complete integration"))
	if err != nil {
		t.Errorf("Failed to send message in network: %v", err)
	}

	// Perform storage operations
	err = node.Storage.Put("integrationKey", "integrationValue", 3600)
	if err != nil {
		t.Errorf("Failed to store data in DHT: %v", err)
	}

	retrievedValue, err := node.Storage.Get("integrationKey")
	if err != nil || retrievedValue != "integrationValue" {
		t.Errorf("Failed to retrieve the correct value from storage, expected 'integrationValue', got '%s', error: %v", retrievedValue, err)
	}

	// Stop the server
	network.StopServer()
}
