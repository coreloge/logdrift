// Package limiter provides a token-bucket rate limiter that caps the number
// of log entries forwarded per second across the entire stream.
package limiter

import (
	"context"
	"sync"
	"time"

	"github.com/logdrift/internal/diff"
)

// Options configures the global rate limiter.
type Options struct {
	// Rate is the maximum number of entries allowed per second.
	// A value of 0 disables limiting.
	Rate int
	// BurstSize is the maximum number of tokens that can accumulate.
	// Defaults to Rate when zero.
	BurstSize int
}

// DefaultOptions returns an Options that imposes no limit.
func DefaultOptions() Options {
	return Options{Rate: 0, BurstSize: 0}
}

// Limiter is a token-bucket rate limiter for log entries.
type Limiter struct {
	mu        sync.Mutex
	rate      int
	burst     int
	tokens    float64
	lastCheck time.Time
}

// New creates a Limiter from opts. If opts.Rate is zero, Allow always
// returns true.
func New(opts Options) *Limiter {
	burst := opts.BurstSize
	if burst == 0 {
		burst = opts.Rate
	}
	return &Limiter{
		rate:      opts.Rate,
		burst:     burst,
		tokens:    float64(burst),
		lastCheck: time.Now(),
	}
}

// Allow reports whether the entry should be forwarded. It refills tokens
// based on elapsed time and consumes one token per accepted entry.
func (l *Limiter) Allow(_ diff.Entry) bool {
	if l.rate == 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(l.lastCheck).Seconds()
	l.lastCheck = now
	l.tokens += elapsed * float64(l.rate)
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}
	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

// Stream wraps an input channel and forwards only entries that pass the
// rate limit. Dropped entries are silently discarded.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) <-chan diff.Entry {
	out := make(chan diff.Entry, 64)
	l := New(opts)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-in:
				if !ok {
					return
				}
				if l.Allow(e) {
					select {
					case out <- e:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
