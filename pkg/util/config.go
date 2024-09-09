package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	BootstrapRetryInterval time.Duration // Retry interval for bootstrap in seconds
	MaxBootstrapRetries    int           // Max number of bootstrap retries
}

type BootstrapNode struct {
	IP   string
	Port int
}

func LoadConfig(filename string) *Config {
	cfg, err := ini.Load(filename)
	if err != nil {
		Log().Fatalf("Failed to read config file: %v", err)
	}

	p2pAddress := cfg.Section("node").Key("p2p_address").String()
	apiAddress := cfg.Section("node").Key("api_address").String()
	encryptionKey := []byte(cfg.Section("security").Key("encryption_key").String())
	ttl, _ := cfg.Section("node").Key("cleanup_interval").Int()

	bootstrapRetryInterval, _ := cfg.Section("node").Key("bootstrap_retry_interval").Int()
	maxBootstrapRetries, _ := cfg.Section("node").Key("max_bootstrap_retries").Int()

	// Load bootstrap retry settings
	rateLimiterRate, _ := cfg.Section("rate_limiter").Key("requests_per_second").Int()
	rateLimiterBurst, _ := cfg.Section("rate_limiter").Key("burst_size").Int()

	// Convert interval to time.Duration
	retryInterval := time.Duration(bootstrapRetryInterval) * time.Second

	bootstrapNodes := LoadBootstrapNodes(cfg)

	return &Config{
		P2PAddress:            p2pAddress,
		APIAddress:			   apiAddress,
		EncryptionKey:         encryptionKey,
		TTL:                   ttl,
		Difficulty:            cfg.Section("security").Key("difficulty").MustInt(4),
		RateLimiterRate:	rateLimiterRate,
		RateLimiterBurst:	rateLimiterBurst,
		BootstrapRetryInterval: retryInterval,
		MaxBootstrapRetries:    maxBootstrapRetries,
		BootstrapNodes:        bootstrapNodes,
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
