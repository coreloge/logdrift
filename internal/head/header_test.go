package head

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/logdrift/internal/diff"
)

func makeEntry(msg string) diff.Entry {
	return diff.Entry{Service: "svc", Level: "info", Message: msg}
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
	if opts.Max != 10 {
		t.Fatalf("expected Max=10, got %d", opts.Max)
	}
}

func TestNew_NegativeMax_ReturnsError(t *testing.T) {
	_, err := New(Options{Max: -1})
	if err == nil {
		t.Fatal("expected error for Max=-1")
	}
}

func TestNew_ZeroMax_ReturnsError(t *testing.T) {
	_, err := New(Options{Max: 0})
	if err == nil {
		t.Fatal("expected error for Max=0")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(Options{Max: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStream_ForwardsUpToMax(t *testing.T) {
	h, _ := New(Options{Max: 3})
	entries := []diff.Entry{
		makeEntry("a"), makeEntry("b"), makeEntry("c"), makeEntry("d"), makeEntry("e"),
	}
	out := h.Stream(context.Background(), feedEntries(entries))
	got := drain(out)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[0].Message != "a" || got[1].Message != "b" || got[2].Message != "c" {
		t.Fatalf("unexpected messages: %v", got)
	}
}

func TestStream_FewerThanMax_ForwardsAll(t *testing.T) {
	h, _ := New(Options{Max: 10})
	entries := []diff.Entry{makeEntry("x"), makeEntry("y")}
	out := h.Stream(context.Background(), feedEntries(entries))
	got := drain(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	h, _ := New(Options{Max: 100})
	// Unbuffered channel — producer never sends.
	in := make(chan diff.Entry)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	out := h.Stream(ctx, in)
	got := drain(out)
	if len(got) != 0 {
		t.Fatalf("expected 0 entries after cancel, got %d", len(got))
	}
}

func TestStream_ExactlyMax_ForwardsAll(t *testing.T) {
	h, _ := New(Options{Max: 3})
	entries := []diff.Entry{makeEntry("a"), makeEntry("b"), makeEntry("c")}
	out := h.Stream(context.Background(), feedEntries(entries))
	got := drain(out)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
}
