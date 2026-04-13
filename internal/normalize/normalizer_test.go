package normalize_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/normalize"
)

func makeEntry(level, message string, fields map[string]string) diff.Entry {
	if fields == nil {
		fields = map[string]string{}
	}
	return diff.Entry{
		Service:   "svc",
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}
}

func TestApply_NoRules_PassesThrough(t *testing.T) {
	n := normalize.New(normalize.Options{})
	e := makeEntry("INFO", "  hello  ", nil)
	out := n.Apply(e)
	if out.Level != "INFO" || out.Message != "  hello  " {
		t.Fatalf("expected unchanged entry, got level=%q message=%q", out.Level, out.Message)
	}
}

func TestApply_LowerLevel(t *testing.T) {
	n := normalize.New(normalize.Options{
		Rules: []normalize.Rule{{Field: "level", Op: normalize.OpLower}},
	})
	out := n.Apply(makeEntry("WARNING", "msg", nil))
	if out.Level != "warning" {
		t.Fatalf("expected %q, got %q", "warning", out.Level)
	}
}

func TestApply_TrimMessage(t *testing.T) {
	n := normalize.New(normalize.Options{
		Rules: []normalize.Rule{{Field: "message", Op: normalize.OpTrim}},
	})
	out := n.Apply(makeEntry("info", "  spaced  ", nil))
	if out.Message != "spaced" {
		t.Fatalf("expected %q, got %q", "spaced", out.Message)
	}
}

func TestApply_CollapseCustomField(t *testing.T) {
	n := normalize.New(normalize.Options{
		Rules: []normalize.Rule{{Field: "trace", Op: normalize.OpCollapse}},
	})
	e := makeEntry("info", "msg", map[string]string{"trace": "a  b   c"})
	out := n.Apply(e)
	if out.Fields["trace"] != "a b c" {
		t.Fatalf("expected %q, got %q", "a b c", out.Fields["trace"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	n := normalize.New(normalize.DefaultOptions())
	e := makeEntry("ERROR", "  raw  ", nil)
	_ = n.Apply(e)
	if e.Level != "ERROR" || e.Message != "  raw  " {
		t.Fatal("original entry was mutated")
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

func TestStream_NormalisesEntries(t *testing.T) {
	n := normalize.New(normalize.Options{
		Rules: []normalize.Rule{{Field: "level", Op: normalize.OpUpper}},
	})
	in := feedEntries([]diff.Entry{
		makeEntry("info", "a", nil),
		makeEntry("warn", "b", nil),
	})
	results := drainStream(normalize.Stream(context.Background(), n, in))
	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	for _, r := range results {
		if r.Level != "INFO" && r.Level != "WARN" {
			t.Errorf("unexpected level %q", r.Level)
		}
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	n := normalize.New(normalize.DefaultOptions())
	in := make(chan diff.Entry) // never sends
	results := drainStream(normalize.Stream(ctx, n, in))
	if len(results) != 0 {
		t.Fatalf("expected 0 entries after cancel, got %d", len(results))
	}
}
