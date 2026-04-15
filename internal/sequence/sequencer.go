// Package sequence assigns monotonically increasing sequence numbers to log
// entries as they pass through the pipeline. This makes it easy to detect
// gaps or reordering across services.
package sequence

import (
	"sync"

	"github.com/yourorg/logdrift/internal/diff"
)

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Field:   "_seq",
		PerService: false,
	}
}

// Options controls sequencer behaviour.
type Options struct {
	// Field is the entry Extra key that will hold the sequence number.
	Field string

	// PerService, when true, maintains a separate counter per service label.
	// When false a single global counter is used.
	PerService bool
}

// Sequencer stamps each entry with a monotonically increasing integer.
type Sequencer struct {
	opts    Options
	mu      sync.Mutex
	global  uint64
	service map[string]uint64
}

// New returns an initialised Sequencer.
func New(opts Options) (*Sequencer, error) {
	if opts.Field == "" {
		opts.Field = DefaultOptions().Field
	}
	return &Sequencer{
		opts:    opts,
		service: make(map[string]uint64),
	}, nil
}

// Apply stamps e with the next sequence number and returns the updated entry.
// The original entry is not mutated.
func (s *Sequencer) Apply(e diff.Entry) diff.Entry {
	s.mu.Lock()
	var n uint64
	if s.opts.PerService {
		s.service[e.Service]++
		n = s.service[e.Service]
	} else {
		s.global++
		n = s.global
	}
	s.mu.Unlock()

	out := e
	extra := make(map[string]string, len(e.Extra)+1)
	for k, v := range e.Extra {
		extra[k] = v
	}
	extra[s.opts.Field] = formatUint(n)
	out.Extra = extra
	return out
}

// Reset zeroes all counters.
func (s *Sequencer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.global = 0
	s.service = make(map[string]uint64)
}

func formatUint(n uint64) string {
	return fmt.Sprintf("%d", n)
}
