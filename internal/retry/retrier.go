// Package retry provides a configurable retry mechanism for log pipeline
// operations that may fail transiently (e.g. file reads, network sinks).
package retry

import (
	"context"
	"errors"
	"time"
)

// ErrMaxAttemptsReached is returned when all retry attempts are exhausted.
var ErrMaxAttemptsReached = errors.New("retry: max attempts reached")

// Options configures the retry behaviour.
type Options struct {
	// MaxAttempts is the total number of attempts (including the first). 0 means unlimited.
	MaxAttempts int
	// InitialDelay is the wait time before the second attempt.
	InitialDelay time.Duration
	// MaxDelay caps the exponential back-off delay.
	MaxDelay time.Duration
	// Multiplier scales the delay after each failure.
	Multiplier float64
}

// DefaultOptions returns sensible defaults for transient retries.
func DefaultOptions() Options {
	return Options{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// Retrier executes an operation with exponential back-off.
type Retrier struct {
	opts Options
}

// New creates a Retrier with the given options.
func New(opts Options) *Retrier {
	if opts.Multiplier <= 0 {
		opts.Multiplier = 2.0
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = 30 * time.Second
	}
	return &Retrier{opts: opts}
}

// Do calls fn repeatedly until it returns nil, the context is cancelled, or
// the maximum number of attempts is exhausted.
func (r *Retrier) Do(ctx context.Context, fn func() error) error {
	delay := r.opts.InitialDelay
	for attempt := 1; ; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if r.opts.MaxAttempts > 0 && attempt >= r.opts.MaxAttempts {
			return errors.Join(ErrMaxAttemptsReached, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay = time.Duration(float64(delay) * r.opts.Multiplier)
		if delay > r.opts.MaxDelay {
			delay = r.opts.MaxDelay
		}
	}
}
