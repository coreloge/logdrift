package limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/internal/diff"
	"github.com/logdrift/internal/limiter"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{Service: svc, Message: msg, Level: "info"}
}

func TestDefaultOptions_NoLimit(t *testing.T) {
	opts := limiter.DefaultOptions()
	if opts.Rate != 0 {
		t.Fatalf("expected rate 0, got %d", opts.Rate)
	}
}

func TestAllow_ZeroRate_AlwaysTrue(t *testing.T) {
	l := limiter.New(limiter.Options{Rate: 0})
	for i := 0; i < 100; i++ {
		if !l.Allow(makeEntry("svc", "msg")) {
			t.Fatal("expected Allow to return true for zero rate")
		}
	}
}

func TestAllow_LimitEnforced(t *testing.T) {
	// burst of 3 means first 3 calls succeed immediately
	l := limiter.New(limiter.Options{Rate: 3, BurstSize: 3})
	allowed := 0
	for i := 0; i < 10; i++ {
		if l.Allow(makeEntry("svc", "msg")) {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("expected 3 allowed, got %d", allowed)
	}
}

func TestAllow_TokensRefillOverTime(t *testing.T) {
	l := limiter.New(limiter.Options{Rate: 100, BurstSize: 1})
	// consume the single token
	if !l.Allow(makeEntry("svc", "a")) {
		t.Fatal("first call should be allowed")
	}
	if l.Allow(makeEntry("svc", "b")) {
		t.Fatal("second immediate call should be denied")
	}
	// wait long enough for at least one token to refill (rate=100/s)
	time.Sleep(20 * time.Millisecond)
	if !l.Allow(makeEntry("svc", "c")) {
		t.Fatal("call after refill period should be allowed")
	}
}

func TestStream_ForwardsAll_WhenUnlimited(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan diff.Entry, 5)
	for i := 0; i < 5; i++ {
		in <- makeEntry("svc", "msg")
	}
	close(in)

	out := limiter.Stream(ctx, in, limiter.DefaultOptions())
	count := 0
	for range out {
		count++
	}
	if count != 5 {
		t.Fatalf("expected 5 entries, got %d", count)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan diff.Entry)
	out := limiter.Stream(ctx, in, limiter.DefaultOptions())
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after context cancel")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for stream to stop")
	}
}
