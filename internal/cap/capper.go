// Package cap provides a pipeline stage that caps (limits) the total number
// of log entries forwarded downstream. Once the cap is reached the stream is
// closed gracefully, making it useful for bounded replay and testing scenarios.
package cap

import (
	"context"
	"errors"

	"github.com/user/logdrift/internal/diff"
)

// Options controls the behaviour of the Capper.
type Options struct {
	// MaxEntries is the maximum number of entries to forward before closing
	// the output channel. A value of 0 means unlimited (pass-through).
	MaxEntries int
}

// DefaultOptions returns an Options with no cap applied.
func DefaultOptions() Options {
	return Options{MaxEntries: 0}
}

// Capper forwards up to MaxEntries log entries then stops.
type Capper struct {
	opts Options
}

// New creates a Capper. Returns an error if MaxEntries is negative.
func New(opts Options) (*Capper, error) {
	if opts.MaxEntries < 0 {
		return nil, errors.New("cap: MaxEntries must be >= 0")
	}
	return &Capper{opts: opts}, nil
}

// Stream reads from in and writes to the returned channel, stopping after
// MaxEntries entries have been forwarded (or when ctx is cancelled).
// If MaxEntries is 0 all entries are forwarded until in is closed.
func (c *Capper) Stream(ctx context.Context, in <-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry)
	go func() {
		defer close(out)
		var count int
		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- entry:
				case <-ctx.Done():
					return
				}
				count++
				if c.opts.MaxEntries > 0 && count >= c.opts.MaxEntries {
					return
				}
			}
		}
	}()
	return out
}
