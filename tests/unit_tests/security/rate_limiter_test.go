package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create a rate limiter that allows 5 requests per second with a burst of 10
	rl := security.NewRateLimiter(5, 10)

	// Initially, allow the maximum burst size of requests
	for i := 0; i < 10; i++ {
		assert.True(t, rl.Allow(), "Expected request to be allowed within burst size")
	}

	// After burst limit is reached, requests should be denied
	assert.False(t, rl.Allow(), "E.xpected request to be denied after exceeding burst size")

	// Wait for 1 second to allow for rate limiter refill
	time.Sleep(1 * time.Second)

	// Now the limiter should allow more requests, according to the rate
	assert.True(t, rl.Allow(), "Expected request to be allowed after refill")
}

func TestRateLimiter_Refill(t *testing.T) {
	// Create a rate limiter that allows 1 request per second with a burst of 1
	rl := security.NewRateLimiter(1, 1)

	// Initially, the first request should be allowed
	assert.True(t, rl.Allow(), "Expected first request to be allowed")

	// Subsequent request should be denied as we exceeded the burst limit
	assert.False(t, rl.Allow(), "Expected second request to be denied due to rate limiting")

	// Wait for 1 second to allow for rate limiter refill
	time.Sleep(1 * time.Second)

	// Now the limiter should allow one more request
	assert.True(t, rl.Allow(), "Expected request to be allowed after 1 second")
}

func TestRateLimiter_Burst(t *testing.T) {
	// Create a rate limiter that allows 2 requests per second with a burst of 3
	rl := security.NewRateLimiter(2, 3)

	// Initially, allow the maximum burst size of requests
	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow(), "Expected request to be allowed within burst size")
	}

	// After burst limit is reached, requests should be denied
	assert.False(t, rl.Allow(), "Expected request to be denied after exceeding burst size")

	// Wait for 0.5 seconds and check that we still can't send a request
	time.Sleep(500 * time.Millisecond)
	assert.False(t, rl.Allow(), "Expected request to be denied due to insufficient refill time")

	// Wait for another 0.5 seconds and check that one request is allowed
	time.Sleep(500 * time.Millisecond)
	assert.True(t, rl.Allow(), "Expected request to be allowed after sufficient refill time")
}