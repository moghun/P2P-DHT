package tests

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func TestRateLimiterIntegration(t *testing.T) {
	// Dynamic port allocation
	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	// Create a real node instance with necessary configuration
    config := &util.Config{
        P2PAddress:    fmt.Sprintf("127.0.0.1:%d", port),
        EncryptionKey: []byte("1234567890123456"),
        RateLimiterRate:  10,
		RateLimiterBurst: 20,
		Difficulty: 4,
    }
    api.InitRateLimiter(config)

	realNode := node.NewNode(config, 86400)

	go func() {
		err := api.StartServer(fmt.Sprintf("127.0.0.1:%d", port), realNode)
		assert.NoError(t, err, "Failed to start API server")
	}()

	// Give the server time to start
	time.Sleep(1 * time.Second)

	// Test exceeding the rate limit
	for i := 0; i < 25; i++ { // Attempt more than the burst + rate limit
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			t.Logf("Connection attempt %d failed: %v", i+1, err)
		} else {
			if i < 20 { // First 20 requests should be allowed
				assert.NoError(t, err, "Expected connection to be allowed")
				t.Logf("Connection attempt %d allowed", i+1)
				conn.Close()
			} else { // Next 5 requests should be rate-limited
				assert.NoError(t, err, "Expected connection to be denied due to rate limiting")
				t.Logf("Connection attempt %d unexpectedly allowed", i+1)
				conn.Close()
			}
		}
	}

	// Wait for 1 second to let the rate limiter refill
	time.Sleep(1 * time.Second)

	// Now another connection should be allowed
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	assert.NoError(t, err, "Expected connection to be allowed after rate limiter refill")
	conn.Close()
}