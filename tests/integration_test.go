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
    dhtInstance := dht.NewDHT()
    _, network, ip, port, err := startNodeWithDynamicPort(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start node: %v", err)
    }

    go func() {
        err := network.StartServer(ip, port)
        if err != nil {
            t.Errorf("Failed to start server: %v", err)
        }
    }()
    time.Sleep(1 * time.Second) // Wait for server to start

    err = network.SendMessage(ip, port, []byte("Hello"))
    if err != nil {
        t.Errorf("Failed to send message through network: %v", err)
    }

    network.StopServer()
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

func NetworkStorageIntegration(t *testing.T) {
    config := util.LoadConfig("../config.ini")
    ip, _ := config.DHT.GetP2PIPPort() // Get IP and ignore static port here

    key := []byte("12345678901234567890123456789012")
    dhtInstance := dht.NewDHT()
    node, network, ip, port, err := startNodeWithDynamicPort(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start node: %v", err)
    }

    go func() {
        err := network.StartServer(ip, port)
        if err != nil {
            t.Fatalf("Server failed to start: %v", err)
        }
    }()
    time.Sleep(1 * time.Second) // Ensure server starts

    actualPort := network.GetListeningPort()
    if actualPort == 0 {
        t.Fatalf("Failed to retrieve the dynamic port")
    }

    err = network.SendMessage(ip, actualPort, []byte("Test message for complete integration"))
    if err != nil {
        t.Errorf("Failed to send message in network: %v", err)
    }

    err = node.Storage.Put("integrationKey", "integrationValue", 3600)
    if err != nil {
        t.Errorf("Failed to store data in DHT: %v", err)
    }

    retrievedValue, err := node.Storage.Get("integrationKey")
    if err != nil || retrievedValue != "integrationValue" {
        t.Errorf("Failed to retrieve the correct value from storage, expected 'integrationValue', got '%s', error: %v", retrievedValue, err)
    }

    network.StopServer()
}

func startNodeWithDynamicPort(dhtInstance *dht.DHT, networkKey []byte) (*dht.Node, *networking.Network, string, int, error) {
    listener, err := net.Listen("tcp", "127.0.0.1:0") // Listen on a free port
    if err != nil {
        return nil, nil, "", 0, err
    }
    defer listener.Close()

    port := listener.Addr().(*net.TCPAddr).Port
    ip := "127.0.0.1"

    node := dht.NewNode(ip, port, true, networkKey)
    if err := node.SetupTLS(); err != nil {
        return nil, nil, "", 0, err
    }
    dhtInstance.JoinNetwork(node)
    network := networking.NewNetwork(dhtInstance)
    go network.StartServer(ip, port)

    return node, network, ip, port, nil
}

func InitializeNode(dhtInstance *dht.DHT, networkKey []byte) (*dht.Node, string, int, error) {
    listener, err := net.Listen("tcp", "127.0.0.1:0") // Listen on a free port
    if err != nil {
        return nil, "", 0, err
    }
    defer listener.Close()

    port := listener.Addr().(*net.TCPAddr).Port
    ip := "127.0.0.1"

    node := dht.NewNode(ip, port, true, networkKey)
    if err := node.SetupTLS(); err != nil {
        return nil, "", 0, err
    }
    dhtInstance.JoinNetwork(node)

    return node, ip, port, nil
}

