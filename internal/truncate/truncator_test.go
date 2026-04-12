package truncate_test

import (
	"context"
	"testing"
	"time"

	"logdrift/internal/diff"
	"logdrift/internal/truncate"
)

func makeEntry(msg string, fields map[string]string) diff.Entry {
	if fields == nil {
		fields = map[string]string{}
	}
	return diff.Entry{Service: "svc", Level: "info", Message: msg, Fields: fields}
}

func TestApply_NoLimits_PassesThrough(t *testing.T) {
	tr := truncate.New(truncate.Options{})
	e := makeEntry("hello world", map[string]string{"key": "value"})
	out := tr.Apply(e)
	if out.Message != e.Message {
		t.Fatalf("expected %q, got %q", e.Message, out.Message)
	}
	if out.Fields["key"] != "value" {
		t.Fatalf("field unexpectedly changed")
	}
}

func TestApply_TruncatesLongMessage(t *testing.T) {
	tr := truncate.New(truncate.Options{MaxMessageLen: 5, Ellipsis: "..."})
	e := makeEntry("hello world", nil)
	out := tr.Apply(e)
	if out.Message != "hello..." {
		t.Fatalf("expected %q, got %q", "hello...", out.Message)
	}
}

func TestApply_TruncatesLongField(t *testing.T) {
	tr := truncate.New(truncate.Options{MaxFieldLen: 4, Ellipsis: "…"})
	e := makeEntry("msg", map[string]string{"k": "abcdefgh"})
	out := tr.Apply(e)
	if out.Fields["k"] != "abcd…" {
		t.Fatalf("expected %q, got %q", "abcd…", out.Fields["k"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	tr := truncate.New(truncate.Options{MaxMessageLen: 3, Ellipsis: "."})
	e := makeEntry("hello", map[string]string{"x": "world"})
	_ = tr.Apply(e)
	if e.Message != "hello" {
		t.Fatal("original entry was mutated")
	}
}

func TestDefaultOptions_LimitsApplied(t *testing.T) {
	opts := truncate.DefaultOptions()
	tr := truncate.New(opts)
	long := string(make([]rune, 300))
	for i := range []rune(long) {
		_ = i
	}
	long = "abcdefghij"
	for len([]rune(long)) < 300 {
		long += "x"
	}
	e := makeEntry(long, nil)
	out := tr.Apply(e)
	if len([]rune(out.Message)) > opts.MaxMessageLen+len([]rune(opts.Ellipsis)) {
		t.Fatalf("message not truncated: len=%d", len([]rune(out.Message)))
	}
}

func TestStream_TruncatesEntries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := make(chan diff.Entry, 2)
	in <- makeEntry("hello world", nil)
	in <- makeEntry("short", nil)
	close(in)

	out := truncate.Stream(ctx, in, truncate.Options{MaxMessageLen: 5, Ellipsis: "…"})
	var results []diff.Entry
	for e := range out {
		results = append(results, e)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	if results[0].Message != "hello…" {
		t.Fatalf("expected truncated message, got %q", results[0].Message)
	}
	if results[1].Message != "short" {
		t.Fatalf("expected unchanged message, got %q", results[1].Message)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan diff.Entry)
	out := truncate.Stream(ctx, in, truncate.DefaultOptions())
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("stream did not close after context cancel")
	}
}
