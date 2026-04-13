// Package correlate groups log entries across services by a shared correlation ID field.
// It maintains a time-bounded index so that entries arriving within a configurable
// window can be retrieved together for cross-service analysis.
package correlate

import (
	"sync"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Field:  "correlation_id",
		Window: 30 * time.Second,
		MaxIDs: 1000,
	}
}

// Options controls the behaviour of a Correlator.
type Options struct {
	// Field is the log-entry field whose value is used as the correlation key.
	Field string
	// Window is the maximum age of a group before it is evicted.
	Window time.Duration
	// MaxIDs caps the number of tracked correlation IDs at any one time.
	MaxIDs int
}

type group struct {
	entries  []diff.Entry
	updated  time.Time
}

// Correlator indexes log entries by a shared field value.
type Correlator struct {
	opts   Options
	mu     sync.Mutex
	groups map[string]*group
}

// New creates a Correlator with the provided options.
func New(opts Options) *Correlator {
	if opts.Field == "" {
		opts.Field = DefaultOptions().Field
	}
	if opts.Window <= 0 {
		opts.Window = DefaultOptions().Window
	}
	if opts.MaxIDs <= 0 {
		opts.MaxIDs = DefaultOptions().MaxIDs
	}
	return &Correlator{
		opts:   opts,
		groups: make(map[string]*group),
	}
}

// Record adds an entry to the group identified by its correlation field value.
// Entries whose field is absent are silently ignored.
// Eviction of stale groups is performed on every call.
func (c *Correlator) Record(e diff.Entry) {
	id, ok := e.Fields[c.opts.Field]
	if !ok || id == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evict()
	g, exists := c.groups[id]
	if !exists {
		if len(c.groups) >= c.opts.MaxIDs {
			return // drop when at capacity
		}
		g = &group{}
		c.groups[id] = g
	}
	g.entries = append(g.entries, e)
	g.updated = time.Now()
}

// Get returns all entries recorded under the given correlation ID.
// A nil slice is returned when the ID is unknown or has been evicted.
func (c *Correlator) Get(id string) []diff.Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evict()
	if g, ok := c.groups[id]; ok {
		out := make([]diff.Entry, len(g.entries))
		copy(out, g.entries)
		return out
	}
	return nil
}

// evict removes groups whose last update is older than the configured window.
// Caller must hold c.mu.
func (c *Correlator) evict() {
	cutoff := time.Now().Add(-c.opts.Window)
	for id, g := range c.groups {
		if g.updated.Before(cutoff) {
			delete(c.groups, id)
		}
	}
}
