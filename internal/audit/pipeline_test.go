package audit

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

func feedResults(entries []DriftResult) <-chan DriftResult {
	ch := make(chan DriftResult, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drainEntries(ch <-chan Entry) []Entry {
	var out []Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	a := New(DefaultOptions())
	input := []DriftResult{
		{Entry: Entry{Service: "svc", Level: "info", Message: "hello"}},
		{Entry: Entry{Service: "svc", Level: "warn", Message: "world"}},
	}
	ctx := context.Background()
	out := Stream(ctx, a, feedResults(input))
	entries := drainEntries(out)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestStream_RecordsDriftWhenDeltent(t *testing.T) {
	a := New(DefaultOptions())
	input := []DriftResult{
		{
			Entry:  Entry{Service: "svc", Level: "info", Message: "msg"},
			Deltas: []diff.Delta{{Field: "level", Got: "info", Want: "error"}},
		},
	}
	ctx := context.Background()
	out := Stream(ctx, a, feedResults(input))
	drainEntries(out)
	records := a.All()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if !records[0].Drift {
		t.Error("expected Drift=true")
	}
}

func TestStream_RecordsEntryWhenNoDelta(t *testing.T) {
	a := New(DefaultOptions())
	input := []DriftResult{
		{Entry: Entry{Service: "svc", Level: "debug", Message: "ok"}},
	}
	ctx := context.Background()
	out := Stream(ctx, a, feedResults(input))
	drainEntries(out)
	records := a.All()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Drift {
		t.Error("expected Drift=false")
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	a := New(DefaultOptions())
	ch := make(chan DriftResult)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, a, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for stream to stop")
	}
}
