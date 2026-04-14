package suppress

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(svc, msg, level string) diff.LogEntry {
	return diff.LogEntry{Service: svc, Message: msg, Level: level}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := New(Options{Pattern: "["})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_FirstOccurrence_ReturnsTrue(t *testing.T) {
	s, _ := New(DefaultOptions())
	e := makeEntry("svc", "hello", "info")
	if !s.Allow(e) {
		t.Fatal("expected first occurrence to be allowed")
	}
}

func TestAllow_SecondOccurrence_WithinCooldown_ReturnsFalse(t *testing.T) {
	s, _ := New(Options{Field: "message", Cooldown: 10 * time.Second})
	e := makeEntry("svc", "hello", "info")
	s.Allow(e)
	if s.Allow(e) {
		t.Fatal("expected second occurrence within cooldown to be suppressed")
	}
}

func TestAllow_DifferentService_IndependentCooldown(t *testing.T) {
	s, _ := New(Options{Field: "message", Cooldown: 10 * time.Second})
	a := makeEntry("svc-a", "hello", "info")
	b := makeEntry("svc-b", "hello", "info")
	s.Allow(a)
	if !s.Allow(b) {
		t.Fatal("different service should not share cooldown")
	}
}

func TestAllow_PatternMismatch_AlwaysPasses(t *testing.T) {
	s, _ := New(Options{Field: "message", Pattern: `error`, Cooldown: 10 * time.Second})
	e := makeEntry("svc", "all good", "info")
	s.Allow(e)
	if !s.Allow(e) {
		t.Fatal("non-matching entry should always pass")
	}
}

func TestAllow_AfterCooldownExpires_ReturnsTrue(t *testing.T) {
	s, _ := New(Options{Field: "message", Cooldown: 10 * time.Millisecond})
	e := makeEntry("svc", "tick", "warn")
	s.Allow(e)
	time.Sleep(20 * time.Millisecond)
	if !s.Allow(e) {
		t.Fatal("expected entry to be allowed after cooldown expires")
	}
}

func TestAllow_ExtraField_Suppressed(t *testing.T) {
	s, _ := New(Options{Field: "request_id", Cooldown: 10 * time.Second})
	e := diff.LogEntry{
		Service: "svc",
		Message: "request",
		Level:   "info",
		Extra:   map[string]string{"request_id": "abc123"},
	}
	s.Allow(e)
	if s.Allow(e) {
		t.Fatal("expected repeated extra-field value to be suppressed")
	}
}
