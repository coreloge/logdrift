package audit

import (
	"testing"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeDelta(field, a, b string) diff.Delta {
	return diff.Delta{Field: field, Got: a, Want: b}
}

func TestNew_DefaultMaxRecords(t *testing.T) {
	a := New(Options{})
	if a.opts.MaxRecords != 1000 {
		t.Fatalf("expected 1000, got %d", a.opts.MaxRecords)
	}
}

func TestRecordEntry_AppendsRecord(t *testing.T) {
	a := New(DefaultOptions())
	a.RecordEntry("svc-a", "info", "hello")
	records := a.All()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	r := records[0]
	if r.Service != "svc-a" || r.Level != "info" || r.Message != "hello" {
		t.Errorf("unexpected record: %+v", r)
	}
	if r.Drift {
		t.Error("expected Drift=false for plain entry")
	}
}

func TestRecordDrift_SetsDriftFlag(t *testing.T) {
	a := New(DefaultOptions())
	deltas := []diff.Delta{makeDelta("level", "info", "error")}
	a.RecordDrift("svc-b", deltas)
	records := a.All()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	r := records[0]
	if !r.Drift {
		t.Error("expected Drift=true")
	}
	if len(r.Deltas) != 1 {
		t.Errorf("expected 1 delta, got %d", len(r.Deltas))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	a := New(DefaultOptions())
	a.RecordEntry("svc", "debug", "msg")
	records := a.All()
	records[0].Service = "mutated"
	original := a.All()
	if original[0].Service == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}

func TestReset_ClearsRecords(t *testing.T) {
	a := New(DefaultOptions())
	a.RecordEntry("svc", "info", "msg")
	a.Reset()
	if len(a.All()) != 0 {
		t.Error("expected empty records after Reset")
	}
}

func TestMaxRecords_EvictsOldest(t *testing.T) {
	a := New(Options{MaxRecords: 3})
	for i := 0; i < 5; i++ {
		a.RecordEntry("svc", "info", "msg")
	}
	records := a.All()
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
}
