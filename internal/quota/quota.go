// Package quota enforces per-service log entry quotas over a rolling time window.
// Entries exceeding the configured limit are dropped until the window resets.
package quota

import (
	"errors"
	"sync"
	"time"
)

// DefaultOptions returns a permissive Options with no quota enforced.
func DefaultOptions() Options {
	return Options{
		Max:    0,
		Window: time.Minute,
	}
}

// Options configures the Quota enforcer.
type Options struct {
	// Max is the maximum number of entries allowed per service per Window.
	// Zero means unlimited.
	Max int
	// Window is the rolling duration over which Max is counted.
	Window time.Duration
}

type bucket struct {
	count     int
	windowEnd time.Time
}

// Quota tracks per-service entry counts and enforces a rolling window limit.
type Quota struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*bucket
	now     func() time.Time
}

// New creates a Quota enforcer with the given options.
func New(opts Options) (*Quota, error) {
	if opts.Max < 0 {
		return nil, errors.New("quota: Max must be non-negative")
	}
	if opts.Window <= 0 {
		return nil, errors.New("quota: Window must be positive")
	}
	return &Quota{
		opts:    opts,
		buckets: make(map[string]*bucket),
		now:     time.Now,
	}, nil
}

// Allow reports whether the entry for the given service is within quota.
// It increments the counter for the service and returns false when the limit
// is exceeded within the current window.
func (q *Quota) Allow(service string) bool {
	if q.opts.Max == 0 {
		return true
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	now := q.now()
	b, ok := q.buckets[service]
	if !ok || now.After(b.windowEnd) {
		q.buckets[service] = &bucket{count: 1, windowEnd: now.Add(q.opts.Window)}
		return true
	}
	b.count++
	return b.count <= q.opts.Max
}

// Reset clears all per-service counters.
func (q *Quota) Reset() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.buckets = make(map[string]*bucket)
}
