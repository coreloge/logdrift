package metrics

import (
	"testing"
)

func buildStore(t *testing.T, maxLen int) (*SnapshotStore, *Counter) {
	t.Helper()
	store := NewSnapshotStore(maxLen)
	c := New()
	return store, c
}

func TestNewSnapshotStore_DefaultMaxLen(t *testing.T) {
	s := NewSnapshotStore(0)
	if s.maxLen != 60 {
		t.Fatalf("expected default maxLen 60, got %d", s.maxLen)
	}
}

func TestSnapshotStore_EmptyLatest(t *testing.T) {
	s, _ := buildStore(t, 10)
	_, ok := s.Latest()
	if ok {
		t.Fatal("expected false for empty store")
	}
}

func TestSnapshotStore_CaptureAndLatest(t *testing.T) {
	s, c := buildStore(t, 10)
	c.RecordEntry("svc-a")
	c.RecordEntry("svc-a")
	c.RecordDrift("svc-a")

	s.Capture(c)

	rec, ok := s.Latest()
	if !ok {
		t.Fatal("expected a record")
	}
	if rec.Entries["svc-a"] != 2 {
		t.Errorf("expected entries 2, got %d", rec.Entries["svc-a"])
	}
	if rec.Drifts["svc-a"] != 1 {
		t.Errorf("expected drifts 1, got %d", rec.Drifts["svc-a"])
	}
}

func TestSnapshotStore_All_Chronological(t *testing.T) {
	s, c := buildStore(t, 10)
	for i := 0; i < 3; i++ {
		c.RecordEntry("svc")
		s.Capture(c)
	}
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 records, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i].Timestamp.Before(all[i-1].Timestamp) {
			t.Errorf("records not in chronological order at index %d", i)
		}
	}
}

func TestSnapshotStore_BoundedByMaxLen(t *testing.T) {
	const max = 5
	s, c := buildStore(t, max)
	for i := 0; i < 12; i++ {
		c.RecordEntry("svc")
		s.Capture(c)
	}
	if s.Len() != max {
		t.Errorf("expected store len %d, got %d", max, s.Len())
	}
}

func TestSnapshotStore_All_ReturnsCopy(t *testing.T) {
	s, c := buildStore(t, 10)
	s.Capture(c)
	all := s.All()
	all[0].Entries["injected"] = 99
	_, ok := s.Latest()
	if !ok {
		t.Fatal("expected record")
	}
	// Mutating the returned slice must not affect internal state.
	all2 := s.All()
	if _, found := all2[0].Entries["injected"]; found {
		t.Error("mutation of returned slice affected internal state")
	}
}
