// Package window provides a sliding time-window aggregator that groups
// log entries into fixed-duration buckets for trend analysis.
package window

import (
	"sync"
	"time"

	"logdrift/internal/diff"
)

// Options configures the sliding window behaviour.
type Options struct {
	// Width is the duration of each bucket.
	Width time.Duration
	// MaxBuckets is the maximum number of buckets retained in memory.
	MaxBuckets int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Width:      30 * time.Second,
		MaxBuckets: 10,
	}
}

// Bucket holds all entries that arrived within a single time window.
type Bucket struct {
	Start   time.Time
	Entries []diff.Entry
}

// Window accumulates entries into time-bounded buckets.
type Window struct {
	mu      sync.Mutex
	opts    Options
	buckets []Bucket
}

// New creates a Window with the given options.
func New(opts Options) *Window {
	if opts.Width <= 0 {
		opts.Width = DefaultOptions().Width
	}
	if opts.MaxBuckets <= 0 {
		opts.MaxBuckets = DefaultOptions().MaxBuckets
	}
	return &Window{opts: opts}
}

// Add places an entry into the appropriate bucket, creating one if needed.
func (w *Window) Add(e diff.Entry) {
	w.mu.Lock()
	defer w.mu.Unlock()

	bucketStart := e.Timestamp.Truncate(w.opts.Width)

	// Find existing bucket.
	for i := range w.buckets {
		if w.buckets[i].Start.Equal(bucketStart) {
			w.buckets[i].Entries = append(w.buckets[i].Entries, e)
			return
		}
	}

	// Create new bucket.
	w.buckets = append(w.buckets, Bucket{Start: bucketStart, Entries: []diff.Entry{e}})

	// Evict oldest buckets beyond MaxBuckets.
	if len(w.buckets) > w.opts.MaxBuckets {
		w.buckets = w.buckets[len(w.buckets)-w.opts.MaxBuckets:]
	}
}

// Buckets returns a snapshot of the current buckets in chronological order.
func (w *Window) Buckets() []Bucket {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]Bucket, len(w.buckets))
	copy(out, w.buckets)
	return out
}

// Reset clears all buckets.
func (w *Window) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buckets = nil
}
