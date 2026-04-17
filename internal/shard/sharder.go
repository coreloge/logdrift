// Package shard distributes log entries across a fixed number of buckets
// based on a configurable field value, enabling parallel downstream processing.
package shard

import (
	"errors"
	"hash/fnv"

	"github.com/logdrift/logdrift/internal/diff"
)

// DefaultOptions returns a safe default configuration.
func DefaultOptions() Options {
	return Options{
		Field:  "service",
		Shards: 4,
	}
}

// Options controls sharding behaviour.
type Options struct {
	// Field is the entry field whose value determines the shard.
	Field string
	// Shards is the total number of output buckets (must be >= 1).
	Shards int
}

// Sharder hashes a field value to select one of N output channels.
type Sharder struct {
	opts Options
	out  []chan diff.Entry
}

// New creates a Sharder and allocates output channels.
func New(opts Options) (*Sharder, error) {
	if opts.Field == "" {
		return nil, errors.New("shard: field must not be empty")
	}
	if opts.Shards < 1 {
		return nil, errors.New("shard: shards must be >= 1")
	}
	out := make([]chan diff.Entry, opts.Shards)
	for i := range out {
		out[i] = make(chan diff.Entry, 64)
	}
	return &Sharder{opts: opts, out: out}, nil
}

// Outputs returns the read-only shard channels.
func (s *Sharder) Outputs() []<-chan diff.Entry {
	result := make([]<-chan diff.Entry, len(s.out))
	for i, ch := range s.out {
		result[i] = ch
	}
	return result
}

// Assign routes entry to the appropriate shard channel.
func (s *Sharder) Assign(e diff.Entry) {
	v := fieldValue(e, s.opts.Field)
	h := fnv.New32a()
	_, _ = h.Write([]byte(v))
	idx := int(h.Sum32()) % len(s.out)
	s.out[idx] <- e
}

// Close closes all output channels.
func (s *Sharder) Close() {
	for _, ch := range s.out {
		close(ch)
	}
}

func fieldValue(e diff.Entry, field string) string {
	switch field {
	case "service":
		return e.Service
	case "level":
		return e.Level
	case "message":
		return e.Message
	default:
		if e.Extra != nil {
			if v, ok := e.Extra[field]; ok {
				return v
			}
		}
		return ""
	}
}
