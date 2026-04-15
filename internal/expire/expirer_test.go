package expire_test

import (
	"context"
	"testing"
	"time"

	"github.com/robmorgan/logdrift/internal/diff"
	"github.com/robmorgan/logdrift/internal/expire"
)

func makeEntry(service, msg string, ts time.Time) diff.Entry {
	return diff.Entry{
		Service:   service,
		Message:   msg,
		Level:     "info",
		Timestamp: ts,
	}
}

func feedEntries(entries []diff.Entry) <-chan diff.Entry {
	ch := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drainStream(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestNew_ZeroTTL_ReturnsError(t *testing.T) {
	_, err := expire.New(expire.Options{TTL: 0})
	if err == nil {
		t.Fatal("expected error for zero TTL")
	}
}

func TestNew_NegativeTTL_ReturnsError(t *testing.T) {
	_, err := expire.New(expire.Options{TTL: -time.Second})
	if err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestAllow_FreshEntry_ReturnsTrue(t *testing.T) {
	now := time.Now()
	e, _ := expire.New(expire.Options{TTL: time.Minute, Now: func() time.Time { return now }})
	entry := makeEntry("svc", "hello", now.Add(-10*time.Second))
	if !e.Allow(entry) {
		t.Error("expected fresh entry to be allowed")
	}
}

func TestAllow_StaleEntry_ReturnsFalse(t *testing.T) {
	now := time.Now()
	e, _ := expire.New(expire.Options{TTL: time.Minute, Now: func() time.Time { return now }})
	entry := makeEntry("svc", "old", now.Add(-2*time.Minute))
	if e.Allow(entry) {
		t.Error("expected stale entry to be rejected")
	}
}

func TestAllow_ZeroTimestamp_PassesThrough(t *testing.T) {
	e, _ := expire.New(expire.DefaultOptions())
	entry := makeEntry("svc", "no-ts", time.Time{})
	if !e.Allow(entry) {
		t.Error("expected zero-timestamp entry to pass through")
	}
}

func TestStream_DropsExpiredEntries(t *testing.T) {
	now := time.Now()
	e, _ := expire.New(expire.Options{TTL: time.Minute, Now: func() time.Time { return now }})

	entries := []diff.Entry{
		makeEntry("svc", "fresh", now.Add(-10*time.Second)),
		makeEntry("svc", "stale", now.Add(-5*time.Minute)),
		makeEntry("svc", "also-fresh", now.Add(-30*time.Second)),
	}

	out := expire.Stream(context.Background(), e, feedEntries(entries))
	got := drainStream(out)

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Message != "fresh" || got[1].Message != "also-fresh" {
		t.Errorf("unexpected messages: %v", got)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	e, _ := expire.New(expire.DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := make(chan diff.Entry)
	close(in)
	out := expire.Stream(ctx, e, in)
	drainStream(out) // must not block
}
