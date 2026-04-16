package drop

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
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
	d, _ := New(DefaultOptions())
	entries := []diff.Entry{
		{Service: "a", Level: "info", Message: "one", Timestamp: time.Now()},
		{Service: "b", Level: "debug", Message: "two", Timestamp: time.Now()},
	}
	out := drainStream(Stream(context.Background(), feedEntries(entries), d))
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestStream_DropsMatchingEntries(t *testing.T) {
	d, _ := New(Options{Rules: []Rule{{Field: "level", Pattern: "^debug$"}}})
	entries := []diff.Entry{
		{Service: "a", Level: "debug", Message: "verbose", Timestamp: time.Now()},
		{Service: "a", Level: "info", Message: "keep", Timestamp: time.Now()},
	}
	out := drainStream(Stream(context.Background(), feedEntries(entries), d))
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].Level != "info" {
		t.Fatalf("expected info entry, got %s", out[0].Level)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	d, _ := New(DefaultOptions())
	ch := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, ch, d)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to close after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
