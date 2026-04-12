package checkpoint_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/checkpoint"
	"github.com/yourorg/logdrift/internal/diff"
)

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

func TestStream_ForwardsAllEntries(t *testing.T) {
	store, _ := checkpoint.New(tempPath(t))
	entries := []diff.Entry{
		{Service: "api", Level: "info", Message: "started"},
		{Service: "api", Level: "warn", Message: "slow"},
	}
	in := feedEntries(entries)
	out := checkpoint.Stream(context.Background(), in, store)
	got := drain(out)
	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
}

func TestStream_UpdatesCheckpoint(t *testing.T) {
	store, _ := checkpoint.New(tempPath(t))
	entries := []diff.Entry{
		{Service: "worker", Level: "info", Message: "job done"},
	}
	in := feedEntries(entries)
	out := checkpoint.Stream(context.Background(), in, store)
	drain(out)

	offset, err := store.Get("worker")
	if err != nil {
		t.Fatalf("expected checkpoint for worker: %v", err)
	}
	if offset <= 0 {
		t.Fatalf("expected positive offset, got %d", offset)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	store, _ := checkpoint.New(tempPath(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := make(chan diff.Entry)
	out := checkpoint.Stream(ctx, in, store)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected closed channel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
