package debounce

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(svc, level, msg string) diff.Entry {
	return diff.Entry{Service: svc, Level: level, Message: msg}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Window <= 0 {
		t.Fatalf("expected positive window, got %v", opts.Window)
	}
}

func TestAllow_FirstOccurrence_ReturnsTrue(t *testing.T) {
	d := New(DefaultOptions())
	if !d.Allow(makeEntry("svc", "info", "hello")) {
		t.Fatal("expected first occurrence to be allowed")
	}
}

func TestAllow_DuplicateWithinWindow_ReturnsFalse(t *testing.T) {
	opts := Options{Window: 1 * time.Hour}
	d := New(opts)
	e := makeEntry("svc", "info", "hello")
	d.Allow(e)
	if d.Allow(e) {
		t.Fatal("expected duplicate within window to be suppressed")
	}
}

func TestAllow_AfterWindowExpires_ReturnsTrue(t *testing.T) {
	opts := Options{Window: 50 * time.Millisecond}
	d := New(opts)
	now := time.Now()
	d.nowFn = func() time.Time { return now }

	e := makeEntry("svc", "warn", "burst")
	d.Allow(e)

	// advance clock past the window
	d.nowFn = func() time.Time { return now.Add(100 * time.Millisecond) }
	if !d.Allow(e) {
		t.Fatal("expected entry to be allowed after window expires")
	}
}

func TestAllow_DifferentMessages_BothAllowed(t *testing.T) {
	d := New(Options{Window: 1 * time.Hour})
	if !d.Allow(makeEntry("svc", "info", "msg-a")) {
		t.Fatal("expected msg-a to be allowed")
	}
	if !d.Allow(makeEntry("svc", "info", "msg-b")) {
		t.Fatal("expected msg-b to be allowed")
	}
}

func TestReset_ClearsCache(t *testing.T) {
	d := New(Options{Window: 1 * time.Hour})
	e := makeEntry("svc", "error", "boom")
	d.Allow(e)
	d.Reset()
	if !d.Allow(e) {
		t.Fatal("expected entry to be allowed after reset")
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

func TestStream_SuppressesDuplicates(t *testing.T) {
	ctx := context.Background()
	e := makeEntry("svc", "info", "repeated")
	in := feedEntries([]diff.Entry{e, e, e})
	out := Stream(ctx, in, Options{Window: 1 * time.Hour})

	var []diff.Entry
	for entry := range out {
		got = append(got, entry)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := make(chan diff.Entry)
	close(in)
	out := Stream(ctx, in, DefaultOptions())
	// channel must close without blocking
	for range out {
	}
}
