package severity_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/severity"
)

func feedEntries(entries []diff.Entry) <-chan diff.Entry {
	ch := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drainFiltered(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestFilter_DefaultOptions_PassesAll(t *testing.T) {
	entries := []diff.Entry{
		{Service: "svc", Level: "debug", Message: "a"},
		{Service: "svc", Level: "info", Message: "b"},
		{Service: "svc", Level: "error", Message: "c"},
	}
	ch := feedEntries(entries)
	out := severity.Filter(context.Background(), ch, severity.DefaultFilterOptions())
	got := drainFiltered(out)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
}

func TestFilter_DropsEntriesBelowMinLevel(t *testing.T) {
	entries := []diff.Entry{
		{Service: "svc", Level: "debug", Message: "dropped"},
		{Service: "svc", Level: "info", Message: "dropped"},
		{Service: "svc", Level: "warn", Message: "kept"},
		{Service: "svc", Level: "error", Message: "kept"},
	}
	opts := severity.FilterOptions{MinLevel: severity.Warn}
	ch := feedEntries(entries)
	out := severity.Filter(context.Background(), ch, opts)
	got := drainFiltered(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	for _, e := range got {
		if e.Message == "dropped" {
			t.Fatalf("entry with message %q should have been dropped", e.Message)
		}
	}
}

func TestFilter_StopsOnContextCancel(t *testing.T) {
	ch := make(chan diff.Entry) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := severity.Filter(ctx, ch, severity.DefaultFilterOptions())
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
