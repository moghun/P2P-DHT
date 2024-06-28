package util

import (
	"log"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	Address string
	DHT     DHTConfig
}

type DHTConfig struct {
	P2PAddress string `ini:"p2p address"`
	APIAddress string `ini:"api address"`
}

// LoadConfig reads configuration from the specified file and returns a Config object.
func LoadConfig(filename string) *Config {
	cfg, err := ini.Load(filename)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	p2pAddress := cfg.Section("dht").Key("p2p address").String()
	apiAddress := cfg.Section("dht").Key("api address").String()
	log.Printf("Loaded P2P Address: %s", p2pAddress)
	log.Printf("Loaded API Address: %s", apiAddress)

	return &Config{
		Address: cfg.Section("").Key("address").String(),
		DHT: DHTConfig{
			P2PAddress: p2pAddress,
			APIAddress: apiAddress,
		},
	}
}

// GetP2PIPPort parses and returns the IP and port from the P2P address.
func (d *DHTConfig) GetP2PIPPort() (string, int) {
	parts := strings.Split(d.P2PAddress, ":")
	if len(parts) < 2 {
		log.Fatalf("Invalid P2P address format in config")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("Invalid port number in P2P address: %v", err)
	}
	log.Printf("Parsed P2P IP: %s, Port: %d", parts[0], port)
	return parts[0], port
}

// GetAPIIPPort parses and returns the IP and port from the API address.
func (d *DHTConfig) GetAPIIPPort() (string, int) {
	parts := strings.Split(d.APIAddress, ":")
	if len(parts) < 2 {
		log.Fatalf("Invalid API address format in config")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("Invalid port number in API address: %v", err)
	}
	log.Printf("Parsed API IP: %s, Port: %d", parts[0], port)
	return parts[0], port
}