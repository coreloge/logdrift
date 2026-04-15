package trim_test

import (
	"context"
	"testing"
	"time"

	"logdrift/internal/diff"
	"logdrift/internal/trim"
)

func makeEntry(svc, level, msg string, extra map[string]string) diff.Entry {
	return diff.Entry{
		Service: svc,
		Level:   level,
		Message: msg,
		Extra:   extra,
	}
}

func TestDefaultOptions_TrimsMessageOnly(t *testing.T) {
	opts := trim.DefaultOptions()
	tr := trim.New(opts)

	e := makeEntry("  svc  ", "  info  ", "  hello world  ", nil)
	out := tr.Apply(e)

	if out.Message != "hello world" {
		t.Errorf("expected trimmed message, got %q", out.Message)
	}
	if out.Service != "  svc  " {
		t.Errorf("service should not be trimmed, got %q", out.Service)
	}
	if out.Level != "  info  " {
		t.Errorf("level should not be trimmed, got %q", out.Level)
	}
}

func TestApply_TrimsLevel(t *testing.T) {
	tr := trim.New(trim.Options{Fields: []string{"level"}})
	e := makeEntry("svc", "  warn  ", "msg", nil)
	out := tr.Apply(e)
	if out.Level != "warn" {
		t.Errorf("expected %q, got %q", "warn", out.Level)
	}
}

func TestApply_TrimsExtraField(t *testing.T) {
	tr := trim.New(trim.Options{Fields: []string{"trace_id"}})
	e := makeEntry("svc", "info", "msg", map[string]string{"trace_id": "  abc123  "})
	out := tr.Apply(e)
	if out.Extra["trace_id"] != "abc123" {
		t.Errorf("expected trimmed extra field, got %q", out.Extra["trace_id"])
	}
}

func TestApply_TrimExtra_AllKeys(t *testing.T) {
	tr := trim.New(trim.Options{TrimExtra: true})
	e := makeEntry("svc", "info", "msg", map[string]string{
		"a": "  val1  ",
		"b": "\tval2\n",
	})
	out := tr.Apply(e)
	if out.Extra["a"] != "val1" {
		t.Errorf("a: expected %q, got %q", "val1", out.Extra["a"])
	}
	if out.Extra["b"] != "val2" {
		t.Errorf("b: expected %q, got %q", "val2", out.Extra["b"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	tr := trim.New(trim.DefaultOptions())
	e := makeEntry("svc", "info", "  original  ", nil)
	_ = tr.Apply(e)
	if e.Message != "  original  " {
		t.Errorf("original entry was mutated")
	}
}

func TestStream_TrimsAndForwards(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := make(chan diff.Entry, 2)
	in <- makeEntry("svc", "info", "  trimmed  ", nil)
	in <- makeEntry("svc", "warn", "clean", nil)
	close(in)

	out := trim.Stream(ctx, in, trim.DefaultOptions())

	var results []diff.Entry
	for e := range out {
		results = append(results, e)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	if results[0].Message != "trimmed" {
		t.Errorf("expected %q, got %q", "trimmed", results[0].Message)
	}
	if results[1].Message != "clean" {
		t.Errorf("expected %q, got %q", "clean", results[1].Message)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan diff.Entry)
	out := trim.Stream(ctx, in, trim.DefaultOptions())
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
