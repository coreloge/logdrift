package pin

import (
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{Service: svc, Message: msg, Level: "info", Timestamp: time.Now()}
}

func TestNew_NegativeMax_ReturnsError(t *testing.T) {
	_, err := New(Options{MaxPinned: -1})
	if err == nil {
		t.Fatal("expected error for negative MaxPinned")
	}
}

func TestNew_ZeroMax_ReturnsError(t *testing.T) {
	_, err := New(Options{MaxPinned: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxPinned")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPin_And_All(t *testing.T) {
	p, _ := New(DefaultOptions())
	p.Pin(makeEntry("svcA", "hello"))
	p.Pin(makeEntry("svcB", "world"))
	all := p.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 pinned entries, got %d", len(all))
	}
}

func TestPin_EvictsOldestWhenFull(t *testing.T) {
	p, _ := New(Options{MaxPinned: 2})
	p.Pin(makeEntry("svc", "first"))
	p.Pin(makeEntry("svc", "second"))
	p.Pin(makeEntry("svc", "third"))
	all := p.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries after eviction, got %d", len(all))
	}
	if all[0].Message != "second" {
		t.Errorf("expected oldest surviving entry to be 'second', got %q", all[0].Message)
	}
}

func TestClear_RemovesAll(t *testing.T) {
	p, _ := New(DefaultOptions())
	p.Pin(makeEntry("svc", "msg"))
	p.Clear()
	if len(p.All()) != 0 {
		t.Fatal("expected empty after Clear")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	p, _ := New(DefaultOptions())
	p.Pin(makeEntry("svc", "msg"))
	a := p.All()
	a[0].Message = "mutated"
	if p.All()[0].Message == "mutated" {
		t.Fatal("All should return a copy, not a reference")
	}
}
