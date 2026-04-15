// Package sample implements reservoir sampling for log entry streams.
// It maintains a fixed-size sample of entries seen so far, giving each
// entry an equal probability of appearing in the final sample.
package sample

import (
	"context"
	"math/rand"
	"sync"

	"github.com/yourorg/logdrift/internal/diff"
)

// Options configures the reservoir sampler.
type Options struct {
	// Size is the maximum number of entries to retain.
	Size int
	// Seed is the random seed. Zero means non-deterministic.
	Seed int64
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Size: 100,
	}
}

// Reservoir holds a fixed-size random sample of log entries.
type Reservoir struct {
	mu      sync.Mutex
	opts    Options
	buf     []diff.Entry
	count   int
	rng     *rand.Rand
}

// New creates a new Reservoir with the given options.
// Returns an error if Size is less than 1.
func New(opts Options) (*Reservoir, error) {
	if opts.Size < 1 {
		return nil, fmt.Errorf("sample: size must be at least 1, got %d", opts.Size)
	}
	src := rand.NewSource(opts.Seed)
	return &Reservoir{
		opts: opts,
		buf:  make([]diff.Entry, 0, opts.Size),
		rng:  rand.New(src),
	}, nil
}

// Add adds an entry to the reservoir using Algorithm R.
func (r *Reservoir) Add(e diff.Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.count++
	if len(r.buf) < r.opts.Size {
		r.buf = append(r.buf, e)
		return
	}
	j := r.rng.Intn(r.count)
	if j < r.opts.Size {
		r.buf[j] = e
	}
}

// Snapshot returns a copy of the current sample.
func (r *Reservoir) Snapshot() []diff.Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]diff.Entry, len(r.buf))
	copy(out, r.buf)
	return out
}

// Len returns the number of entries currently in the reservoir.
func (r *Reservoir) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.buf)
}

// Reset clears the reservoir.
func (r *Reservoir) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf = r.buf[:0]
	r.count = 0
}

// Stream reads entries from in, adds each to the reservoir, and forwards
// them unchanged to the returned channel. It closes the output channel
// when ctx is cancelled or in is closed.
func Stream(ctx context.Context, r *Reservoir, in <-chan diff.Entry) <-chan diff.Entry {
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
				r.Add(e)
				select {
				case out <- e:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
