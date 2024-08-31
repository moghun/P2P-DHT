package security

import (
	"sync"
	"time"
)

type RateLimiter struct {
	rate      int           // requests per second
	burst     int           // maximum burst size
	tokens    int           // current number of tokens
	lastToken time.Time     // last time the bucket was refilled
	mu        sync.Mutex    // mutex to protect shared state
}

func NewRateLimiter(rate int, burst int) *RateLimiter {
	return &RateLimiter{
		rate:      rate,
		burst:     burst,
		tokens:    burst,
		lastToken: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastToken)
	refillTokens := int(elapsed.Seconds()) * rl.rate

	if refillTokens > 0 {
		rl.tokens = min(rl.burst, rl.tokens+refillTokens)
		rl.lastToken = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
