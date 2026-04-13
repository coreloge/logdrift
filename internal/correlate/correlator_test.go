package correlate

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(service, corrID, msg string) diff.Entry {
	fields := map[string]string{"msg": msg}
	if corrID != "" {
		fields["correlation_id"] = corrID
	}
	return diff.Entry{
		Service: service,
		Level:   "info",
		Message: msg,
		Fields:  fields,
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Field == "" {
		t.Error("expected non-empty default field")
	}
	if opts.Window <= 0 {
		t.Error("expected positive default window")
	}
	if opts.MaxIDs <= 0 {
		t.Error("expected positive default MaxIDs")
	}
}

func TestRecord_NoField_Ignored(t *testing.T) {
	c := New(DefaultOptions())
	c.Record(makeEntry("svc", "", "hello"))
	if entries := c.Get("anything"); entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}

func TestRecord_And_Get(t *testing.T) {
	c := New(DefaultOptions())
	c.Record(makeEntry("svc-a", "req-1", "start"))
	c.Record(makeEntry("svc-b", "req-1", "end"))

	entries := c.Get("req-1")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestGet_UnknownID_ReturnsNil(t *testing.T) {
	c := New(DefaultOptions())
	if got := c.Get("unknown"); got != nil {
		t.Errorf("expected nil for unknown ID, got %v", got)
	}
}

func TestEviction_StaleGroupRemoved(t *testing.T) {
	opts := DefaultOptions()
	opts.Window = 50 * time.Millisecond
	c := New(opts)

	c.Record(makeEntry("svc", "req-old", "msg"))
	time.Sleep(100 * time.Millisecond)

	// Trigger eviction via a new record.
	c.Record(makeEntry("svc", "req-new", "msg"))

	if got := c.Get("req-old"); got != nil {
		t.Errorf("expected stale group to be evicted, got %v", got)
	}
}

func TestMaxIDs_CapEnforced(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxIDs = 2
	c := New(opts)

	c.Record(makeEntry("svc", "id-1", "a"))
	c.Record(makeEntry("svc", "id-2", "b"))
	// Third ID should be dropped.
	c.Record(makeEntry("svc", "id-3", "c"))

	if got := c.Get("id-3"); got != nil {
		t.Errorf("expected id-3 to be dropped at capacity, got %v", got)
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	c := New(DefaultOptions())
	c.Record(makeEntry("svc", "req-1", "original"))

	a := c.Get("req-1")
	a[0].Message = "mutated"

	b := c.Get("req-1")
	if b[0].Message == "mutated" {
		t.Error("Get should return a copy, not a reference to internal slice")
	}
}
