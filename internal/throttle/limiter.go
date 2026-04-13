package throttle

import (
	"sync"

	"github.com/user/logdrift/internal/diff"
)

// DefaultOptions returns an Options with no level limits applied.
func DefaultOptions() Options {
	return Options{
		LevelRates: map[string]float64{},
	}
}

// Options configures the per-level throttle.
type Options struct {
	// LevelRates maps log level strings (e.g. "error", "warn") to a maximum
	// number of entries allowed per second for that level.
	LevelRates map[string]float64
}

// Limiter enforces per-level rate limits on a stream of log entries.
type Limiter struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*bucket
}

// New creates a Limiter with the provided options.
func New(opts Options) *Limiter {
	return &Limiter{
		opts:    opts,
		buckets: make(map[string]*bucket),
	}
}

// Allow returns true if the entry's level is within its configured rate limit.
// Levels with no configured rate are always allowed.
func (l *Limiter) Allow(entry diff.Entry) bool {
	rate, ok := l.opts.LevelRates[entry.Level]
	if !ok || rate <= 0 {
		return true
	}

	l.mu.Lock()
	b, exists := l.buckets[entry.Level]
	if !exists {
		b = newBucket(rate)
		l.buckets[entry.Level] = b
	}
	l.mu.Unlock()

	return b.allow()
}
