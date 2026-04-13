// Package throttle provides per-level rate limiting for log entry streams.
// Entries exceeding the configured rate for a given log level are dropped.
package throttle

import (
	"sync"
	"time"
)

// bucket tracks the token count for a single log level.
type bucket struct {
	mu       sync.Mutex
	tokens   float64
	last     time.Time
	rate     float64 // tokens per second
	capacity float64
}

func newBucket(rate float64) *bucket {
	return &bucket{
		tokens:   rate,
		last:     time.Now(),
		rate:     rate,
		capacity: rate,
	}
}

// allow returns true if a token is available and consumes it.
func (b *bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.last = now

	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
