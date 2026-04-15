package merge_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/merge"
)

func makeEntry(service, msg string, ts time.Time) diff.Entry {
	return diff.Entry{
		Service:   service,
		Message:   msg,
		Level:     "info",
		Timestamp: ts,
	}
}

func feedChan(entries []diff.Entry) <-chan diff.Entry {
	ch := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drainAll(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestMerger_SingleSource(t *testing.T) {
	now := time.Now()
	entries := []diff.Entry{
		makeEntry("svc-a", "first", now),
		makeEntry("svc-a", "second", now.Add(time.Second)),
	}
	m := merge.New(merge.DefaultOptions())
	out := m.Stream(context.Background(), []<-chan diff.Entry{feedChan(entries)})
	got := drainAll(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Message != "first" || got[1].Message != "second" {
		t.Errorf("unexpected order: %v", got)
	}
}

func TestMerger_MultipleSources_ChronologicalOrder(t *testing.T) {
	now := time.Now()
	ch1 := feedChan([]diff.Entry{
		makeEntry("a", "a1", now.Add(0)),
		makeEntry("a", "a3", now.Add(2*time.Second)),
	})
	ch2 := feedChan([]diff.Entry{
		makeEntry("b", "b2", now.Add(1*time.Second)),
		makeEntry("b", "b4", now.Add(3*time.Second)),
	})
	m := merge.New(merge.DefaultOptions())
	out := m.Stream(context.Background(), []<-chan diff.Entry{ch1, ch2})
	got := drainAll(out)
	expected := []string{"a1", "b2", "a3", "b4"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(got))
	}
	for i, e := range got {
		if e.Message != expected[i] {
			t.Errorf("pos %d: expected %q, got %q", i, expected[i], e.Message)
		}
	}
}

func TestMerger_EmptySources(t *testing.T) {
	m := merge.New(merge.DefaultOptions())
	out := m.Stream(context.Background(), []<-chan diff.Entry{})
	got := drainAll(out)
	if len(got) != 0 {
		t.Errorf("expected 0 entries, got %d", len(got))
	}
}

func TestMerger_StopsOnContextCancel(t *testing.T) {
	now := time.Now()
	// Unbuffered channel that never closes — simulates a live stream.
	blocking := make(chan diff.Entry)
	defer close(blocking)
	seeded := feedChan([]diff.Entry{makeEntry("svc", "msg", now)})
	ctx, cancel := context.WithCancel(context.Background())
	m := merge.New(merge.DefaultOptions())
	out := m.Stream(ctx, []<-chan diff.Entry{seeded, blocking})
	// Drain the one seeded entry then cancel.
	<-out
	cancel()
	// Drain remaining; should terminate promptly.
	for range out {
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := merge.DefaultOptions()
	if opts.BufferSize <= 0 {
		t.Errorf("expected positive BufferSize, got %d", opts.BufferSize)
	}
}
