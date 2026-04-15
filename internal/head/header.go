// Package head provides a pipeline stage that forwards only the first N
// log entries from a stream, then stops — analogous to the Unix `head` command.
package head

import (
	"context"
	"errors"

	"github.com/your-org/logdrift/internal/diff"
)

// Options configures the Header.
type Options struct {
	// Max is the maximum number of entries to forward. Must be >= 1.
	Max int
}

// DefaultOptions returns a conservative default: forward the first 10 entries.
func DefaultOptions() Options {
	return Options{Max: 10}
}

// Header forwards at most Max entries from the input channel, then closes the
// output channel and returns.
type Header struct {
	opts Options
}

// New creates a Header. Returns an error if opts.Max is less than 1.
func New(opts Options) (*Header, error) {
	if opts.Max < 1 {
		return nil, errors.New("head: Max must be >= 1")
	}
	return &Header{opts: opts}, nil
}

// Stream reads from in and writes up to Max entries to the returned channel.
// The output channel is closed when the limit is reached or ctx is cancelled.
func (h *Header) Stream(ctx context.Context, in <-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry)
	go func() {
		defer close(out)
		count := 0
		for {
			if count >= h.opts.Max {
				return
			}
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- entry:
					count++
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
