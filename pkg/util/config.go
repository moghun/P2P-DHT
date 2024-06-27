package util

import (
    "gopkg.in/ini.v1"
    "log"
)

type Config struct {
    Address    string
    P2PAddress string
    APIAddress string
}

func LoadConfig(filename string) *Config {
    cfg, err := ini.Load(filename)
    if err != nil {
        log.Printf("Failed to load config file: %v", err)
        return nil
    }

    return &Config{
        Address:    cfg.Section("").Key("address").String(),
        P2PAddress: cfg.Section("dht").Key("p2p address").String(),
        APIAddress: cfg.Section("dht").Key("api address").String(),
    }
}
