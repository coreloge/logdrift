// Package coalesce merges log entries from multiple services into a single
// unified entry when they share a common correlation ID within a time window.
package coalesce

import (
	"sync"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

// DefaultOptions returns a sensible default configuration.
func DefaultOptions() Options {
	return Options{
		CorrelationField: "request_id",
		Window:           2 * time.Second,
		MinSources:       2,
	}
}

// Options controls coalescer behaviour.
type Options struct {
	// CorrelationField is the entry field used to group related log lines.
	CorrelationField string
	// Window is the maximum duration to wait before flushing an incomplete group.
	Window time.Duration
	// MinSources is the minimum number of distinct services required before
	// emitting a coalesced entry.
	MinSources int
}

// group holds buffered entries sharing the same correlation ID.
type group struct {
	entries  []diff.Entry
	services map[string]struct{}
	createdAt time.Time
}

// Coalescer buffers log entries by a correlation field and emits merged groups.
type Coalescer struct {
	opts   Options
	mu     sync.Mutex
	groups map[string]*group
}

// New creates a Coalescer with the provided options.
func New(opts Options) *Coalescer {
	return &Coalescer{
		opts:   opts,
		groups: make(map[string]*group),
	}
}

// Record adds an entry to the appropriate group. It returns a merged slice of
// entries if the group is ready to emit (MinSources met), or nil otherwise.
func (c *Coalescer) Record(e diff.Entry) []diff.Entry {
	cid, ok := e.Fields[c.opts.CorrelationField]
	if !ok || cid == "" {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	g, exists := c.groups[cid]
	if !exists {
		g = &group{
			services:  make(map[string]struct{}),
			createdAt: time.Now(),
		}
		c.groups[cid] = g
	}

	g.entries = append(g.entries, e)
	g.services[e.Service] = struct{}{}

	if len(g.services) >= c.opts.MinSources {
		result := g.entries
		delete(c.groups, cid)
		return result
	}
	return nil
}

// Flush returns and removes all groups whose window has expired, regardless of
// whether MinSources was reached.
func (c *Coalescer) Flush() [][]diff.Entry {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expired [][]diff.Entry
	now := time.Now()
	for cid, g := range c.groups {
		if now.Sub(g.createdAt) >= c.opts.Window {
			expired = append(expired, g.entries)
			delete(c.groups, cid)
		}
	}
	return expired
}
