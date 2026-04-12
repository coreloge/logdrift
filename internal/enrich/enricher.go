// Package enrich provides field enrichment for log entries,
// allowing static or dynamic key-value pairs to be attached
// to every entry passing through the pipeline.
package enrich

import (
	"sync"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the Enricher.
type Options struct {
	// StaticFields are added to every entry unconditionally.
	StaticFields map[string]string
	// OverwriteExisting controls whether static fields overwrite
	// values already present in the entry's Extra map.
	OverwriteExisting bool
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		StaticFields:      map[string]string{},
		OverwriteExisting: false,
	}
}

// Enricher attaches additional fields to log entries.
type Enricher struct {
	mu   sync.RWMutex
	opts Options
}

// New creates a new Enricher with the given options.
func New(opts Options) *Enricher {
	if opts.StaticFields == nil {
		opts.StaticFields = map[string]string{}
	}
	return &Enricher{opts: opts}
}

// Apply returns a copy of entry with static fields merged into Extra.
func (e *Enricher) Apply(entry diff.Entry) diff.Entry {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.opts.StaticFields) == 0 {
		return entry
	}

	extra := make(map[string]string, len(entry.Extra)+len(e.opts.StaticFields))
	for k, v := range entry.Extra {
		extra[k] = v
	}
	for k, v := range e.opts.StaticFields {
		if _, exists := extra[k]; !exists || e.opts.OverwriteExisting {
			extra[k] = v
		}
	}
	entry.Extra = extra
	return entry
}

// SetField adds or updates a single static field at runtime.
func (e *Enricher) SetField(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.opts.StaticFields[key] = value
}

// RemoveField removes a static field at runtime.
func (e *Enricher) RemoveField(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.opts.StaticFields, key)
}
