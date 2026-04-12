package ratelimit_test

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/ratelimit"
)

func makeEntry(service string) diff.Entry {
	return diff.Entry{Service: service, Level: "info", Message: "test"}
}

func TestDefaultOptions_NoLimit(t *testing.T) {
	opts := ratelimit.DefaultOptions()
	if opts.MaxPerSecond != 0 {
		t.Fatalf("expected 0, got %d", opts.MaxPerSecond)
	}
}

func TestAllow_ZeroRate_PassesAll(t *testing.T) {
	th := ratelimit.New(ratelimit.Options{MaxPerSecond: 0})
	for i := 0; i < 100; i++ {
		if !th.Allow(makeEntry("svc-a")) {
			t.Fatal("expected entry to be allowed with no limit")
		}
	}
}

func TestAllow_LimitEnforced(t *testing.T) {
	th := ratelimit.New(ratelimit.Options{MaxPerSecond: 3})
	allowed := 0
	for i := 0; i < 10; i++ {
		if th.Allow(makeEntry("svc-b")) {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("expected 3 allowed, got %d", allowed)
	}
}

func TestAllow_PerServiceIsolation(t *testing.T) {
	th := ratelimit.New(ratelimit.Options{MaxPerSecond: 2})
	for i := 0; i < 2; i++ {
		if !th.Allow(makeEntry("svc-x")) {
			t.Fatalf("svc-x entry %d should be allowed", i)
		}
	}
	// svc-x exhausted; svc-y should still have its own bucket
	if !th.Allow(makeEntry("svc-y")) {
		t.Fatal("svc-y should be allowed independently")
	}
}

func TestAllow_ResetsAfterOneSecond(t *testing.T) {
	th := ratelimit.New(ratelimit.Options{MaxPerSecond: 1})
	if !th.Allow(makeEntry("svc-c")) {
		t.Fatal("first entry should be allowed")
	}
	if th.Allow(makeEntry("svc-c")) {
		t.Fatal("second entry within same second should be rejected")
	}
	time.Sleep(1010 * time.Millisecond)
	if !th.Allow(makeEntry("svc-c")) {
		t.Fatal("entry after reset window should be allowed")
	}
}

func TestApply_FiltersChannel(t *testing.T) {
	th := ratelimit.New(ratelimit.Options{MaxPerSecond: 2})
	in := make(chan diff.Entry, 6)
	for i := 0; i < 6; i++ {
		in <- makeEntry("svc-d")
	}
	close(in)

	out := th.Apply(in)
	count := 0
	for range out {
		count++
	}
	if count != 2 {
		t.Fatalf("expected 2 entries through channel, got %d", count)
	}
}
