package metrics

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/user/logdrift/internal/diff"
)

func buildCounter(t *testing.T) *Counter {
	t.Helper()
	c := New()
	c.RecordEntry("api")
	c.RecordEntry("api")
	c.RecordEntry("worker")
	c.RecordDrift("api", diff.DiffResult{
		Deltas: []diff.Delta{{Field: "level", A: "info", B: "warn"}},
	})
	return c
}

func TestNewExporter_DefaultsToText(t *testing.T) {
	c := New()
	e := NewExporter(c, "")
	if e.format != "text" {
		t.Errorf("expected text, got %q", e.format)
	}
}

func TestExport_TextFormat(t *testing.T) {
	c := buildCounter(t)
	var buf bytes.Buffer
	e := NewExporter(c, "text")
	if err := e.Export(&buf); err != nil {
		t.Fatalf("Export error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"api", "worker", "TOTAL", "ENTRIES", "DRIFTS"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\noutput:\n%s", want, out)
		}
	}
}

func TestExport_JSONFormat(t *testing.T) {
	c := buildCounter(t)
	var buf bytes.Buffer
	e := NewExporter(c, "json")
	if err := e.Export(&buf); err != nil {
		t.Fatalf("Export error: %v", err)
	}
	var snap Snapshot
	if err := json.Unmarshal(buf.Bytes(), &snap); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if snap.TotalEntries != 3 {
		t.Errorf("expected TotalEntries=3, got %d", snap.TotalEntries)
	}
	if snap.TotalDrifts != 1 {
		t.Errorf("expected TotalDrifts=1, got %d", snap.TotalDrifts)
	}
	if snap.Entries["api"] != 2 {
		t.Errorf("expected api entries=2, got %d", snap.Entries["api"])
	}
}

func TestExport_EmptyCounter_JSON(t *testing.T) {
	c := New()
	var buf bytes.Buffer
	e := NewExporter(c, "json")
	if err := e.Export(&buf); err != nil {
		t.Fatalf("Export error: %v", err)
	}
	var snap Snapshot
	if err := json.Unmarshal(buf.Bytes(), &snap); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if snap.TotalEntries != 0 || snap.TotalDrifts != 0 {
		t.Errorf("expected zeroed totals, got entries=%d drifts=%d",
			snap.TotalEntries, snap.TotalDrifts)
	}
}

func TestExport_UnknownFormat(t *testing.T) {
	c := New()
	var buf bytes.Buffer
	e := NewExporter(c, "xml")
	if err := e.Export(&buf); err == nil {
		t.Error("expected error for unknown format, got nil")
	}
}
