package suppress

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func feedEntries(entries []diff.LogEntry) <-chan diff.LogEntry {
	ch := make(chan diff.LogEntry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drainStream(ch <-chan diff.LogEntry) []diff.LogEntry {
	var out []diff.LogEntry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestStream_ForwardsAllEntries_NoCooldownHit(t *testing.T) {
	s, _ := New(Options{Field: "message", Cooldown: 10 * time.Second})
	entries := []diff.LogEntry{
		makeEntry("a", "msg1", "info"),
		makeEntry("b", "msg2", "warn"),
	}
	out := Stream(context.Background(), s, feedEntries(entries))
	got := drainStream(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestStream_SuppressesDuplicates(t *testing.T) {
	s, _ := New(Options{Field: "message", Cooldown: 10 * time.Second})
	e := makeEntry("svc", "repeated", "error")
	entries := []diff.LogEntry{e, e, e}
	out := Stream(context.Background(), s, feedEntries(entries))
	got := drainStream(out)
	if len(got) != 1 {
		t.Fatalf("expected 1 forwarded entry, got %d", len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	s, _ := New(DefaultOptions())
	ch := make(chan diff.LogEntry)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, s, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after context cancel")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for channel close")
	}
}
