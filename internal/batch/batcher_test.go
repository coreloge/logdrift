package batch_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/batch"
	"github.com/user/logdrift/internal/diff"
)

func makeEntry(svc, msg string) diff.LogEntry {
	return diff.LogEntry{Service: svc, Message: msg, Level: "info"}
}

func feedEntries(entries []diff.LogEntry) <-chan diff.LogEntry {
	ch := make(chan diff.LogEntry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func TestDefaultOptions(t *testing.T) {
	opts := batch.DefaultOptions()
	if opts.MaxSize <= 0 {
		t.Fatalf("expected positive MaxSize, got %d", opts.MaxSize)
	}
	if opts.FlushInterval <= 0 {
		t.Fatalf("expected positive FlushInterval, got %v", opts.FlushInterval)
	}
}

func TestBatcher_FlushesOnChannelClose(t *testing.T) {
	entries := []diff.LogEntry{
		makeEntry("svc-a", "hello"),
		makeEntry("svc-b", "world"),
	}
	in := feedEntries(entries)
	b := batch.New(batch.Options{MaxSize: 100, FlushInterval: 5 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	batches := b.Stream(ctx, in)
	var got []diff.LogEntry
	for batch := range batches {
		got = append(got, batch...)
	}
	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
}

func TestBatcher_FlushesOnMaxSize(t *testing.T) {
	const total = 10
	entries := make([]diff.LogEntry, total)
	for i := range entries {
		entries[i] = makeEntry("svc", "msg")
	}
	in := feedEntries(entries)
	b := batch.New(batch.Options{MaxSize: 3, FlushInterval: 5 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	for batch := range b.Stream(ctx, in) {
		count += len(batch)
	}
	if count != total {
		t.Fatalf("expected %d total entries across batches, got %d", total, count)
	}
}

func TestBatcher_StopsOnContextCancel(t *testing.T) {
	ch := make(chan diff.LogEntry)
	defer close(ch)
	b := batch.New(batch.Options{MaxSize: 10, FlushInterval: 50 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	out := b.Stream(ctx, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after context cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for batcher to stop")
	}
}

func TestBatcher_IntervalFlush(t *testing.T) {
	ch := make(chan diff.LogEntry, 1)
	ch <- makeEntry("svc", "tick")
	b := batch.New(batch.Options{MaxSize: 100, FlushInterval: 50 * time.Millisecond})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out := b.Stream(ctx, ch)
	select {
	case batch, ok := <-out:
		if !ok || len(batch) != 1 {
			t.Fatalf("expected 1-entry batch, got ok=%v len=%d", ok, len(batch))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for interval flush")
	}
}
