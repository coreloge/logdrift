// Package distinct provides a pipeline stage that forwards only entries
// with a unique value for a given field, dropping subsequent duplicates.
package distinct

import (
	"context"
	"errors"
	"sync"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the Distinct filter.
type Options struct {
	// Field is the entry field to deduplicate on ("message", "level", or an extra key).
	Field string
	// MaxTracked caps the number of distinct values remembered (0 = unlimited).
	MaxTracked int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Field:      "message",
		MaxTracked: 0,
	}
}

// Distinct drops entries whose field value has already been seen.
type Distinct struct {
	opts Options
	mu   sync.Mutex
	seen map[string]struct{}
}

// New creates a Distinct filter. Returns an error if Field is empty.
func New(opts Options) (*Distinct, error) {
	if opts.Field == "" {
		return nil, errors.New("distinct: field must not be empty")
	}
	return &Distinct{opts: opts, seen: make(map[string]struct{})}, nil
}

// Allow returns true the first time a given field value is encountered.
func (d *Distinct) Allow(e diff.Entry) bool {
	v := fieldValue(e, d.opts.Field)
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.seen[v]; ok {
		return false
	}
	if d.opts.MaxTracked > 0 && len(d.seen) >= d.opts.MaxTracked {
		return true // capacity full – pass through but don't track
	}
	d.seen[v] = struct{}{}
	return true
}

// Reset clears all remembered values.
func (d *Distinct) Reset() {
	d.mu.Lock()
	d.seen = make(map[string]struct{})
	d.mu.Unlock()
}

// Stream forwards only entries with a previously unseen field value.
func Stream(ctx context.Context, in <-chan diff.Entry, d *Distinct) <-chan diff.Entry {
	out := make(chan diff.Entry)
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
				if d.Allow(e) {
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

func fieldValue(e diff.Entry, field string) string {
	switch field {
	case "message":
		return e.Message
	case "level":
		return e.Level
	case "service":
		return e.Service
	default:
		if e.Extra != nil {
			if v, ok := e.Extra[field]; ok {
				return v
			}
		}
		return ""
	}
}
