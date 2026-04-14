// Package debounce provides a pipeline stage that suppresses repeated log
// entries within a configurable quiet window. Only the first entry in each
// burst is forwarded; subsequent identical (service + level + message) entries
// are dropped until the window expires.
package debounce

import (
	"context"
	"sync"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Window: 2 * time.Second,
	}
}

// Options controls debounce behaviour.
type Options struct {
	// Window is the quiet period after the first occurrence during which
	// duplicate entries are suppressed.
	Window time.Duration
}

type key struct {
	service, level, message string
}

// Debouncer suppresses bursts of identical log entries.
type Debouncer struct {
	opts  Options
	mu    sync.Mutex
	seen  map[key]time.Time
	nowFn func() time.Time
}

// New creates a Debouncer with the supplied options.
func New(opts Options) *Debouncer {
	return &Debouncer{
		opts:  opts,
		seen:  make(map[key]time.Time),
		nowFn: time.Now,
	}
}

// Allow returns true when the entry should be forwarded downstream.
// The first occurrence within any window is always allowed; duplicates
// within the window are suppressed.
func (d *Debouncer) Allow(e diff.Entry) bool {
	k := key{service: e.Service, level: e.Level, message: e.Message}
	now := d.nowFn()

	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.seen[k]; ok && now.Sub(t) < d.opts.Window {
		return false
	}
	d.seen[k] = now
	return true
}

// Reset clears all tracked entries.
func (d *Debouncer) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[key]time.Time)
}

// Stream reads entries from in, forwards non-duplicate entries to the returned
// channel, and stops when ctx is cancelled.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) <-chan diff.Entry {
	out := make(chan diff.Entry)
	d := New(opts)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-in:
				if !ok {
					return
				}
				if d.Allow(e) {
					select {
					case out <- e:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
