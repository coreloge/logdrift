// Package counter provides a per-field occurrence counter that tracks how
// many times each distinct value appears across a stream of log entries.
package counter

import (
	"errors"
	"fmt"
	"sync"

	"github.com/yourorg/logdrift/internal/diff"
)

// Options configures the Counter.
type Options struct {
	// Field is the entry field to count ("level", "service", or any Extra key).
	Field string
	// OutputField is the Extra key where the running count is written.
	// Defaults to "_count".
	OutputField string
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Field:       "level",
		OutputField: "_count",
	}
}

// Counter counts occurrences of distinct values for a configured field.
type Counter struct {
	opts   Options
	mu     sync.Mutex
	counts map[string]int64
}

// New creates a Counter from opts.
func New(opts Options) (*Counter, error) {
	if opts.Field == "" {
		return nil, errors.New("counter: field must not be empty")
	}
	if opts.OutputField == "" {
		opts.OutputField = "_count"
	}
	return &Counter{
		opts:   opts,
		counts: make(map[string]int64),
	}, nil
}

// Record increments the counter for the field value found in entry and returns
// a copy of the entry with the running count written into OutputField.
func (c *Counter) Record(entry diff.Entry) diff.Entry {
	val := fieldValue(entry, c.opts.Field)

	c.mu.Lock()
	c.counts[val]++
	n := c.counts[val]
	c.mu.Unlock()

	out := copyEntry(entry)
	if out.Extra == nil {
		out.Extra = make(map[string]string)
	}
	out.Extra[c.opts.OutputField] = fmt.Sprintf("%d", n)
	return out
}

// Counts returns a snapshot of the current counts by value.
func (c *Counter) Counts() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int64, len(c.counts))
	for k, v := range c.counts {
		out[k] = v
	}
	return out
}

// Reset zeroes all counters.
func (c *Counter) Reset() {
	c.mu.Lock()
	c.counts = make(map[string]int64)
	c.mu.Unlock()
}

func fieldValue(e diff.Entry, field string) string {
	switch field {
	case "level":
		return e.Level
	case "service":
		return e.Service
	case "message":
		return e.Message
	default:
		if e.Extra != nil {
			return e.Extra[field]
		}
		return ""
	}
}

func copyEntry(e diff.Entry) diff.Entry {
	out := e
	if e.Extra != nil {
		out.Extra = make(map[string]string, len(e.Extra))
		for k, v := range e.Extra {
			out.Extra[k] = v
		}
	}
	return out
}
