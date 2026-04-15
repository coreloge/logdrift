// Package ceiling provides a pipeline stage that enforces a maximum
// number of entries emitted per service within a rolling time window.
package ceiling

import (
	"context"
	"fmt"
	"sync"
	"time"

	"logdrift/internal/diff"
)

// DefaultOptions returns an Options with no ceiling enforced.
func DefaultOptions() Options {
	return Options{
		Window:     time.Minute,
		MaxPerWindow: 0,
	}
}

// Options configures the Ceiling.
type Options struct {
	// Window is the rolling duration over which entries are counted.
	Window time.Duration
	// MaxPerWindow is the maximum number of entries allowed per service
	// within the window. Zero means unlimited.
	MaxPerWindow int
}

// bucket tracks entry timestamps for a single service.
type bucket struct {
	mu   sync.Mutex
	times []time.Time
}

// Ceiling enforces a per-service entry rate ceiling.
type Ceiling struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*bucket
}

// New creates a new Ceiling. Returns an error if the window is zero or negative.
func New(opts Options) (*Ceiling, error) {
	if opts.Window <= 0 {
		return nil, fmt.Errorf("ceiling: window must be positive, got %v", opts.Window)
	}
	return &Ceiling{
		opts:    opts,
		buckets: make(map[string]*bucket),
	}, nil
}

// Allow returns true if the entry should be forwarded, false if it exceeds the ceiling.
func (c *Ceiling) Allow(entry diff.Entry) bool {
	if c.opts.MaxPerWindow <= 0 {
		return true
	}
	c.mu.Lock()
	b, ok := c.buckets[entry.Service]
	if !ok {
		b = &bucket{}
		c.buckets[entry.Service] = b
	}
	c.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-c.opts.Window)

	b.mu.Lock()
	defer b.mu.Unlock()

	// evict expired timestamps
	valid := b.times[:0]
	for _, t := range b.times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	b.times = valid

	if len(b.times) >= c.opts.MaxPerWindow {
		return false
	}
	b.times = append(b.times, now)
	return true
}

// Stream reads entries from in, drops those that exceed the ceiling, and
// forwards the rest to the returned channel. It closes the output channel
// when ctx is cancelled or in is closed.
func Stream(ctx context.Context, in <-chan diff.Entry, c *Ceiling) <-chan diff.Entry {
	out := make(chan diff.Entry)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-in:
				if !ok {
					return
				}
				if c.Allow(entry) {
					select {
					case out <- entry:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
