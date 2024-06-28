package tests

import (
	"os"
	"testing"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
address = 127.0.0.1
[dht]
p2p_address = 192.168.0.1:8000
api_address = 192.168.0.1:9000
`
	tmpFile, err := os.CreateTemp("", "config.ini")
	if err != nil {
		t.Fatalf("Failed to create temporary config file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up after the test

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temporary config file: %v", err)
	}
	tmpFile.Close()

	// Load the configuration
	config := util.LoadConfig(tmpFile.Name())

	// Validate the configuration
	if config.Address != "127.0.0.1" {
		t.Errorf("Expected address 127.0.0.1, got %s", config.Address)
	}
	if config.DHT.P2PAddress != "192.168.0.1:8000" {
		t.Errorf("Expected DHT P2P address 192.168.0.1:8000, got %s", config.DHT.P2PAddress)
	}
	if config.DHT.APIAddress != "192.168.0.1:9000" {
		t.Errorf("Expected DHT API address 192.168.0.1:9000, got %s", config.DHT.APIAddress)
	}
}

func TestDHTConfig_GetP2PIPPort(t *testing.T) {
	dhtConfig := util.DHTConfig{P2PAddress: "192.168.0.1:8000"}
	ip, port := dhtConfig.GetP2PIPPort()

	if ip != "192.168.0.1" {
		t.Errorf("Expected IP 192.168.0.1, got %s", ip)
	}
	if port != 8000 {
		t.Errorf("Expected port 8000, got %d", port)
	}
}

func TestDHTConfig_GetAPIIPPort(t *testing.T) {
	dhtConfig := util.DHTConfig{APIAddress: "192.168.0.1:9000"}
	ip, port := dhtConfig.GetAPIIPPort()

	if ip != "192.168.0.1" {
		t.Errorf("Expected IP 192.168.0.1, got %s", ip)
	}
	if port != 9000 {
		t.Errorf("Expected port 9000, got %d", port)
	}
}
