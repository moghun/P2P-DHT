package util

import (
	"fmt"
	"log"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	Address string
	DHT     DHTConfig
}

type DHTConfig struct {
	P2PAddress string `ini:"p2p_address"`
	APIAddress string `ini:"api_address"`
}

func LoadConfig(filename string) *Config {
	cfg, err := ini.Load(filename)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	return &Config{
		Address: cfg.Section("").Key("address").String(),
		DHT: DHTConfig{
			P2PAddress: cfg.Section("dht").Key("p2p_address").String(),
			APIAddress: cfg.Section("dht").Key("api_address").String(),
		},
	}
}

func (d *DHTConfig) GetP2PIPPort() (string, int) {
	parts := strings.Split(d.P2PAddress, ":")
	ip := parts[0]
	port := 0
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &port)
	}
	return ip, port
}

func (d *DHTConfig) GetAPIIPPort() (string, int) {
	parts := strings.Split(d.APIAddress, ":")
	ip := parts[0]
	port := 0
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &port)
	}
	return ip, port
}
