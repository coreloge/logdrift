package timeskew

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(service string, ts time.Time) diff.Entry {
	return diff.Entry{
		Service:   service,
		Timestamp: ts,
		Level:     "info",
		Message:   "test message",
	}
}

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_ZeroMaxSkew_ReturnsError(t *testing.T) {
	_, err := New(Options{MaxSkew: 0, ReferenceNow: time.Now})
	if err == nil {
		t.Fatal("expected error for zero MaxSkew, got nil")
	}
}

func TestNew_NegativeMaxSkew_ReturnsError(t *testing.T) {
	_, err := New(Options{MaxSkew: -1 * time.Second, ReferenceNow: time.Now})
	if err == nil {
		t.Fatal("expected error for negative MaxSkew, got nil")
	}
}

func TestNew_NilReferenceNow_DefaultsToTimeNow(t *testing.T) {
	d, err := New(Options{MaxSkew: time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.opts.ReferenceNow == nil {
		t.Fatal("expected ReferenceNow to be set, got nil")
	}
}

func TestCheck_WithinSkew_ReturnsNil(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	d, _ := New(Options{MaxSkew: 5 * time.Second, ReferenceNow: fixedNow(now)})

	entry := makeEntry("svc-a", now.Add(3*time.Second))
	if v := d.Check(entry); v != nil {
		t.Fatalf("expected nil violation, got: %v", v)
	}
}

func TestCheck_ExceedsSkew_ReturnsViolation(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	d, _ := New(Options{MaxSkew: 5 * time.Second, ReferenceNow: fixedNow(now)})

	entry := makeEntry("svc-b", now.Add(10*time.Second))
	v := d.Check(entry)
	if v == nil {
		t.Fatal("expected violation, got nil")
	}
	if v.Service != "svc-b" {
		t.Errorf("expected service %q, got %q", "svc-b", v.Service)
	}
	if v.Skew != 10*time.Second {
		t.Errorf("expected skew 10s, got %s", v.Skew)
	}
}

func TestCheck_PastTimestamp_ReturnsViolation(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	d, _ := New(Options{MaxSkew: 2 * time.Second, ReferenceNow: fixedNow(now)})

	entry := makeEntry("svc-c", now.Add(-30*time.Second))
	v := d.Check(entry)
	if v == nil {
		t.Fatal("expected violation for past timestamp, got nil")
	}
	if v.Skew != 30*time.Second {
		t.Errorf("expected skew 30s, got %s", v.Skew)
	}
}

func TestCheck_ExactlyAtMaxSkew_ReturnsNil(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	d, _ := New(Options{MaxSkew: 5 * time.Second, ReferenceNow: fixedNow(now)})

	entry := makeEntry("svc-d", now.Add(5*time.Second))
	if v := d.Check(entry); v != nil {
		t.Fatalf("expected nil at exact boundary, got: %v", v)
	}
}

func TestViolation_Error_ContainsService(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	v := &Violation{
		Service: "svc-e",
		EntryTS: now.Add(20 * time.Second),
		Now:     now,
		Skew:    20 * time.Second,
		MaxSkew: 5 * time.Second,
	}
	msg := v.Error()
	if len(msg) == 0 {
		t.Fatal("expected non-empty error message")
	}
}
