// Package timeskew detects log entries whose timestamps deviate significantly
// from wall-clock time or from a reference service's timestamp stream.
package timeskew

import (
	"fmt"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// DefaultOptions returns a Options with conservative defaults.
func DefaultOptions() Options {
	return Options{
		MaxSkew:        5 * time.Second,
		ReferenceNow:   time.Now,
	}
}

// Options controls skew detection behaviour.
type Options struct {
	// MaxSkew is the maximum allowed difference between an entry's timestamp
	// and the reference time before it is flagged.
	MaxSkew time.Duration

	// ReferenceNow returns the current reference time. Defaults to time.Now.
	// Inject a fixed function in tests for determinism.
	ReferenceNow func() time.Time
}

// Violation describes a single skew violation.
type Violation struct {
	Service  string
	EntryTS  time.Time
	Now      time.Time
	Skew     time.Duration
	MaxSkew  time.Duration
}

func (v Violation) Error() string {
	return fmt.Sprintf("timeskew: service %q entry ts %s deviates %s (max %s)",
		v.Service, v.EntryTS.Format(time.RFC3339), v.Skew.Round(time.Millisecond), v.MaxSkew)
}

// Detector checks log entries for timestamp skew.
type Detector struct {
	opts Options
}

// New creates a Detector with the provided options.
func New(opts Options) (*Detector, error) {
	if opts.MaxSkew <= 0 {
		return nil, fmt.Errorf("timeskew: MaxSkew must be positive, got %s", opts.MaxSkew)
	}
	if opts.ReferenceNow == nil {
		opts.ReferenceNow = time.Now
	}
	return &Detector{opts: opts}, nil
}

// Check returns a Violation if the entry's timestamp deviates beyond MaxSkew,
// or nil when the entry is within the acceptable window.
func (d *Detector) Check(entry diff.Entry) *Violation {
	now := d.opts.ReferenceNow()
	skew := entry.Timestamp.Sub(now)
	if skew < 0 {
		skew = -skew
	}
	if skew <= d.opts.MaxSkew {
		return nil
	}
	return &Violation{
		Service: entry.Service,
		EntryTS: entry.Timestamp,
		Now:     now,
		Skew:    skew,
		MaxSkew: d.opts.MaxSkew,
	}
}
