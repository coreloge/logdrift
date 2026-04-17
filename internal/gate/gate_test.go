package gate_test

import (
	"context"
	"testing"
	"time"

	"github.com/humanlogio/logdrift/internal/diff"
	"github.com/humanlogio/logdrift/internal/gate"
)

func makeEntry(level, msg string) diff.Entry {
	return diff.Entry{Level: level, Message: msg, Service: "svc"}
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

func TestNew_NilPredicate_ReturnsError(t *testing.T) {
	_, err := gate.New(gate.Options{})
	if err == nil {
		t.Fatal("expected error for nil predicate")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := gate.New(gate.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_PassesMatchingEntries(t *testing.T) {
	g, _ := gate.New(gate.Options{
		Predicate: func(e diff.Entry) bool { return e.Level == "error" },
	})
	if !g.Allow(makeEntry("error", "boom")) {
		t.Error("expected error entry to pass")
	}
	if g.Allow(makeEntry("info", "ok")) {
		t.Error("expected info entry to be blocked")
	}
}

func TestAllow_Invert_FlipsBehaviour(t *testing.T) {
	g, _ := gate.New(gate.Options{
		Predicate: func(e diff.Entry) bool { return e.Level == "error" },
		Invert:    true,
	})
	if g.Allow(makeEntry("error", "boom")) {
		t.Error("expected error entry to be blocked when inverted")
	}
	if !g.Allow(makeEntry("info", "ok")) {
		t.Error("expected info entry to pass when inverted")
	}
}

func TestStream_ForwardsMatchingEntries(t *testing.T) {
	g, _ := gate.New(gate.Options{
		Predicate: func(e diff.Entry) bool { return e.Level == "error" },
	})
	entries := []diff.Entry{
		makeEntry("info", "a"),
		makeEntry("error", "b"),
		makeEntry("warn", "c"),
		makeEntry("error", "d"),
	}
	out := drain(gate.Stream(context.Background(), g, feedEntries(entries)))
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	g, _ := gate.New(gate.DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch := make(chan diff.Entry)
	close(ch)
	resultCh := gate.Stream(ctx, g, ch)
	select {
	case <-resultCh:
	case <-time.After(time.Second):
		t.Fatal("stream did not close after context cancel")
	}
}
