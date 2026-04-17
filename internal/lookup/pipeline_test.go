package lookup

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
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
	l, _ := New(DefaultOptions())
	entries := []diff.Entry{
		makeEntry("svc", "info", "a"),
		makeEntry("svc", "warn", "b"),
	}
	out := drainStream(Stream(context.Background(), l, feedEntries(entries)))
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestStream_ResolvesLookup(t *testing.T) {
	opts := DefaultOptions()
	opts.Table = map[string]string{"auth": "platform"}
	l, _ := New(opts)

	in := feedEntries([]diff.Entry{makeEntry("auth", "info", "ok")})
	out := drainStream(Stream(context.Background(), l, in))

	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].Extra["team"] != "platform" {
		t.Fatalf("expected team=platform, got %q", out[0].Extra["team"])
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	l, _ := New(DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	blocking := make(chan diff.Entry)
	done := make(chan struct{})
	go func() {
		drainStream(Stream(ctx, l, blocking))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stream did not stop after context cancel")
	}
}
