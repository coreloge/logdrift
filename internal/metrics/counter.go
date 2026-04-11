// Package metrics provides lightweight in-process counters for tracking
// log entry rates and drift events across services.
package metrics

import (
	"sync"
	"time"
)

// Counter tracks per-service log entry counts and drift event totals.
type Counter struct {
	mu       sync.RWMutex
	entries  map[string]int64
	drifts   map[string]int64
	started  time.Time
}

// New returns an initialised Counter.
func New() *Counter {
	return &Counter{
		entries: make(map[string]int64),
		drifts:  make(map[string]int64),
		started: time.Now(),
	}
}

// RecordEntry increments the entry count for the given service.
func (c *Counter) RecordEntry(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[service]++
}

// RecordDrift increments the drift count for the given service.
func (c *Counter) RecordDrift(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.drifts[service]++
}

// Entries returns a snapshot of entry counts keyed by service name.
func (c *Counter) Entries() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]int64, len(c.entries))
	for k, v := range c.entries {
		out[k] = v
	}
	return out
}

// Drifts returns a snapshot of drift counts keyed by service name.
func (c *Counter) Drifts() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]int64, len(c.drifts))
	for k, v := range c.drifts {
		out[k] = v
	}
	return out
}

// Uptime returns the duration since the counter was created.
func (c *Counter) Uptime() time.Duration {
	return time.Since(c.started)
}

// Reset zeroes all counters without resetting the start time.
func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]int64)
	c.drifts = make(map[string]int64)
}
