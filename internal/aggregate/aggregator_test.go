package aggregate

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(level, service string, ts time.Time) diff.Entry {
	return diff.Entry{
		Service:   service,
		Timestamp: ts,
		Fields:    map[string]string{"level": level, "service": service},
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.KeyField != "level" {
		t.Fatalf("expected key field 'level', got %q", opts.KeyField)
	}
	if opts.Window != time.Minute {
		t.Fatalf("expected window 1m, got %v", opts.Window)
	}
}

func TestRecord_SingleEntry(t *testing.T) {
	a := New(DefaultOptions())
	a.Record(makeEntry("info", "svc-a", time.Now()))

	buckets := a.Snapshot()
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if buckets[0].Key != "info" {
		t.Errorf("expected key 'info', got %q", buckets[0].Key)
	}
	if buckets[0].Count != 1 {
		t.Errorf("expected count 1, got %d", buckets[0].Count)
	}
}

func TestRecord_MultipleEntriesSameKey(t *testing.T) {
	a := New(DefaultOptions())
	now := time.Now()
	for i := 0; i < 5; i++ {
		a.Record(makeEntry("error", "svc-b", now.Add(time.Duration(i)*time.Second)))
	}

	buckets := a.Snapshot()
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if buckets[0].Count != 5 {
		t.Errorf("expected count 5, got %d", buckets[0].Count)
	}
}

func TestRecord_MultipleKeys(t *testing.T) {
	a := New(DefaultOptions())
	now := time.Now()
	a.Record(makeEntry("info", "svc-a", now))
	a.Record(makeEntry("warn", "svc-a", now))
	a.Record(makeEntry("error", "svc-a", now))

	buckets := a.Snapshot()
	if len(buckets) != 3 {
		t.Errorf("expected 3 buckets, got %d", len(buckets))
	}
}

func TestEvict_RemovesExpiredBuckets(t *testing.T) {
	opts := Options{KeyField: "level", Window: 100 * time.Millisecond}
	a := New(opts)

	past := time.Now().Add(-200 * time.Millisecond)
	a.Record(makeEntry("debug", "svc-a", past))

	// Record a fresh entry to trigger eviction
	a.Record(makeEntry("info", "svc-a", time.Now()))

	buckets := a.Snapshot()
	for _, b := range buckets {
		if b.Key == "debug" {
			t.Error("expected 'debug' bucket to be evicted")
		}
	}
}

func TestReset_ClearsBuckets(t *testing.T) {
	a := New(DefaultOptions())
	a.Record(makeEntry("info", "svc-a", time.Now()))
	a.Reset()

	if len(a.Snapshot()) != 0 {
		t.Error("expected empty snapshot after reset")
	}
}

func TestRecord_UnknownKeyField_UsesUnknown(t *testing.T) {
	opts := Options{KeyField: "nonexistent", Window: time.Minute}
	a := New(opts)
	a.Record(makeEntry("info", "svc-a", time.Now()))

	buckets := a.Snapshot()
	if len(buckets) != 1 || buckets[0].Key != "(unknown)" {
		t.Errorf("expected '(unknown)' bucket, got %+v", buckets)
	}
}
