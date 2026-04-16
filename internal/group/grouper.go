package group

import (
	"errors"
	"sync"

	"github.com/logdrift/logdrift/internal/diff"
)

// DefaultOptions returns a safe default Options.
func DefaultOptions() Options {
	return Options{
		Field:    "service",
		MaxGroups: 64,
	}
}

// Options configures the Grouper.
type Options struct {
	// Field is the log entry field used as the group key.
	Field string

	// MaxGroups caps the number of distinct groups retained.
	MaxGroups int
}

// Grouper partitions log entries into named buckets keyed by a field value.
type Grouper struct {
	opts   Options
	mu     sync.RWMutex
	groups map[string][]diff.Entry
}

// New creates a Grouper. Returns an error when opts.Field is empty or
// opts.MaxGroups is non-positive.
func New(opts Options) (*Grouper, error) {
	if opts.Field == "" {
		return nil, errors.New("group: Field must not be empty")
	}
	if opts.MaxGroups <= 0 {
		return nil, errors.New("group: MaxGroups must be positive")
	}
	return &Grouper{
		opts:   opts,
		groups: make(map[string][]diff.Entry),
	}, nil
}

// Record adds entry to the appropriate group bucket. If the key is not yet
// known and the group limit has been reached the entry is silently dropped.
func (g *Grouper) Record(entry diff.Entry) {
	key := fieldValue(entry, g.opts.Field)
	if key == "" {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.groups[key]; !ok {
		if len(g.groups) >= g.opts.MaxGroups {
			return
		}
		g.groups[key] = nil
	}
	g.groups[key] = append(g.groups[key], entry)
}

// Get returns a copy of the entries recorded under key, or nil when absent.
func (g *Grouper) Get(key string) []diff.Entry {
	g.mu.RLock()
	defer g.mu.RUnlock()
	src := g.groups[key]
	if src == nil {
		return nil
	}
	out := make([]diff.Entry, len(src))
	copy(out, src)
	return out
}

// Keys returns all known group keys in an unspecified order.
func (g *Grouper) Keys() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	keys := make([]string, 0, len(g.groups))
	for k := range g.groups {
		keys = append(keys, k)
	}
	return keys
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
