package render_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/render"
)

func baseEntry(level, msg string) diff.Entry {
	return diff.Entry{
		Timestamp: time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC),
		Level:     level,
		Message:   msg,
	}
}

func TestEntry_ContainsServiceAndMessage(t *testing.T) {
	var buf bytes.Buffer
	r := render.New(&buf, true)
	r.Entry("api", baseEntry("info", "server started"))
	out := buf.String()
	if !strings.Contains(out, "[api]") {
		t.Errorf("expected service tag in output, got: %s", out)
	}
	if !strings.Contains(out, "server started") {
		t.Errorf("expected message in output, got: %s", out)
	}
}

func TestEntry_ContainsLevel(t *testing.T) {
	var buf bytes.Buffer
	r := render.New(&buf, true)
	r.Entry("worker", baseEntry("error", "crashed"))
	out := buf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected level ERROR in output, got: %s", out)
	}
}

func TestEntry_ContainsTimestamp(t *testing.T) {
	var buf bytes.Buffer
	r := render.New(&buf, true)
	r.Entry("api", baseEntry("info", "server started"))
	out := buf.String()
	// baseEntry uses 2024-01-02 15:04:05 UTC
	if !strings.Contains(out, "2024") {
		t.Errorf("expected timestamp year in output, got: %s", out)
	}
}

func TestDiffResult_Equal(t *testing.T) {
	var buf bytes.Buffer
	r := render.New(&buf, true)
	r.DiffResult(diff.Result{Equal: true})
	out := buf.String()
	if !strings.Contains(out, "match") {
		t.Errorf("expected match indicator, got: %s", out)
	}
}

func TestDiffResult_Deltas(t *testing.T) {
	var buf bytes.Buffer
	r := render.New(&buf, true)
	result := diff.Result{
		Equal: false,
		Deltas: []diff.Delta{
			{Field: "level", A: "info", B: "error"},
		},
	}
	r.DiffResult(result)
	out := buf.String()
	if !strings.Contains(out, "level") {
		t.Errorf("expected field name in diff output, got: %s", out)
	}
}

func TestNew_NilWriter_UsesStdout(t *testing.T) {
	// Should not panic when out is nil.
	r := render.New(nil, true)
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}
