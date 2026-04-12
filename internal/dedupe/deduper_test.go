package dedupe_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/internal/dedupe"
	"github.com/logdrift/internal/diff"
)

func makeEntry(svc, level, msg string) diff.LogEntry {
	return diff.LogEntry{Service: svc, Level: level, Message: msg}
}

func TestIsDuplicate_FirstSeen_ReturnsFalse(t *testing.T) {
	d := dedupe.New(dedupe.DefaultOptions())
	if d.IsDuplicate(makeEntry("api", "info", "started")) {
		t.Fatal("expected false for first occurrence")
	}
}

func TestIsDuplicate_SecondSeen_WithinWindow_ReturnsTrue(t *testing.T) {
	d := dedupe.New(dedupe.DefaultOptions())
	e := makeEntry("api", "info", "started")
	d.IsDuplicate(e)
	if !d.IsDuplicate(e) {
		t.Fatal("expected duplicate within window")
	}
}

func TestIsDuplicate_AfterWindow_ReturnsFalse(t *testing.T) {
	opts := dedupe.Options{Window: 10 * time.Millisecond, MaxCache: 64}
	d := dedupe.New(opts)
	e := makeEntry("api", "warn", "slow query")
	d.IsDuplicate(e)
	time.Sleep(20 * time.Millisecond)
	if d.IsDuplicate(e) {
		t.Fatal("expected false after window expiry")
	}
}

func TestReset_ClearsCache(t *testing.T) {
	d := dedupe.New(dedupe.DefaultOptions())
	e := makeEntry("svc", "error", "boom")
	d.IsDuplicate(e)
	d.Reset()
	if d.IsDuplicate(e) {
		t.Fatal("expected false after reset")
	}
}

func TestStream_SuppressesDuplicates(t *testing.T) {
	d := dedupe.New(dedupe.DefaultOptions())
	in := make(chan diff.LogEntry, 4)
	e := makeEntry("svc", "info", "hello")
	in <- e
	in <- e
	in <- makeEntry("svc", "info", "world")
	close(in)

	ctx := context.Background()
	out := dedupe.Stream(ctx, in, d)

	var got []diff.LogEntry
	for entry := range out {
		got = append(got, entry)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	d := dedupe.New(dedupe.DefaultOptions())
	in := make(chan diff.LogEntry)
	ctx, cancel := context.WithCancel(context.Background())
	out := dedupe.Stream(ctx, in, d)
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel to be closed after context cancel")
	}
}
