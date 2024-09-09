package util

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	Address          string
	P2PAddress       string
	APIAddress       string
	BootstrapNodes   []BootstrapNode
	EncryptionKey    []byte
	TTL              int // TTL in seconds
	RateLimiterRate  int // Requests per second
	RateLimiterBurst int // Burst size
	Difficulty       int // PoW Difficulty
}

type BootstrapNode struct {
	IP   string
	Port int
}

func LoadConfig(filename string) *Config {
	cfg, err := ini.Load(filename)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	p2pAddress := cfg.Section("node").Key("p2p_address").String()
	apiAddress := cfg.Section("node").Key("api_address").String()
	encryptionKey := []byte(cfg.Section("security").Key("encryption_key").String())
	ttl, _ := cfg.Section("node").Key("ttl").Int()

	// Load rate limiter settings from config
	rateLimiterRate, _ := cfg.Section("rate_limiter").Key("requests_per_second").Int()
	rateLimiterBurst, _ := cfg.Section("rate_limiter").Key("burst_size").Int()

	// Load PoW difficulty from config
	difficulty, _ := cfg.Section("security").Key("difficulty").Int()

	bootstrapNodes := LoadBootstrapNodes(cfg)

	return &Config{
		Address:          cfg.Section("").Key("address").String(),
		P2PAddress:       p2pAddress,
		APIAddress:       apiAddress,
		BootstrapNodes:   bootstrapNodes,
		EncryptionKey:    encryptionKey,
		TTL:              ttl,
		RateLimiterRate:  rateLimiterRate,
		RateLimiterBurst: rateLimiterBurst,
		Difficulty:       difficulty,
	}
}

// loadBootstrapNodes parses the bootstrap nodes from the config file.
func LoadBootstrapNodes(cfg *ini.File) []BootstrapNode {
	var bootstrapNodes []BootstrapNode
	for _, key := range cfg.Section("bootstrap").KeyStrings() {
		nodeAddr := cfg.Section("bootstrap").Key(key).String()
		ip, port, _ := ParseAddress(nodeAddr)
		bootstrapNodes = append(bootstrapNodes, BootstrapNode{IP: ip, Port: port})
	}
	return bootstrapNodes
}

func ParseAddress(address string) (string, int, error) {
	parts := strings.Split(address, ":")
	if len(parts) < 2 {
		return "", 0, fmt.Errorf("invalid address format in config")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port number in address: %v", err)
	}
	return parts[0], port, nil
}
