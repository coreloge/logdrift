package counter

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

func TestStream_ForwardsAllEntries(t *testing.T) {
	c, _ := New(DefaultOptions())
	entries := []diff.Entry{
		makeEntry("svc", "info", "a"),
		makeEntry("svc", "warn", "b"),
		makeEntry("svc", "error", "c"),
	}

	out := drainStream(Stream(context.Background(), c, feedEntries(entries)))
	if len(out) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(out))
	}
}

func TestStream_AnnotatesCount(t *testing.T) {
	c, _ := New(DefaultOptions())
	entries := []diff.Entry{
		makeEntry("svc", "error", "first"),
		makeEntry("svc", "error", "second"),
	}

	out := drainStream(Stream(context.Background(), c, feedEntries(entries)))
	if out[0].Extra["_count"] != "1" {
		t.Fatalf("expected first _count=1, got %q", out[0].Extra["_count"])
	}
	if out[1].Extra["_count"] != "2" {
		t.Fatalf("expected second _count=2, got %q", out[1].Extra["_count"])
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	c, _ := New(DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())

	blocking := make(chan diff.Entry)
	outCh := Stream(ctx, c, blocking)

	cancel()

	select {
	case _, ok := <-outCh:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
