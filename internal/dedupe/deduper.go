// Package dedupe provides log entry deduplication for logdrift.
// Entries are considered duplicates when their service, level, and message
// fields match within a configurable time window.
package dedupe

import (
	"sync"
	"time"

	"github.com/logdrift/internal/diff"
)

// DefaultOptions returns a Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Window:   5 * time.Second,
		MaxCache: 1024,
	}
}

// Options controls deduplication behaviour.
type Options struct {
	// Window is the duration within which identical entries are suppressed.
	Window time.Duration
	// MaxCache is the maximum number of distinct keys held in memory.
	MaxCache int
}

type cacheEntry struct {
	seen time.Time
}

// Deduper suppresses repeated log entries within a sliding time window.
type Deduper struct {
	opts  Options
	mu    sync.Mutex
	cache map[string]cacheEntry
}

// New creates a Deduper with the provided Options.
func New(opts Options) *Deduper {
	if opts.Window <= 0 {
		opts.Window = DefaultOptions().Window
	}
	if opts.MaxCache <= 0 {
		opts.MaxCache = DefaultOptions().MaxCache
	}
	return &Deduper{
		opts:  opts,
		cache: make(map[string]cacheEntry, opts.MaxCache),
	}
}

// IsDuplicate returns true when an equivalent entry was seen within the window.
// It records the entry if it is new or has expired.
func (d *Deduper) IsDuplicate(entry diff.LogEntry) bool {
	key := entry.Service + "\x00" + entry.Level + "\x00" + entry.Message
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	if ce, ok := d.cache[key]; ok {
		if now.Sub(ce.seen) < d.opts.Window {
			return true
		}
	}

	// Evict oldest entry when cache is full.
	if len(d.cache) >= d.opts.MaxCache {
		d.evictOldest(now)
	}

	d.cache[key] = cacheEntry{seen: now}
	return false
}

// Reset clears the deduplication cache.
func (d *Deduper) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[string]cacheEntry, d.opts.MaxCache)
}

// evictOldest removes the single oldest entry; caller must hold d.mu.
func (d *Deduper) evictOldest(now time.Time) {
	var oldestKey string
	var oldestTime time.Time
	for k, v := range d.cache {
		if oldestKey == "" || v.seen.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.seen
		}
	}
	if oldestKey != "" {
		delete(d.cache, oldestKey)
	}
}
