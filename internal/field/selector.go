// Package field provides utilities for selecting and projecting specific
// fields from log entries, emitting a reduced entry downstream.
package field

import (
	"errors"

	"github.com/logdrift/logdrift/internal/diff"
)

// Options configures the field selector.
type Options struct {
	// Fields lists the entry fields to retain. "message", "level",
	// "service", and "timestamp" are always kept.
	Fields []string
}

// DefaultOptions returns an Options that retains all extra fields.
func DefaultOptions() Options {
	return Options{}
}

// Selector projects log entries to a subset of their extra fields.
type Selector struct {
	opts Options
	keep map[string]struct{}
}

// New creates a Selector from opts. Returns an error if Fields contains
// empty strings.
func New(opts Options) (*Selector, error) {
	keep := make(map[string]struct{}, len(opts.Fields))
	for _, f := range opts.Fields {
		if f == "" {
			return nil, errors.New("field: empty field name")
		}
		keep[f] = struct{}{}
	}
	return &Selector{opts: opts, keep: keep}, nil
}

// Apply returns a copy of e with Extra reduced to the configured fields.
// If no fields are configured all Extra fields are retained.
func (s *Selector) Apply(e diff.Entry) diff.Entry {
	if len(s.keep) == 0 || len(e.Extra) == 0 {
		return e
	}
	out := e
	out.Extra = make(map[string]string, len(s.keep))
	for k, v := range etif _, ok := s.keep[k]; ok {
			out.Extra[k] = v
		}
	}
	return out
}

// Stream reads entries, applies the selector, and forwards results
// to the returned channel. The channel is closed when ctx is done or in
// is closed.
func Stream(ctx interface{ Done() <-chan struct{} }, in <-chan diff.Entry, s *Selector) <-chan diff.Entry {
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
				out <- s.Apply(e)
			}
		}
	}()
	return out
}
