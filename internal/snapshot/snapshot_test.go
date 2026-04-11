package snapshot_test

import (
	"testing"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/snapshot"
)

func makeEntry(level, msg string) diff.Entry {
	return diff.Entry{Level: level, Message: msg, Fields: map[string]string{}}
}

func TestNew_EmptySnapshot(t *testing.T) {
	s := snapshot.New()
	if s == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if len(s.Services()) != 0 {
		t.Errorf("expected 0 services, got %d", len(s.Services()))
	}
}

func TestAdd_AndGet(t *testing.T) {
	s := snapshot.New()
	e := makeEntry("info", "hello")
	s.Add("svc-a", e)

	entries := s.Get("svc-a")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Message != "hello" {
		t.Errorf("unexpected message: %s", entries[0].Message)
	}
}

func TestGet_MissingService(t *testing.T) {
	s := snapshot.New()
	if entries := s.Get("missing"); entries != nil {
		t.Errorf("expected nil for missing service, got %v", entries)
	}
}

func TestServices_ReturnsAllKeys(t *testing.T) {
	s := snapshot.New()
	s.Add("alpha", makeEntry("info", "a"))
	s.Add("beta", makeEntry("warn", "b"))

	svcs := s.Services()
	if len(svcs) != 2 {
		t.Errorf("expected 2 services, got %d", len(svcs))
	}
}

func TestCompare_DetectsDiff(t *testing.T) {
	base := snapshot.New()
	base.Add("svc", makeEntry("info", "original"))

	curr := snapshot.New()
	curr.Add("svc", makeEntry("error", "original"))

	results := snapshot.Compare(base, curr, "svc")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Equal {
		t.Error("expected diff to be detected")
	}
}

func TestCompare_NoEntries(t *testing.T) {
	base := snapshot.New()
	curr := snapshot.New()
	results := snapshot.Compare(base, curr, "svc")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
