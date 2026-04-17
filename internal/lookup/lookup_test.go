package lookup

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(service, level, msg string) diff.Entry {
	return diff.Entry{
		Service:   service,
		Level:     level,
		Message:   msg,
		Timestamp: time.Now(),
	}
}

func TestNew_EmptySourceField_ReturnsError(t *testing.T) {
	opts := DefaultOptions()
	opts.SourceField = ""
	_, err := New(opts)
	if err == nil {
		t.Fatal("expected error for empty SourceField")
	}
}

func TestNew_EmptyOutputField_ReturnsError(t *testing.T) {
	opts := DefaultOptions()
	opts.OutputField = ""
	_, err := New(opts)
	if err == nil {
		t.Fatal("expected error for empty OutputField")
	}
}

func TestNew_NilTable_ReturnsError(t *testing.T) {
	opts := DefaultOptions()
	opts.Table = nil
	_, err := New(opts)
	if err == nil {
		t.Fatal("expected error for nil Table")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_ResolvesMatch(t *testing.T) {
	opts := DefaultOptions()
	opts.Table = map[string]string{"auth": "platform"}
	l, _ := New(opts)

	e := makeEntry("auth", "info", "login ok")
	out := l.Apply(e)

	if out.Extra["team"] != "platform" {
		t.Fatalf("expected team=platform, got %q", out.Extra["team"])
	}
}

func TestApply_NoMatch_WritesDefault(t *testing.T) {
	opts := DefaultOptions()
	opts.Table = map[string]string{}
	opts.Default = "unknown"
	l, _ := New(opts)

	e := makeEntry("payments", "warn", "timeout")
	out := l.Apply(e)

	if out.Extra["team"] != "unknown" {
		t.Fatalf("expected team=unknown, got %q", out.Extra["team"])
	}
}

func TestApply_NoMatch_NoDefault_PassesThrough(t *testing.T) {
	l, _ := New(DefaultOptions())
	e := makeEntry("payments", "warn", "timeout")
	out := l.Apply(e)
	if v, ok := out.Extra["team"]; ok {
		t.Fatalf("expected no team field, got %q", v)
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	opts := DefaultOptions()
	opts.Table = map[string]string{"auth": "platform"}
	l, _ := New(opts)

	e := makeEntry("auth", "info", "ok")
	_ = l.Apply(e)

	if e.Extra != nil {
		t.Fatal("original entry Extra was mutated")
	}
}
