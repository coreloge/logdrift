package pin

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

func TestStream_ForwardsAllEntries(t *testing.T) {
	p, _ := New(DefaultOptions())
	entries := []diff.Entry{
		{Service: "a", Message: "one", Level: "info", Timestamp: time.Now()},
		{Service: "b", Message: "two", Level: "warn", Timestamp: time.Now()},
	}
	out := Stream(context.Background(), p, feedEntries(entries))
	result := drainStream(out)
	if len(result) != 2 {
		t.Fatalf("expected 2 forwarded entries, got %d", len(result))
	}
}

func TestStream_PinsEntries(t *testing.T) {
	p, _ := New(DefaultOptions())
	entries := []diff.Entry{
		{Service: "svc", Message: "hello", Level: "info", Timestamp: time.Now()},
	}
	out := Stream(context.Background(), p, feedEntries(entries))
	drainStream(out)
	if len(p.All()) != 1 {
		t.Fatalf("expected 1 pinned entry, got %d", len(p.All()))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	p, _ := New(DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch := make(chan diff.Entry)
	out := Stream(ctx, p, ch)
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after context cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
