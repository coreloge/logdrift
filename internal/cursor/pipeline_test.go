package cursor

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

func drainStream(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	entries := []diff.Entry{
		{Service: "api", Message: "started"},
		{Service: "api", Message: "ready"},
	}
	store := New()
	out := Stream(context.Background(), store, feedEntries(entries))
	got := drainStream(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestStream_UpdatesCursor(t *testing.T) {
	entries := []diff.Entry{
		{Service: "db", Message: "ping"},
		{Service: "db", Message: "pong"},
	}
	store := New()
	out := Stream(context.Background(), store, feedEntries(entries))
	drainStream(out)

	p, err := store.Get("db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Line != 2 {
		t.Fatalf("expected line=2, got %d", p.Line)
	}
	expectedOffset := int64(len("ping")+1) + int64(len("pong")+1)
	if p.Offset != expectedOffset {
		t.Fatalf("expected offset=%d, got %d", expectedOffset, p.Offset)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	blocking := make(chan diff.Entry)
	store := New()
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, store, blocking)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected closed channel after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
