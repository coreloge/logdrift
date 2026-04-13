// Package partition groups log entries into named buckets based on a
// field value extracted from each entry. This is useful for splitting a
// mixed log stream into per-service or per-tenant sub-streams before
// further processing.
package partition

import (
	"context"
	"sync"

	"github.com/user/logdrift/internal/diff"
)

// Options controls how the Partitioner behaves.
type Options struct {
	// Field is the entry field whose value is used as the partition key.
	// Defaults to "service" when empty.
	Field string

	// BufferSize is the channel buffer for each partition bucket.
	// Defaults to 64.
	BufferSize int
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Field:      "service",
		BufferSize: 64,
	}
}

// Partitioner fans a single entry stream out into per-key channels.
type Partitioner struct {
	opts    Options
	mu      sync.RWMutex
	buckets map[string]chan diff.Entry
}

// New creates a Partitioner with the given options.
func New(opts Options) *Partitioner {
	if opts.Field == "" {
		opts.Field = DefaultOptions().Field
	}
	if opts.BufferSize <= 0 {
		opts.BufferSize = DefaultOptions().BufferSize
	}
	return &Partitioner{
		opts:    opts,
		buckets: make(map[string]chan diff.Entry),
	}
}

// Bucket returns the channel for the given partition key, creating it if
// it does not yet exist.
func (p *Partitioner) Bucket(key string) <-chan diff.Entry {
	p.mu.Lock()
	defer p.mu.Unlock()
	if ch, ok := p.buckets[key]; ok {
		return ch
	}
	ch := make(chan diff.Entry, p.opts.BufferSize)
	p.buckets[key] = ch
	return ch
}

// Keys returns the current set of partition keys in an unspecified order.
func (p *Partitioner) Keys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	keys := make([]string, 0, len(p.buckets))
	for k := range p.buckets {
		keys = append(keys, k)
	}
	return keys
}

// Stream reads entries from in, routes each one to the appropriate bucket
// channel, and returns when ctx is cancelled or in is closed.
func (p *Partitioner) Stream(ctx context.Context, in <-chan diff.Entry) {
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-in:
			if !ok {
				return
			}
			key := entry.Service
			if p.opts.Field != "service" {
				if v, exists := entry.Fields[p.opts.Field]; exists {
					key = v
				}
			}
			p.mu.Lock()
			ch, ok := p.buckets[key]
			if !ok {
				ch = make(chan diff.Entry, p.opts.BufferSize)
				p.buckets[key] = ch
			}
			p.mu.Unlock()
			select {
			case ch <- entry:
			case <-ctx.Done():
				return
			}
		}
	}
}
