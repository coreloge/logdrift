package split_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/split"
)

func makeEntry(service, level, msg string) diff.Entry {
	return diff.Entry{
		Service: service,
		Level:   level,
		Message: msg,
		Fields:  map[string]string{},
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

func drain(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestDefaultOptions(t *testing.T) {
	opts := split.DefaultOptions()
	if opts.BufferSize <= 0 {
		t.Fatalf("expected positive BufferSize, got %d", opts.BufferSize)
	}
}

func TestSplitter_MatchesGoToLeft(t *testing.T) {
	pred := func(e diff.Entry) bool { return e.Level == "error" }
	s := split.New(pred, split.DefaultOptions())

	src := feedEntries([]diff.Entry{
		makeEntry("svc", "error", "bad"),
		makeEntry("svc", "info", "ok"),
		makeEntry("svc", "error", "also bad"),
	})

	matched, rest := s.Stream(context.Background(), src)
	got := drain(matched)
	other := drain(rest)

	if len(got) != 2 {
		t.Fatalf("expected 2 matched entries, got %d", len(got))
	}
	if len(other) != 1 {
		t.Fatalf("expected 1 rest entry, got %d", len(other))
	}
}

func TestSplitter_NoneMatch_AllGoRight(t *testing.T) {
	pred := func(e diff.Entry) bool { return false }
	s := split.New(pred, split.DefaultOptions())

	src := feedEntries([]diff.Entry{
		makeEntry("a", "info", "one"),
		makeEntry("b", "warn", "two"),
	})

	matched, rest := s.Stream(context.Background(), src)
	if len(drain(matched)) != 0 {
		t.Fatal("expected no matched entries")
	}
	if len(drain(rest)) != 2 {
		t.Fatal("expected 2 rest entries")
	}
}

func TestSplitter_StopsOnContextCancel(t *testing.T) {
	pred := func(e diff.Entry) bool { return true }
	s := split.New(pred, split.DefaultOptions())

	src := make(chan diff.Entry) // never sends
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	matched, rest := s.Stream(ctx, src)
	drain(matched)
	drain(rest)
	// reaching here without deadlock means the goroutine exited on cancel
}
