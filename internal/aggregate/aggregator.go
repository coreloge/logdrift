// Package aggregate groups log entries by a key field and computes
// summary statistics (count, first-seen, last-seen) over a sliding window.
package aggregate

import (
	"sync"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// Options controls aggregation behaviour.
type Options struct {
	// KeyField is the entry field used to group log lines (e.g. "level", "service").
	KeyField string
	// Window is the duration over which buckets are retained.
	Window time.Duration
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		KeyField: "level",
		Window:   time.Minute,
	}
}

// Bucket holds aggregated statistics for a single key value.
type Bucket struct {
	Key       string
	Count     int
	FirstSeen time.Time
	LastSeen  time.Time
}

// Aggregator accumulates log entries and exposes per-key summaries.
type Aggregator struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*Bucket
}

// New creates an Aggregator with the provided options.
func New(opts Options) *Aggregator {
	if opts.KeyField == "" {
		opts.KeyField = DefaultOptions().KeyField
	}
	if opts.Window <= 0 {
		opts.Window = DefaultOptions().Window
	}
	return &Aggregator{
		opts:    opts,
		buckets: make(map[string]*Bucket),
	}
}

// Record ingests a single log entry, updating the appropriate bucket.
func (a *Aggregator) Record(entry diff.Entry) {
	key := entry.Fields[a.opts.KeyField]
	if key == "" {
		key = "(unknown)"
	}
	now := entry.Timestamp
	if now.IsZero() {
		now = time.Now()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.evict(now)

	b, ok := a.buckets[key]
	if !ok {
		b = &Bucket{Key: key, FirstSeen: now}
		a.buckets[key] = b
	}
	b.Count++
	b.LastSeen = now
}

// Snapshot returns a copy of all current buckets.
func (a *Aggregator) Snapshot() []Bucket {
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make([]Bucket, 0, len(a.buckets))
	for _, b := range a.buckets {
		out = append(out, *b)
	}
	return out
}

// Reset clears all buckets.
func (a *Aggregator) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buckets = make(map[string]*Bucket)
}

// evict removes buckets whose LastSeen is outside the window.
// Caller must hold a.mu.
func (a *Aggregator) evict(now time.Time) {
	cutoff := now.Add(-a.opts.Window)
	for k, b := range a.buckets {
		if b.LastSeen.Before(cutoff) {
			delete(a.buckets, k)
		}
	}
}