func TestTwoNodesInteraction(t *testing.T) {
    key := []byte("12345678901234567890123456789012")

    // Initialize a single DHT instance
    dhtInstance := dht.NewDHT()

    // Initialize a single network instance
    network := networking.NewNetwork(dhtInstance)

    // Start the network server once
    go func() {
        if err := network.StartServer("127.0.0.1", 0); err != nil {
            t.Fatalf("Failed to start network server: %v", err)
        }
    }()
    time.Sleep(1 * time.Second) // Ensure server starts

    // Start the first node
    node1, ip1, port1, err := InitializeNode(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start first node: %v", err)
    }
    t.Logf("Node1: %s:%d", ip1, port1)
    time.Sleep(1 * time.Second) // Ensure server starts

    // Verify that Node1 can store and retrieve its own value using DHT methods
    err = dhtInstance.DhtPut("selfKey", "selfValue", 3600)
    if err != nil {
        t.Fatalf("Failed to store self data in node1: %v", err)
    }
    t.Logf("Stored 'selfKey' with value 'selfValue' in Node1")

    selfValue, err := dhtInstance.DhtGet("selfKey")
    if err != nil || selfValue != "selfValue" {
        t.Fatalf("Failed to retrieve self value from node1, expected 'selfValue', got '%s', error: %v", selfValue, err)
    }
    t.Logf("Retrieved 'selfKey' with value '%s' from Node1", selfValue)

    // Start the second node
    node2, ip2, port2, err := InitializeNode(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start second node: %v", err)
    }
    t.Logf("Node2: %s:%d", ip2, port2)
    time.Sleep(1 * time.Second) // Ensure server starts

    // Log all peers in the network for verification
    allPeers := dhtInstance.GetNumNodes()
    t.Logf("All peers in the network: %v", allPeers)

    // Ensure Node2 is aware of Node1
    node2KnownPeers := node2.GetAllPeers()
    t.Logf("Node2 known peers: %v", node2KnownPeers)
    node1Present := false
    for _, peer := range node2KnownPeers {
        if peer.ID == node1.ID {
            node1Present = true
            break
        }
    }
    if !node1Present {
        t.Fatalf("Node2 does not recognize Node1 as a peer")
    }

    // Store a value in node1 using DHT methods
    err = dhtInstance.DhtPut("sharedKey", "sharedValue", 3600)
    if err != nil {
        t.Fatalf("Failed to store data in node1: %v", err)
    }
    t.Logf("Stored 'sharedKey' with value 'sharedValue' in Node1")

    // Verify that Node1 can retrieve the value using DHT methods
    valueInNode1, err := dhtInstance.DhtGet("sharedKey")
    if err != nil || valueInNode1 != "sharedValue" {
        t.Fatalf("Failed to retrieve 'sharedKey' from node1, expected 'sharedValue', got '%s', error: %v", valueInNode1, err)
    }
    t.Logf("Node1 retrieved 'sharedKey' with value '%s'", valueInNode1)

    // Allow some time for the network to sync
    time.Sleep(3 * time.Second)

    // Retrieve the value from node2 using DHT methods
    retrievedValue, err := dhtInstance.DhtGet("sharedKey")
    if err != nil || retrievedValue != "sharedValue" {
        t.Errorf("Failed to retrieve the correct value from node2, expected 'sharedValue', got '%s', error: %v", retrievedValue, err)
    } else {
        t.Logf("Successfully retrieved 'sharedKey' with value '%s' from Node2", retrievedValue)
    }

    // Node2 leaves the network
    err = dhtInstance.LeaveNetwork(node2)
    if err != nil {
        t.Fatalf("Failed to leave network: %v", err)
    }
    t.Logf("Node2 left the network")

    // Allow some time for the network to sync
    time.Sleep(2 * time.Second)

    // Verify node2 has left the network
    peers := dhtInstance.GetNumNodes()
    node2StillPresent := false
    for _, peer := range peers {
        if peer.ID == node2.ID {
            node2StillPresent = true
            break
        }
    }
    if node2StillPresent {
        t.Errorf("Node2 should have left the network but is still present")
    } else {
        t.Logf("Node2 successfully left the network")
    }

    // Clean up
    network.StopServer()
    t.Logf("Network server stopped")
}

