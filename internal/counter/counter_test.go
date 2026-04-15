package counter

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(service, level, msg string) diff.Entry {
	return diff.Entry{
		Timestamp: time.Now(),
		Service:   service,
		Level:     level,
		Message:   msg,
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Field != "level" {
		t.Fatalf("expected field=level, got %q", opts.Field)
	}
	if opts.OutputField != "_count" {
		t.Fatalf("expected output_field=_count, got %q", opts.OutputField)
	}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := New(Options{Field: ""})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecord_IncrementsCount(t *testing.T) {
	c, _ := New(DefaultOptions())
	e := makeEntry("svc", "error", "boom")

	out := c.Record(e)
	if out.Extra["_count"] != "1" {
		t.Fatalf("expected _count=1, got %q", out.Extra["_count"])
	}

	out2 := c.Record(e)
	if out2.Extra["_count"] != "2" {
		t.Fatalf("expected _count=2, got %q", out2.Extra["_count"])
	}
}

func TestRecord_SeparateValuesTrackedIndependently(t *testing.T) {
	c, _ := New(DefaultOptions())
	c.Record(makeEntry("svc", "error", "a"))
	c.Record(makeEntry("svc", "warn", "b"))
	c.Record(makeEntry("svc", "error", "c"))

	counts := c.Counts()
	if counts["error"] != 2 {
		t.Fatalf("expected error=2, got %d", counts["error"])
	}
	if counts["warn"] != 1 {
		t.Fatalf("expected warn=1, got %d", counts["warn"])
	}
}

func TestReset_ZeroesCounters(t *testing.T) {
	c, _ := New(DefaultOptions())
	c.Record(makeEntry("svc", "error", "x"))
	c.Reset()

	if len(c.Counts()) != 0 {
		t.Fatal("expected empty counts after reset")
	}
}

func TestRecord_DoesNotMutateOriginal(t *testing.T) {
	c, _ := New(DefaultOptions())
	e := makeEntry("svc", "info", "hello")
	c.Record(e)

	if e.Extra != nil {
		t.Fatal("original entry should not have Extra set")
	}
}

func TestRecord_CustomField(t *testing.T) {
	c, _ := New(Options{Field: "service", OutputField: "svc_count"})
	c.Record(makeEntry("alpha", "info", "a"))
	out := c.Record(makeEntry("alpha", "info", "b"))

	if out.Extra["svc_count"] != "2" {
		t.Fatalf("expected svc_count=2, got %q", out.Extra["svc_count"])
	}
}
