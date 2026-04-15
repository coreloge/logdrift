package jitter

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{Service: svc, Message: msg, Level: "info"}
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
	opts := DefaultOptions()
	if opts.MinDelay != 0 {
		t.Errorf("expected MinDelay 0, got %v", opts.MinDelay)
	}
	if opts.MaxDelay != 50*time.Millisecond {
		t.Errorf("expected MaxDelay 50ms, got %v", opts.MaxDelay)
	}
}

func TestNew_InvalidRange_ReturnsError(t *testing.T) {
	opts := Options{MinDelay: 100 * time.Millisecond, MaxDelay: 10 * time.Millisecond}
	_, err := New(opts)
	if err == nil {
		t.Fatal("expected error for MaxDelay < MinDelay")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	opts := DefaultOptions()
	_, err := New(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	entries := []diff.Entry{
		makeEntry("svc-a", "hello"),
		makeEntry("svc-b", "world"),
	}
	opts := Options{
		MinDelay: 0,
		MaxDelay: 0,
		Rand:     rand.New(rand.NewSource(42)),
	}
	ctx := context.Background()
	out, err := Stream(ctx, feedEntries(entries), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := drainStream(out)
	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
	for i, e := range got {
		if e.Message != entries[i].Message {
			t.Errorf("entry %d: expected message %q, got %q", i, entries[i].Message, e.Message)
		}
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	blocking := make(chan diff.Entry) // never closed
	opts := Options{MinDelay: 0, MaxDelay: 0, Rand: rand.New(rand.NewSource(1))}
	ctx, cancel := context.WithCancel(context.Background())
	out, err := Stream(ctx, blocking, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timed out waiting for channel close")
	}
}
