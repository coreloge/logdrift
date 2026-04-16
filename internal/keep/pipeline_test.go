package keep

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

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

func TestStream_PassesAllWhenNoRules(t *testing.T) {
	k, _ := New(DefaultOptions())
	in := feedEntries([]diff.Entry{
		makeEntry("a", "info", "hello"),
		makeEntry("b", "warn", "world"),
	})
	out := drainStream(Stream(context.Background(), in, k))
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestStream_KeepsMatchingEntries(t *testing.T) {
	k, _ := New(Options{Rules: []Rule{{Pattern: "keep"}}})
	in := feedEntries([]diff.Entry{
		makeEntry("svc", "info", "please keep this"),
		makeEntry("svc", "info", "drop this one"),
		makeEntry("svc", "error", "keep me too"),
	})
	out := drainStream(Stream(context.Background(), in, k))
	if len(out) != 2 {
		t.Fatalf("expected 2 kept entries, got %d", len(out))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	k, _ := New(Options{Rules: []Rule{{Pattern: "."}}})
	ch := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, ch, k)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
