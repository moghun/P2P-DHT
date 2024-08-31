package tests

import (
	"testing"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"github.com/stretchr/testify/assert"
)

func TestGenerateNodeIDWithPoW(t *testing.T) {
	ip := "127.0.0.1"
	port := 8080

	// Generate a node ID with PoW
	nodeID, nonce := security.GenerateNodeIDWithPoW(ip, port)

	// Assert that the nodeID is not empty
	assert.NotEmpty(t, nodeID, "Generated node ID should not be empty")

	// Validate the generated node ID
	isValid := security.ValidateNodeIDWithPoW(ip, port, nodeID, nonce)
	assert.True(t, isValid, "Generated node ID should be valid with the correct nonce")
}

func TestValidateNodeIDWithPoW(t *testing.T) {
	ip := "127.0.0.1"
	port := 8080

	// Generate a node ID with PoW
	nodeID, nonce := security.GenerateNodeIDWithPoW(ip, port)

	// Validate the node ID with the correct nonce
	isValid := security.ValidateNodeIDWithPoW(ip, port, nodeID, nonce)
	assert.True(t, isValid, "Node ID should be valid with the correct nonce")

	// Validate the node ID with an incorrect nonce
	isValid = security.ValidateNodeIDWithPoW(ip, port, nodeID, nonce+1)
	assert.False(t, isValid, "Node ID should be invalid with an incorrect nonce")

	// Validate a different node ID with the correct nonce
	differentNodeID, _ := security.GenerateNodeIDWithPoW(ip, port+1)
	isValid = security.ValidateNodeIDWithPoW(ip, port, differentNodeID, nonce)
	assert.False(t, isValid, "A different node ID should be invalid even with the correct nonce")
}

func TestGenerateNodeIDWithPoWDifficulty(t *testing.T) {
	ip := "127.0.0.1"
	port := 8080

	// Generate a node ID with PoW
	nodeID, _ := security.GenerateNodeIDWithPoW(ip, port)

	// Check if the nodeID meets the difficulty requirements (e.g., leading zeros)
	difficulty := 4
	prefix := "0000"
	assert.True(t, len(nodeID) > difficulty && nodeID[:difficulty] == prefix, "Node ID should start with the required number of leading zeros")
}
