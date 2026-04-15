package extract

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(msg, level, service string) diff.LogEntry {
	return diff.LogEntry{
		Timestamp: time.Now(),
		Service:   service,
		Level:     level,
		Message:   msg,
		Extra:     map[string]string{},
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := New(Options{
		Rules: []Rule{{SourceField: "message", Pattern: "(["}},
	})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestNew_EmptySourceField_ReturnsError(t *testing.T) {
	_, err := New(Options{
		Rules: []Rule{{SourceField: "", Pattern: `(?P<id>\d+)`}},
	})
	if err == nil {
		t.Fatal("expected error for empty source field")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_NoRules_PassesThrough(t *testing.T) {
	ex, _ := New(DefaultOptions())
	e := makeEntry("hello world", "info", "svc")
	out := ex.Apply(e)
	if out.Message != e.Message {
		t.Errorf("expected message %q, got %q", e.Message, out.Message)
	}
}

func TestApply_ExtractsNamedGroups(t *testing.T) {
	ex, err := New(Options{
		Rules: []Rule{
			{SourceField: "message", Pattern: `user=(?P<user>\w+) action=(?P<action>\w+)`},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	e := makeEntry("user=alice action=login", "info", "auth")
	out := ex.Apply(e)
	if out.Extra["user"] != "alice" {
		t.Errorf("expected user=alice, got %q", out.Extra["user"])
	}
	if out.Extra["action"] != "login" {
		t.Errorf("expected action=login, got %q", out.Extra["action"])
	}
}

func TestApply_OutputPrefix(t *testing.T) {
	ex, _ := New(Options{
		Rules: []Rule{
			{SourceField: "message", Pattern: `id=(?P<id>\d+)`, OutputPrefix: "ext_"},
		},
	})
	out := ex.Apply(makeEntry("id=42", "info", "svc"))
	if out.Extra["ext_id"] != "42" {
		t.Errorf("expected ext_id=42, got %q", out.Extra["ext_id"])
	}
}

func TestApply_NoMatch_NoExtraFields(t *testing.T) {
	ex, _ := New(Options{
		Rules: []Rule{
			{SourceField: "message", Pattern: `id=(?P<id>\d+)`},
		},
	})
	out := ex.Apply(makeEntry("no numbers here", "info", "svc"))
	if _, ok := out.Extra["id"]; ok {
		t.Error("expected no 'id' field in extra")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	ex, _ := New(Options{
		Rules: []Rule{
			{SourceField: "message", Pattern: `user=(?P<user>\w+)`},
		},
	})
	e := makeEntry("user=bob", "info", "svc")
	_ = ex.Apply(e)
	if _, ok := e.Extra["user"]; ok {
		t.Error("original entry was mutated")
	}
}
