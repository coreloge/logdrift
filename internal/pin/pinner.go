package pin

import (
	"errors"
	"sync"

	"github.com/logdrift/logdrift/internal/diff"
)

// DefaultOptions returns a Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxPinned: 100,
	}
}

// Options configures the Pinner.
type Options struct {
	// MaxPinned is the maximum number of pinned entries retained.
	MaxPinned int
}

// Pinner retains a fixed set of log entries that should always be
// surfaced regardless of filtering or sampling decisions downstream.
type Pinner struct {
	opts   Options
	mu     sync.Mutex
	pinned []diff.Entry
}

// New creates a Pinner with the given options.
func New(opts Options) (*Pinner, error) {
	if opts.MaxPinned <= 0 {
		return nil, errors.New("pin: MaxPinned must be greater than zero")
	}
	return &Pinner{opts: opts}, nil
}

// Pin adds an entry to the pinned set. If the set is full the oldest
// entry is evicted.
func (p *Pinner) Pin(e diff.Entry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.pinned) >= p.opts.MaxPinned {
		p.pinned = p.pinned[1:]
	}
	p.pinned = append(p.pinned, e)
}

// All returns a copy of the currently pinned entries in insertion order.
func (p *Pinner) All() []diff.Entry {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]diff.Entry, len(p.pinned))
	copy(out, p.pinned)
	return out
}

// Clear removes all pinned entries.
func (p *Pinner) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pinned = p.pinned[:0]
}
