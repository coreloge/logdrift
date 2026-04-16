package group_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
	"github.com/logdrift/logdrift/internal/group"
)

func makeEntry(service, level, msg string) diff.Entry {
	return diff.Entry{
		Timestamp: time.Now(),
		Service:   service,
		Level:     level,
		Message:   msg,
	}
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

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := group.New(group.Options{Field: "", MaxGroups: 8})
	if err == nil {
		t.Fatal("expected error for empty Field")
	}
}

func TestNew_ZeroMaxGroups_ReturnsError(t *testing.T) {
	_, err := group.New(group.Options{Field: "service", MaxGroups: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxGroups")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := group.New(group.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecord_GroupsByService(t *testing.T) {
	g, _ := group.New(group.DefaultOptions())
	g.Record(makeEntry("svc-a", "info", "hello"))
	g.Record(makeEntry("svc-b", "warn", "world"))
	g.Record(makeEntry("svc-a", "error", "boom"))

	if got := len(g.Get("svc-a")); got != 2 {
		t.Fatalf("svc-a: want 2 entries, got %d", got)
	}
	if got := len(g.Get("svc-b")); got != 1 {
		t.Fatalf("svc-b: want 1 entry, got %d", got)
	}
}

func TestGet_MissingKey_ReturnsNil(t *testing.T) {
	g, _ := group.New(group.DefaultOptions())
	if got := g.Get("missing"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestRecord_MaxGroupsCap_DropsExtra(t *testing.T) {
	g, _ := group.New(group.Options{Field: "service", MaxGroups: 2})
	g.Record(makeEntry("a", "info", "1"))
	g.Record(makeEntry("b", "info", "2"))
	g.Record(makeEntry("c", "info", "3")) // should be dropped

	keys := g.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(keys))
	}
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	g, _ := group.New(group.DefaultOptions())
	entries := []diff.Entry{
		makeEntry("svc-a", "info", "msg1"),
		makeEntry("svc-b", "error", "msg2"),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := group.Stream(ctx, g, feedEntries(entries))
	result := drainStream(out)

	if len(result) != len(entries) {
		t.Fatalf("want %d entries forwarded, got %d", len(entries), len(result))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	g, _ := group.New(group.DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := make(chan diff.Entry)
	close(in)
	out := group.Stream(ctx, g, in)
	drainStream(out) // should not block
}
