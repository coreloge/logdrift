package window_test

import (
	"context"
	"testing"
	"time"

	"logdrift/internal/diff"
	"logdrift/internal/window"
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
	w := window.New(window.DefaultOptions())
	now := time.Now()
	entries := []diff.Entry{
		makeEntry("svc-a", now),
		makeEntry("svc-b", now.Add(time.Second)),
	}

	out := window.Stream(context.Background(), w, feedEntries(entries))
	got := drainStream(out)

	if len(got) != len(entries) {
		t.Fatalf("expected %d entries forwarded, got %d", len(entries), len(got))
	}
}

func TestStream_PopulatesWindow(t *testing.T) {
	w := window.New(window.Options{Width: time.Minute, MaxBuckets: 5})
	now := time.Now().Truncate(time.Minute)
	entries := []diff.Entry{
		makeEntry("svc-a", now),
		makeEntry("svc-b", now.Add(10*time.Second)),
	}

	out := window.Stream(context.Background(), w, feedEntries(entries))
	drainStream(out)

	if len(w.Buckets()) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(w.Buckets()))
	}
	if len(w.Buckets()[0].Entries) != 2 {
		t.Fatalf("expected 2 entries in bucket, got %d", len(w.Buckets()[0].Entries))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	w := window.New(window.DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan diff.Entry)
	out := window.Stream(ctx, w, in)

	cancel()
	// channel should close without deadlock
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
