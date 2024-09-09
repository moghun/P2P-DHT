package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func TestBootstrapSuccess(t *testing.T) {
	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	config := &util.Config{
		P2PAddress: fmt.Sprintf("127.0.0.1:%d", port),
		Difficulty: 4,
		BootstrapNodes: []util.BootstrapNode{
			{IP: "127.0.0.1", Port: port + 1}, // Adjust port for uniqueness
		},
	}

	nodeInstance := node.NewNode(config, 24 * time.Hour)
	nodeInstance.Network = &MockNetwork{ShouldFail: false}

	err = nodeInstance.Bootstrap()
	assert.NoError(t, err)
}

func TestBootstrapFailure(t *testing.T) {
	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	config := &util.Config{
		P2PAddress: fmt.Sprintf("127.0.0.1:%d", port),
		Difficulty: 4,
		BootstrapNodes: []util.BootstrapNode{
			{IP: "127.0.0.1", Port: port + 1}, // Adjust port for uniqueness
		},
	}

	nodeInstance := node.NewNode(config, 24 * time.Hour)
	nodeInstance.Network = &MockNetwork{ShouldFail: true}

	err = nodeInstance.Bootstrap()
	assert.Error(t, err)
}