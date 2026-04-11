package snapshot_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/snapshot"
)

func feedStream(entries []snapshot.StreamEntry) <-chan snapshot.StreamEntry {
	ch := make(chan snapshot.StreamEntry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func TestCollector_AccumulatesEntries(t *testing.T) {
	c := snapshot.NewCollector()
	ch := feedStream([]snapshot.StreamEntry{
		{Service: "api", Entry: diff.Entry{Level: "info", Message: "started"}},
		{Service: "api", Entry: diff.Entry{Level: "warn", Message: "slow"}},
		{Service: "db", Entry: diff.Entry{Level: "error", Message: "timeout"}},
	})

	c.Collect(context.Background(), ch)
	snap := c.Snapshot()

	if len(snap.Get("api")) != 2 {
		t.Errorf("expected 2 api entries, got %d", len(snap.Get("api")))
	}
	if len(snap.Get("db")) != 1 {
		t.Errorf("expected 1 db entry, got %d", len(snap.Get("db")))
	}
}

func TestCollector_StopsOnContextCancel(t *testing.T) {
	c := snapshot.NewCollector()
	ch := make(chan snapshot.StreamEntry)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		c.Collect(ctx, ch)
		close(done)
	}()

	select {
	case <-done:
		// expected
	case <-time.After(200 * time.Millisecond):
		t.Error("Collect did not stop after context cancellation")
	}
}

func TestCollector_EmptyChannel(t *testing.T) {
	c := snapshot.NewCollector()
	ch := feedStream(nil)
	c.Collect(context.Background(), ch)
	snap := c.Snapshot()
	if len(snap.Services()) != 0 {
		t.Errorf("expected empty snapshot, got %d services", len(snap.Services()))
	}
}
