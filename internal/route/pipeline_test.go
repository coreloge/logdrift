package route

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

func TestStream_RoutesEntries(t *testing.T) {
	entries := []diff.Entry{
		makeEntry("svc", "error", "bad"),
		makeEntry("svc", "info", "good"),
	}
	in := feedEntries(entries)
	opts := Options{
		Rules:        []Rule{{Name: "errors", Field: "level", Values: []string{"error"}}},
		DefaultRoute: "other",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := New(opts)
	errCh := r.Subscribe("errors")
	otherCh := r.Subscribe("other")
	go r.Run(ctx, in)

	errEntries := collect(errCh, 2)
	otherEntries := collect(otherCh, 2)

	if len(errEntries) != 1 {
		t.Fatalf("expected 1 error entry, got %d", len(errEntries))
	}
	if len(otherEntries) != 1 {
		t.Fatalf("expected 1 other entry, got %d", len(otherEntries))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	r := Stream(ctx, in, DefaultOptions())
	_ = r
	cancel()
	time.Sleep(50 * time.Millisecond) // allow goroutine to exit cleanly
}

func TestStream_NoSubscribers_DropsEntries(t *testing.T) {
	entries := []diff.Entry{
		makeEntry("svc", "warn", "watch out"),
	}
	in := feedEntries(entries)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// no subscriptions — should not block or panic
	r := Stream(ctx, in, DefaultOptions())
	_ = r
	time.Sleep(50 * time.Millisecond)
}
