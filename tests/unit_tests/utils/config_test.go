package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gopkg.in/ini.v1"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	configContent := `
		[node]
		p2p_address = 127.0.0.1:8000
		api_address = 127.0.0.1:9000
		ttl = 1000000
		[security]
		encryption_key = 1234567890abcdef
		[bootstrap]
		node1 = 192.168.1.1:8080
		node2 = 192.168.1.2:8081
		`
	tmpFile, err := os.CreateTemp("", "config_test.ini")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(configContent))
	assert.NoError(t, err)
	tmpFile.Close()

	// Load the config from the temporary file
	config := util.LoadConfig(tmpFile.Name())

	// Assert the loaded config values
	assert.Equal(t, "127.0.0.1:8000", config.P2PAddress)
	assert.Equal(t, "127.0.0.1:9000", config.APIAddress)
	assert.Equal(t, []byte("1234567890abcdef"), config.EncryptionKey)
	assert.Equal(t, 1000000, config.TTL)
	assert.Len(t, config.BootstrapNodes, 2)
	assert.Equal(t, "192.168.1.1", config.BootstrapNodes[0].IP)
	assert.Equal(t, 8080, config.BootstrapNodes[0].Port)
	assert.Equal(t, "192.168.1.2", config.BootstrapNodes[1].IP)
	assert.Equal(t, 8081, config.BootstrapNodes[1].Port)
}

func TestLoadBootstrapNodes(t *testing.T) {
	configContent := `
[bootstrap]
node1 = 192.168.1.1:8080
node2 = 192.168.1.2:8081
`
	cfg, err := ini.Load([]byte(configContent))
	assert.NoError(t, err)

	bootstrapNodes := util.LoadBootstrapNodes(cfg)

	assert.Len(t, bootstrapNodes, 2)
	assert.Equal(t, "192.168.1.1", bootstrapNodes[0].IP)
	assert.Equal(t, 8080, bootstrapNodes[0].Port)
	assert.Equal(t, "192.168.1.2", bootstrapNodes[1].IP)
	assert.Equal(t, 8081, bootstrapNodes[1].Port)
}

func TestParseAddress(t *testing.T) {
	ip, port, err := util.ParseAddress("192.168.1.1:8080")
	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.1", ip)
	assert.Equal(t, 8080, port)

	ip, port, err = util.ParseAddress("127.0.0.1:9000")
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", ip)
	assert.Equal(t, 9000, port)

	// Test invalid address format
	_, _, err = util.ParseAddress("invalid_address")
	assert.Error(t, err)
}
