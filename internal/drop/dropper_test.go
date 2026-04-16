package drop

import (
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
)

func makeEntry(svc, level, msg string, extra map[string]string) diff.Entry {
	return diff.Entry{Service: svc, Level: level, Message: msg, Timestamp: time.Now(), Extra: extra}
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
	_, err := New(Options{Rules: []Rule{{Field: "level", Pattern: ""}}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestShouldDrop_NoRules_ReturnsFalse(t *testing.T) {
	d, _ := New(DefaultOptions())
	e := makeEntry("svc", "info", "hello", nil)
	if d.ShouldDrop(e) {
		t.Fatal("expected false with no rules")
	}
}

func TestShouldDrop_MatchesMessage(t *testing.T) {
	d, _ := New(Options{Rules: []Rule{{Field: "message", Pattern: "health"}}})
	if !d.ShouldDrop(makeEntry("svc", "info", "healthcheck ok", nil)) {
		t.Fatal("expected drop on message match")
	}
	if d.ShouldDrop(makeEntry("svc", "info", "normal log", nil)) {
		t.Fatal("expected pass on non-matching message")
	}
}

func TestShouldDrop_MatchesLevel(t *testing.T) {
	d, _ := New(Options{Rules: []Rule{{Field: "level", Pattern: "^debug$"}}})
	if !d.ShouldDrop(makeEntry("svc", "debug", "verbose", nil)) {
		t.Fatal("expected drop on debug level")
	}
	if d.ShouldDrop(makeEntry("svc", "info", "important", nil)) {
		t.Fatal("expected pass on info level")
	}
}

func TestShouldDrop_MatchesExtraField(t *testing.T) {
	d, _ := New(Options{Rules: []Rule{{Field: "env", Pattern: "staging"}}})
	e := makeEntry("svc", "info", "msg", map[string]string{"env": "staging"})
	if !d.ShouldDrop(e) {
		t.Fatal("expected drop on extra field match")
	}
}
