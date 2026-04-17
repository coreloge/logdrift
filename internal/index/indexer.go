// Package index provides field-based indexing of log entries for fast lookup.
package index

import (
	"errors"
	"sync"

	"github.com/humanlogio/logdrift/internal/diff"
)

// ErrEmptyField is returned when an empty index field is provided.
var ErrEmptyField = errors.New("index: field must not be empty")

// Options configures the Indexer.
type Options struct {
	// Field is the entry field to index on (e.g. "request_id").
	Field string
	// MaxEntries caps the total number of indexed entries. Zero means unlimited.
	MaxEntries int
}

// DefaultOptions returns a sensible default configuration.
func DefaultOptions() Options {
	return Options{
		Field:      "request_id",
		MaxEntries: 10_000,
	}
}

// Indexer maintains an in-memory inverted index from field value → entries.
type Indexer struct {
	mu      sync.RWMutex
	field   string
	max     int
	total   int
	index   map[string][]diff.Entry
}

// New creates a new Indexer. Returns ErrEmptyField if opts.Field is empty.
func New(opts Options) (*Indexer, error) {
	if opts.Field == "" {
		return nil, ErrEmptyField
	}
	return &Indexer{
		field: opts.Field,
		max:   opts.MaxEntries,
		index: make(map[string][]diff.Entry),
	}, nil
}

// Add indexes the entry under its field value. If MaxEntries is set and
// reached, the entry is silently dropped.
func (idx *Indexer) Add(e diff.Entry) {
	val, ok := e.Extra[idx.field]
	if !ok {
		return
	}
	key, ok := val.(string)
	if !ok {
		return
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if idx.max > 0 && idx.total >= idx.max {
		return
	}
	idx.index[key] = append(idx.index[key], e)
	idx.total++
}

// Lookup returns all entries indexed under key, or nil if none exist.
func (idx *Indexer) Lookup(key string) []diff.Entry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	if entries, ok := idx.index[key]; ok {
		out := make([]diff.Entry, len(entries))
		copy(out, entries)
		return out
	}
	return nil
}

// Reset clears all indexed data.
func (idx *Indexer) Reset() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.index = make(map[string][]diff.Entry)
	idx.total = 0
}
