// Package jitter adds randomised delay to log entry streams,
// useful for smoothing bursty sources or simulating realistic latency
// in replay and testing scenarios.
package jitter

import (
	"context"
	"math/rand"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

// Options controls the jitter behaviour.
type Options struct {
	// MinDelay is the minimum delay applied to each entry.
	MinDelay time.Duration
	// MaxDelay is the maximum delay applied to each entry.
	MaxDelay time.Duration
	// Rand is the random source; defaults to a package-level source when nil.
	Rand *rand.Rand
}

// DefaultOptions returns Options with zero min and 50 ms max delay.
func DefaultOptions() Options {
	return Options{
		MinDelay: 0,
		MaxDelay: 50 * time.Millisecond,
	}
}

// Jitter holds the configured delay parameters.
type Jitter struct {
	opts Options
	rng  *rand.Rand
}

// New creates a Jitter from opts. If opts.MaxDelay < opts.MinDelay an error
// is returned.
func New(opts Options) (*Jitter, error) {
	if opts.MaxDelay < opts.MinDelay {
		return nil, fmt.Errorf("jitter: MaxDelay (%v) must be >= MinDelay (%v)", opts.MaxDelay, opts.MinDelay)
	}
	rng := opts.Rand
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &Jitter{opts: opts, rng: rng}, nil
}

// delay returns a random duration in [MinDelay, MaxDelay].
func (j *Jitter) delay() time.Duration {
	span := j.opts.MaxDelay - j.opts.MinDelay
	if span == 0 {
		return j.opts.MinDelay
	}
	return j.opts.MinDelay + time.Duration(j.rng.Int63n(int64(span)))
}

// Stream reads entries from in, waits a random jitter delay, then forwards
// them to the returned channel. The output channel is closed when in is
// closed or ctx is cancelled.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) (<-chan diff.Entry, error) {
	j, err := New(opts)
	if err != nil {
		return nil, err
	}
	out := make(chan diff.Entry, cap(in))
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
				select {
				case <-ctx.Done():
					return
				case <-time.After(j.delay()):
				}
				select {
				case out <- entry:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out, nil
}