func TestNodeFailureAndDataReplication(t *testing.T) {
    key := []byte("12345678901234567890123456789012")

    // Initialize a single DHT instance
    dhtInstance := dht.NewDHT()

    // Initialize a single network instance
    network := networking.NewNetwork(dhtInstance)

    // Start the network server once
    go func() {
        if err := network.StartServer("127.0.0.1", 0); err != nil {
            t.Fatalf("Failed to start network server: %v", err)
        }
    }()
    time.Sleep(1 * time.Second) // Ensure server starts

    // Start three nodes
    _, ip1, port1, err := InitializeNode(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start first node: %v", err)
    }
    t.Logf("Node1: %s:%d", ip1, port1)
    time.Sleep(1 * time.Second) // Ensure server starts

    node2, ip2, port2, err := InitializeNode(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start second node: %v", err)
    }
    t.Logf("Node2: %s:%d", ip2, port2)
    time.Sleep(1 * time.Second) // Ensure server starts

    _, ip3, port3, err := InitializeNode(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start third node: %v", err)
    }
    t.Logf("Node3: %s:%d", ip3, port3)
    time.Sleep(1 * time.Second) // Ensure server starts

    // Store a value in node1 using DHT methods
    err = dhtInstance.DhtPut("replicatedKey", "replicatedValue", 3600)
    if err != nil {
        t.Fatalf("Failed to store data in node1: %v", err)
    }
    t.Logf("Stored 'replicatedKey' with value 'replicatedValue' in Node1")

    // Simulate node2 failure
    err = dhtInstance.LeaveNetwork(node2)
    if err != nil {
        t.Fatalf("Failed to remove node2: %v", err)
    }
    t.Logf("Node2 left the network")

    // Allow some time for the network to sync
    time.Sleep(3 * time.Second)

    // Verify that the value can be retrieved from node3
    retrievedValue, err := dhtInstance.DhtGet("replicatedKey")
    if err != nil || retrievedValue != "replicatedValue" {
        t.Errorf("Failed to retrieve the correct value from node3, expected 'replicatedValue', got '%s', error: %v", retrievedValue, err)
    } else {
        t.Logf("Successfully retrieved 'replicatedKey' with value '%s' from Node3", retrievedValue)
    }

    // Clean up
    network.StopServer()
    t.Logf("Network server stopped")
}

func TestNodeRejoiningNetwork(t *testing.T) {
    key := []byte("12345678901234567890123456789012")

    // Initialize a single DHT instance
    dhtInstance := dht.NewDHT()

    // Initialize a single network instance
    network := networking.NewNetwork(dhtInstance)

    // Start the network server once
    go func() {
        if err := network.StartServer("127.0.0.1", 0); err != nil {
            t.Fatalf("Failed to start network server: %v", err)
        }
    }()
    time.Sleep(1 * time.Second) // Ensure server starts

    // Start two nodes
    _, ip1, port1, err := InitializeNode(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start first node: %v", err)
    }
    t.Logf("Node1: %s:%d", ip1, port1)
    time.Sleep(1 * time.Second) // Ensure server starts

    node2, ip2, port2, err := InitializeNode(dhtInstance, key)
    if err != nil {
        t.Fatalf("Failed to start second node: %v", err)
    }
    t.Logf("Node2: %s:%d", ip2, port2)
    time.Sleep(1 * time.Second) // Ensure server starts

    // Store a value in node1 using DHT methods
    err = dhtInstance.DhtPut("rejoinKey", "rejoinValue", 3600)
    if err != nil {
        t.Fatalf("Failed to store data in node1: %v", err)
    }
    t.Logf("Stored 'rejoinKey' with value 'rejoinValue' in Node1")

    // Simulate node2 leaving the network
    err = dhtInstance.LeaveNetwork(node2)
    if err != nil {
        t.Fatalf("Failed to remove node2: %v", err)
    }
    t.Logf("Node2 left the network")

    // Allow some time for the network to sync
    time.Sleep(2 * time.Second)

    // Simulate node2 rejoining the network
    dhtInstance.JoinNetwork(node2)
    t.Logf("Node2 rejoined the network")

    // Allow some time for the network to sync
    time.Sleep(2 * time.Second)

    // Verify that node2 can retrieve the value
    retrievedValue, err := dhtInstance.DhtGet("rejoinKey")
    if err != nil || retrievedValue != "rejoinValue" {
        t.Errorf("Failed to retrieve the correct value from node2, expected 'rejoinValue', got '%s', error: %v", retrievedValue, err)
    } else {
        t.Logf("Successfully retrieved 'rejoinKey' with value '%s' from Node2", retrievedValue)
    }

    // Clean up
    network.StopServer()
    t.Logf("Network server stopped")
}