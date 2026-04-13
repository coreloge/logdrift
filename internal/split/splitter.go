// Package split provides a stream splitter that duplicates log entries
// into two output channels based on a predicate function. Entries matching
// the predicate are sent to the left channel; all others go to the right.
package split

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Predicate is a function that returns true when an entry should be routed
// to the left (matched) channel.
type Predicate func(entry diff.Entry) bool

// Options configures the Splitter.
type Options struct {
	// BufferSize controls the capacity of each output channel.
	BufferSize int
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		BufferSize: 64,
	}
}

// Splitter reads from a source channel and routes each entry to one of two
// output channels depending on whether the predicate matches.
type Splitter struct {
	opts Options
	pred Predicate
}

// New creates a Splitter with the given predicate and options.
func New(pred Predicate, opts Options) *Splitter {
	if opts.BufferSize <= 0 {
		opts.BufferSize = DefaultOptions().BufferSize
	}
	return &Splitter{opts: opts, pred: pred}
}

// Stream reads entries from src and writes matching entries to the first
// returned channel and non-matching entries to the second. Both channels are
// closed when src is exhausted or ctx is cancelled.
func (s *Splitter) Stream(ctx context.Context, src <-chan diff.Entry) (<-chan diff.Entry, <-chan diff.Entry) {
	matched := make(chan diff.Entry, s.opts.BufferSize)
	rest := make(chan diff.Entry, s.opts.BufferSize)

	go func() {
		defer close(matched)
		defer close(rest)
		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-src:
				if !ok {
					return
				}
				if s.pred(entry) {
					matched <- entry
				} else {
					rest <- entry
				}
			}
		}
	}()

	return matched, rest
}
