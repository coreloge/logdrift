// Package ratelimit provides a token-bucket style rate limiter for
// controlling how many log entries are forwarded per second per service.
package ratelimit

import (
	"sync"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the throttle behaviour.
type Options struct {
	// MaxPerSecond is the maximum number of entries allowed per service per second.
	// A value of 0 disables rate limiting.
	MaxPerSecond int
}

// DefaultOptions returns a permissive default configuration.
func DefaultOptions() Options {
	return Options{MaxPerSecond: 0}
}

// Throttle wraps an entry channel and forwards entries that pass the
// per-service token-bucket limit.
type Throttle struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// New creates a Throttle with the given options.
func New(opts Options) *Throttle {
	return &Throttle{
		opts:    opts,
		buckets: make(map[string]*bucket),
	}
}

// Allow returns true if the entry should be forwarded based on the
// current token count for its service.
func (t *Throttle) Allow(entry diff.Entry) bool {
	if t.opts.MaxPerSecond <= 0 {
		return true
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	b, ok := t.buckets[entry.Service]
	if !ok {
		b = &bucket{tokens: t.opts.MaxPerSecond, lastReset: now}
		t.buckets[entry.Service] = b
	}

	if now.Sub(b.lastReset) >= time.Second {
		b.tokens = t.opts.MaxPerSecond
		b.lastReset = now
	}

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// Apply reads entries from in and writes allowed entries to the returned
// channel. It closes the output channel when in is closed.
func (t *Throttle) Apply(in <-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry, 64)
	go func() {
		defer close(out)
		for entry := range in {
			if t.Allow(entry) {
				out <- entry
			}
		}
	}()
	return out
}
