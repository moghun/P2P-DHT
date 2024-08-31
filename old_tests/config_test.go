package tests

import (
	"testing"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)
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
