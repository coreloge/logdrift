package keep

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(svc, level, msg string) diff.Entry {
	return diff.Entry{Service: svc, Level: level, Message: msg, Timestamp: time.Now()}
}

func TestNew_NoRules_NoError(t *testing.T) {
	_, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := New(Options{Rules: []Rule{{Pattern: "[invalid"}}})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestNew_EmptyPattern_ReturnsError(t *testing.T) {
	_, err := New(Options{Rules: []Rule{{Field: "message", Pattern: ""}}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestShouldKeep_NoRules_ReturnsTrue(t *testing.T) {
	k, _ := New(DefaultOptions())
	if !k.ShouldKeep(makeEntry("svc", "info", "hello")) {
		t.Fatal("expected true with no rules")
	}
}

func TestShouldKeep_MatchesMessage(t *testing.T) {
	k, _ := New(Options{Rules: []Rule{{Pattern: "error"}}})
	if !k.ShouldKeep(makeEntry("svc", "info", "an error occurred")) {
		t.Fatal("expected match on message")
	}
	if k.ShouldKeep(makeEntry("svc", "info", "all good")) {
		t.Fatal("expected no match")
	}
}

func TestShouldKeep_MatchesLevel(t *testing.T) {
	k, _ := New(Options{Rules: []Rule{{Field: "level", Pattern: "^error$"}}})
	if !k.ShouldKeep(makeEntry("svc", "error", "boom")) {
		t.Fatal("expected match on level")
	}
	if k.ShouldKeep(makeEntry("svc", "info", "fine")) {
		t.Fatal("expected no match on level")
	}
}

func TestShouldKeep_MatchesExtraField(t *testing.T) {
	k, _ := New(Options{Rules: []Rule{{Field: "env", Pattern: "prod"}}})
	e := makeEntry("svc", "info", "msg")
	e.Extra = map[string]interface{}{"env": "production"}
	if !k.ShouldKeep(e) {
		t.Fatal("expected match on extra field")
	}
}
