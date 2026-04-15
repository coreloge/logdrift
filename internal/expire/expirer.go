// Package expire provides time-based expiry filtering for log entry streams.
// Entries older than the configured TTL are silently dropped.
package expire

import (
	"errors"
	"time"

	"github.com/robmorgan/logdrift/internal/diff"
)

// DefaultOptions returns an Options with a 5-minute TTL and wall-clock time source.
func DefaultOptions() Options {
	return Options{
		TTL: 5 * time.Minute,
		Now: time.Now,
	}
}

// Options configures the Expirer.
type Options struct {
	// TTL is the maximum age of an entry. Entries older than TTL are dropped.
	TTL time.Duration
	// Now is the time source used to evaluate entry age. Defaults to time.Now.
	Now func() time.Time
}

// Expirer drops log entries whose timestamp is older than the configured TTL.
type Expirer struct {
	opts Options
}

// New creates a new Expirer with the given options.
// Returns an error if TTL is zero or negative.
func New(opts Options) (*Expirer, error) {
	if opts.TTL <= 0 {
		return nil, errors.New("expire: TTL must be positive")
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return &Expirer{opts: opts}, nil
}

// Allow returns true if the entry is within the TTL window.
func (e *Expirer) Allow(entry diff.Entry) bool {
	if entry.Timestamp.IsZero() {
		return true
	}
	return e.opts.Now().Sub(entry.Timestamp) <= e.opts.TTL
}
