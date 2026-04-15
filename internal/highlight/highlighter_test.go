package highlight_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/highlight"
)

func makeEntry(level, message, service string) diff.LogEntry {
	return diff.LogEntry{
		Level:   level,
		Message: message,
		Service: service,
	}
}

func TestApply_NoRules_PassesThrough(t *testing.T) {
	h := highlight.New(highlight.Options{})
	e := makeEntry("info", "hello", "svc")
	out := h.Apply(e)
	if out.Message != e.Message {
		t.Fatalf("expected %q, got %q", e.Message, out.Message)
	}
}

func TestApply_ColoursErrorLevel(t *testing.T) {
	h := highlight.New(highlight.DefaultOptions())
	e := makeEntry("error", "boom", "svc")
	out := h.Apply(e)
	if !strings.Contains(out.Message, "boom") {
		t.Fatal("expected message to contain original text")
	}
	if !strings.HasPrefix(out.Message, string(highlight.Red)) {
		t.Fatal("expected message to start with red ANSI code")
	}
	if !strings.HasSuffix(out.Message, string(highlight.Reset)) {
		t.Fatal("expected message to end with reset ANSI code")
	}
}

func TestApply_CaseFold_MatchesUppercase(t *testing.T) {
	h := highlight.New(highlight.DefaultOptions())
	e := makeEntry("WARN", "careful", "svc")
	out := h.Apply(e)
	if !strings.HasPrefix(out.Message, string(highlight.Yellow)) {
		t.Fatal("expected yellow colour for WARN level")
	}
}

func TestApply_NoCaseFold_NoMatch(t *testing.T) {
	opts := highlight.DefaultOptions()
	opts.CaseFold = false
	h := highlight.New(opts)
	e := makeEntry("ERROR", "boom", "svc")
	out := h.Apply(e)
	// "ERROR" won't match rule value "error" without case folding
	if out.Message != e.Message {
		t.Fatalf("expected no colour applied, got %q", out.Message)
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	h := highlight.New(highlight.DefaultOptions())
	e := makeEntry("error", "original", "svc")
	h.Apply(e)
	if e.Message != "original" {
		t.Fatal("Apply must not mutate the original entry")
	}
}

func TestStream_ColoursEntries(t *testing.T) {
	h := highlight.New(highlight.DefaultOptions())
	in := make(chan diff.LogEntry, 2)
	in <- makeEntry("error", "bad", "a")
	in <- makeEntry("info", "ok", "b")
	close(in)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out := highlight.Stream(ctx, h, in)
	var results []diff.LogEntry
	for e := range out {
		results = append(results, e)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	if !strings.HasPrefix(results[0].Message, string(highlight.Red)) {
		t.Error("first entry should be red")
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	h := highlight.New(highlight.DefaultOptions())
	in := make(chan diff.LogEntry)
	ctx, cancel := context.WithCancel(context.Background())
	out := highlight.Stream(ctx, h, in)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
