package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/retry"
)

var errTransient = errors.New("transient error")

func TestDefaultOptions(t *testing.T) {
	opts := retry.DefaultOptions()
	if opts.MaxAttempts != 5 {
		t.Fatalf("expected MaxAttempts=5, got %d", opts.MaxAttempts)
	}
	if opts.Multiplier != 2.0 {
		t.Fatalf("expected Multiplier=2.0, got %f", opts.Multiplier)
	}
}

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	r := retry.New(retry.DefaultOptions())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesUntilSuccess(t *testing.T) {
	opts := retry.Options{
		MaxAttempts:  5,
		InitialDelay: time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}
	r := retry.New(opts)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errTransient
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsMaxAttempts(t *testing.T) {
	opts := retry.Options{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}
	r := retry.New(opts)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return errTransient
	})
	if !errors.Is(err, retry.ErrMaxAttemptsReached) {
		t.Fatalf("expected ErrMaxAttemptsReached, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_RespectsContextCancellation(t *testing.T) {
	opts := retry.Options{
		MaxAttempts:  0, // unlimited
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}
	r := retry.New(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	err := r.Do(ctx, func() error { return errTransient })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}
