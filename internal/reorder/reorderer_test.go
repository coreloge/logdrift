package reorder_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/reorder"
)

func makeEntry(service, msg string, ts time.Time) diff.Entry {
	return diff.Entry{
		Service:   service,
		Message:   msg,
		Level:     "info",
		Timestamp: ts,
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

func drainStream(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestDefaultOptions(t *testing.T) {
	opts := reorder.DefaultOptions()
	if opts.HoldWindow <= 0 {
		t.Fatalf("expected positive HoldWindow, got %v", opts.HoldWindow)
	}
	if opts.MaxBuffer <= 0 {
		t.Fatalf("expected positive MaxBuffer, got %d", opts.MaxBuffer)
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := reorder.New(reorder.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStream_EmitsInTimestampOrder(t *testing.T) {
	now := time.Now()
	entries := []diff.Entry{
		makeEntry("svc", "third", now.Add(2*time.Second)),
		makeEntry("svc", "first", now),
		makeEntry("svc", "second", now.Add(time.Second)),
	}

	r, _ := reorder.New(reorder.Options{HoldWindow: 20 * time.Millisecond, MaxBuffer: 512})
	ctx := context.Background()
	out := r.Stream(ctx, feedEntries(entries))
	got := drainStream(out)

	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	expected := []string{"first", "second", "third"}
	for i, e := range got {
		if e.Message != expected[i] {
			t.Errorf("pos %d: want %q, got %q", i, expected[i], e.Message)
		}
	}
}

func TestStream_FlushesOnMaxBuffer(t *testing.T) {
	now := time.Now()
	entries := make([]diff.Entry, 5)
	for i := range entries {
		entries[i] = makeEntry("svc", "msg", now.Add(time.Duration(4-i)*time.Second))
	}

	r, _ := reorder.New(reorder.Options{HoldWindow: 10 * time.Second, MaxBuffer: 5})
	ctx := context.Background()
	out := r.Stream(ctx, feedEntries(entries))
	got := drainStream(out)

	if len(got) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ch := make(chan diff.Entry)
	r, _ := reorder.New(reorder.Options{HoldWindow: 50 * time.Millisecond, MaxBuffer: 16})
	ctx, cancel := context.WithCancel(context.Background())
	out := r.Stream(ctx, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for stream to stop")
	}
}
