package buffer

import (
	"fmt"
	"testing"

	"github.com/logdrift/internal/diff"
)

func makeEntry(msg string) diff.Entry {
	return diff.Entry{Service: "svc", Level: "info", Message: msg, Fields: map[string]string{}}
}

func TestNew_DefaultCapacity(t *testing.T) {
	r := New(0)
	if r.cap != DefaultCapacity {
		t.Fatalf("expected cap %d, got %d", DefaultCapacity, r.cap)
	}
}

func TestPush_And_Len(t *testing.T) {
	r := New(4)
	for i := 0; i < 3; i++ {
		r.Push(makeEntry(fmt.Sprintf("msg%d", i)))
	}
	if r.Len() != 3 {
		t.Fatalf("expected 3, got %d", r.Len())
	}
}

func TestSnapshot_ChronologicalOrder(t *testing.T) {
	r := New(4)
	msgs := []string{"a", "b", "c"}
	for _, m := range msgs {
		r.Push(makeEntry(m))
	}
	snap := r.Snapshot()
	if len(snap) != len(msgs) {
		t.Fatalf("expected %d entries, got %d", len(msgs), len(snap))
	}
	for i, e := range snap {
		if e.Message != msgs[i] {
			t.Errorf("pos %d: expected %q, got %q", i, msgs[i], e.Message)
		}
	}
}

func TestPush_OverwritesOldest(t *testing.T) {
	r := New(3)
	for i := 0; i < 5; i++ {
		r.Push(makeEntry(fmt.Sprintf("m%d", i)))
	}
	snap := r.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(snap))
	}
	expected := []string{"m2", "m3", "m4"}
	for i, e := range snap {
		if e.Message != expected[i] {
			t.Errorf("pos %d: expected %q, got %q", i, expected[i], e.Message)
		}
	}
}

func TestSnapshot_Empty(t *testing.T) {
	r := New(8)
	if snap := r.Snapshot(); snap != nil {
		t.Fatalf("expected nil snapshot, got %v", snap)
	}
}

func TestReset_ClearsBuffer(t *testing.T) {
	r := New(4)
	r.Push(makeEntry("x"))
	r.Push(makeEntry("y"))
	r.Reset()
	if r.Len() != 0 {
		t.Fatalf("expected 0 after reset, got %d", r.Len())
	}
	if snap := r.Snapshot(); snap != nil {
		t.Fatalf("expected nil snapshot after reset")
	}
}
