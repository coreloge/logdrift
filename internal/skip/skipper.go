// Package skip provides a pipeline stage that drops the first N log entries
// from a stream, forwarding all subsequent entries unchanged. This is useful
// for skipping header lines or warm-up noise at the start of a log file.
package skip

import (
	"context"
	"errors"
	"fmt"

	"logdrift/internal/diff"
)

// Options configures the Skipper.
type Options struct {
	// Count is the number of entries to drop from the head of the stream.
	// Must be >= 0.
	Count int
	// PerService, when true, maintains an independent counter per service
	// label so that each service skips its own first Count entries.
	PerService bool
}

// DefaultOptions returns an Options that skips nothing.
func DefaultOptions() Options {
	return Options{Count: 0}
}

// Skipper drops the first Count entries from an entry stream.
type Skipper struct {
	opts     Options
	global   int
	perSvc   map[string]int
}

// New creates a Skipper with the provided Options.
// Returns an error if Count is negative.
func New(opts Options) (*Skipper, error) {
	if opts.Count < 0 {
		return nil, fmt.Errorf("skip: Count must be >= 0, got %d", opts.Count)
	}
	return &Skipper{
		opts:   opts,
		perSvc: make(map[string]int),
	}, nil
}

// Allow returns true when the entry should be forwarded downstream.
func (s *Skipper) Allow(e diff.Entry) bool {
	if s.opts.Count == 0 {
		return true
	}
	if s.opts.PerService {
		seen := s.perSvc[e.Service]
		if seen < s.opts.Count {
			s.perSvc[e.Service]++
			return false
		}
		return true
	}
	if s.global < s.opts.Count {
		s.global++
		return false
	}
	return true
}

// Stream reads from in, skips the first Count entries according to the
// configured policy, and forwards the rest to the returned channel.
// The output channel is closed when ctx is cancelled or in is closed.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) (<-chan diff.Entry, error) {
	sk, err := New(opts)
	if err != nil {
		return nil, errors.New("skip.Stream: " + err.Error())
	}
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
				if sk.Allow(e) {
					select {
					case out <- e:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out, nil
}
