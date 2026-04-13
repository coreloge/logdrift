package coalesce

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(service, reqID string) diff.Entry {
	return diff.Entry{
		Service: service,
		Level:   "info",
		Message: "test message",
		Fields:  map[string]string{"request_id": reqID},
	}
}

func TestRecord_NoCorrelationField_ReturnsNil(t *testing.T) {
	c := New(DefaultOptions())
	e := diff.Entry{Service: "svc-a", Fields: map[string]string{}}
	if got := c.Record(e); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestRecord_SingleSource_DoesNotEmit(t *testing.T) {
	c := New(DefaultOptions())
	if got := c.Record(makeEntry("svc-a", "req-1")); got != nil {
		t.Fatalf("expected nil before MinSources reached, got %v", got)
	}
}

func TestRecord_MinSourcesMet_EmitsGroup(t *testing.T) {
	c := New(DefaultOptions())
	c.Record(makeEntry("svc-a", "req-1"))
	got := c.Record(makeEntry("svc-b", "req-1"))
	if got == nil {
		t.Fatal("expected merged group, got nil")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestRecord_DifferentCorrelationIDs_Isolated(t *testing.T) {
	c := New(DefaultOptions())
	c.Record(makeEntry("svc-a", "req-1"))
	got := c.Record(makeEntry("svc-b", "req-2"))
	if got != nil {
		t.Fatalf("different IDs should not merge, got %v", got)
	}
}

func TestRecord_SameServiceTwice_DoesNotSatisfyMinSources(t *testing.T) {
	opts := DefaultOptions()
	opts.MinSources = 2
	c := New(opts)
	c.Record(makeEntry("svc-a", "req-1"))
	got := c.Record(makeEntry("svc-a", "req-1"))
	if got != nil {
		t.Fatalf("same service twice should not satisfy MinSources=2, got %v", got)
	}
}

func TestFlush_ExpiredGroups_Returned(t *testing.T) {
	opts := DefaultOptions()
	opts.Window = 10 * time.Millisecond
	c := New(opts)
	c.Record(makeEntry("svc-a", "req-expired"))

	time.Sleep(20 * time.Millisecond)

	flushed := c.Flush()
	if len(flushed) == 0 {
		t.Fatal("expected at least one expired group")
	}
	if len(flushed[0]) != 1 {
		t.Fatalf("expected 1 entry in flushed group, got %d", len(flushed[0]))
	}
}

func TestFlush_FreshGroups_NotReturned(t *testing.T) {
	opts := DefaultOptions()
	opts.Window = 5 * time.Second
	c := New(opts)
	c.Record(makeEntry("svc-a", "req-fresh"))

	if flushed := c.Flush(); len(flushed) != 0 {
		t.Fatalf("expected no flushed groups, got %d", len(flushed))
	}
}
