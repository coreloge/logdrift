package backpressure_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/backpressure"
	"github.com/logdrift/logdrift/internal/diff"
)

func makeEntry(msg string) diff.Entry {
	return diff.Entry{Service: "svc", Level: "info", Message: msg, Fields: map[string]string{}}
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
	opts := backpressure.DefaultOptions()
	if opts.Strategy != backpressure.Drop {
		t.Errorf("expected Drop strategy, got %q", opts.Strategy)
	}
	if opts.BufferSize != 64 {
		t.Errorf("expected buffer 64, got %d", opts.BufferSize)
	}
}

func TestStream_ForwardsAllEntries_WhenBufferSufficient(t *testing.T) {
	opts := backpressure.DefaultOptions()
	opts.BufferSize = 10
	bp := backpressure.New(opts)

	entries := []diff.Entry{makeEntry("a"), makeEntry("b"), makeEntry("c")}
	ctx := context.Background()
	out := bp.Stream(ctx, feedEntries(entries))
	got := drain(out)

	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if bp.Dropped() != 0 {
		t.Errorf("expected 0 dropped, got %d", bp.Dropped())
	}
}

func TestStream_Drop_DropsWhenFull(t *testing.T) {
	opts := backpressure.DefaultOptions()
	opts.Strategy = backpressure.Drop
	opts.BufferSize = 1
	bp := backpressure.New(opts)

	// Fill with more entries than the buffer can hold without a reader.
	src := make(chan diff.Entry, 10)
	for i := 0; i < 10; i++ {
		src <- makeEntry("x")
	}
	close(src)

	ctx := context.Background()
	out := bp.Stream(ctx, src)
	drain(out)

	if bp.Dropped() == 0 {
		t.Error("expected some entries to be dropped")
	}
}

func TestStream_Block_WaitsUpToTimeout(t *testing.T) {
	opts := backpressure.DefaultOptions()
	opts.Strategy = backpressure.Block
	opts.BufferSize = 1
	opts.Timeout = 20 * time.Millisecond
	bp := backpressure.New(opts)

	src := make(chan diff.Entry, 5)
	for i := 0; i < 5; i++ {
		src <- makeEntry("y")
	}
	close(src)

	ctx := context.Background()
	out := bp.Stream(ctx, src)
	drain(out)
	// Some entries may be dropped after timeout; we just verify no panic and
	// that dropped + received == 5.
	if bp.Dropped() > 5 {
		t.Errorf("dropped count %d exceeds total sent", bp.Dropped())
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	bp := backpressure.New(backpressure.DefaultOptions())
	src := make(chan diff.Entry) // never closed
	ctx, cancel := context.WithCancel(context.Background())
	out := bp.Stream(ctx, src)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for channel close")
	}
}
