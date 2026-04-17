// Package gate provides a conditional pass-through gate that opens or closes
// the stream based on a user-supplied predicate evaluated against each entry.
package gate

import (
	"context"
	"errors"

	"github.com/humanlogio/logdrift/internal/diff"
)

// Options configures the Gate.
type Options struct {
	// Predicate is called for every entry. The gate is open (entry forwarded)
	// when Predicate returns true.
	Predicate func(entry diff.Entry) bool

	// Invert flips the predicate result so the gate forwards entries that
	// would otherwise be blocked.
	Invert bool
}

// DefaultOptions returns an Options that passes all entries through.
func DefaultOptions() Options {
	return Options{
		Predicate: func(diff.Entry) bool { return true },
	}
}

// Gate evaluates a predicate for each entry and either forwards or drops it.
type Gate struct {
	opts Options
}

// New creates a Gate. Returns an error if Predicate is nil.
func New(opts Options) (*Gate, error) {
	if opts.Predicate == nil {
		return nil, errors.New("gate: Predicate must not be nil")
	}
	return &Gate{opts: opts}, nil
}

// Allow reports whether the entry should pass through the gate.
func (g *Gate) Allow(entry diff.Entry) bool {
	result := g.opts.Predicate(entry)
	if g.opts.Invert {
		return !result
	}
	return result
}

// Stream reads entries from in, forwards those that pass the gate to the
// returned channel, and stops when ctx is cancelled.
func Stream(ctx context.Context, g *Gate, in <-chan diff.Entry) <-chan diff.Entry {
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
				if g.Allow(entry) {
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
