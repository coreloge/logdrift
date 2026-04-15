package sample

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{
		Service: svc,
		Message: msg,
		Level:   "info",
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

func drain(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Size != 100 {
		t.Fatalf("expected default size 100, got %d", opts.Size)
	}
}

func TestNew_InvalidSize_ReturnsError(t *testing.T) {
	_, err := New(Options{Size: 0})
	if err == nil {
		t.Fatal("expected error for size 0")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(Options{Size: 10, Seed: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdd_FillsUpToSize(t *testing.T) {
	r, _ := New(Options{Size: 3, Seed: 1})
	for i := 0; i < 5; i++ {
		r.Add(makeEntry("svc", "msg"))
	}
	if r.Len() != 3 {
		t.Fatalf("expected len 3, got %d", r.Len())
	}
}

func TestSnapshot_ReturnsCopy(t *testing.T) {
	r, _ := New(Options{Size: 5, Seed: 1})
	r.Add(makeEntry("a", "hello"))
	snap := r.Snapshot()
	snap[0].Message = "mutated"
	snap2 := r.Snapshot()
	if snap2[0].Message == "mutated" {
		t.Fatal("snapshot should be a copy")
	}
}

func TestReset_ClearsEntries(t *testing.T) {
	r, _ := New(Options{Size: 5, Seed: 1})
	r.Add(makeEntry("svc", "msg"))
	r.Reset()
	if r.Len() != 0 {
		t.Fatalf("expected len 0 after reset, got %d", r.Len())
	}
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	r, _ := New(Options{Size: 10, Seed: 1})
	entries := []diff.Entry{
		makeEntry("svc", "a"),
		makeEntry("svc", "b"),
		makeEntry("svc", "c"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out := Stream(ctx, r, feedEntries(entries))
	result := drain(out)
	if len(result) != 3 {
		t.Fatalf("expected 3 forwarded entries, got %d", len(result))
	}
	if r.Len() != 3 {
		t.Fatalf("expected reservoir len 3, got %d", r.Len())
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	r, _ := New(Options{Size: 10, Seed: 1})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := make(chan diff.Entry)
	close(in)
	out := Stream(ctx, r, in)
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
