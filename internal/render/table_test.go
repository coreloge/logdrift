package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/snapshot"
)

func makeCompareResult(rows []snapshot.CompareRow) snapshot.CompareResult {
	return snapshot.CompareResult{Rows: rows}
}

func TestRenderSnapshot_NoRows(t *testing.T) {
	var buf bytes.Buffer
	tr := NewTableRenderer(&buf)

	if err := tr.RenderSnapshot(makeCompareResult(nil)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "SERVICE A") {
		t.Errorf("expected header in output, got:\n%s", out)
	}
}

func TestRenderSnapshot_MatchRow(t *testing.T) {
	var buf bytes.Buffer
	tr := NewTableRenderer(&buf)

	rows := []snapshot.CompareRow{
		{ServiceA: "api", ServiceB: "worker", Deltas: nil},
	}
	if err := tr.RenderSnapshot(makeCompareResult(rows)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "MATCH") {
		t.Errorf("expected MATCH status in output, got:\n%s", out)
	}
	if !strings.Contains(out, "api") || !strings.Contains(out, "worker") {
		t.Errorf("expected service names in output, got:\n%s", out)
	}
}

func TestRenderSnapshot_DiffRow(t *testing.T) {
	var buf bytes.Buffer
	tr := NewTableRenderer(&buf)

	rows := []snapshot.CompareRow{
		{
			ServiceA: "api",
			ServiceB: "worker",
			Deltas: []diff.Delta{
				{Field: "level", A: "info", B: "error"},
			},
		},
	}
	if err := tr.RenderSnapshot(makeCompareResult(rows)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "DIFF") {
		t.Errorf("expected DIFF status in output, got:\n%s", out)
	}
	if !strings.Contains(out, "level") {
		t.Errorf("expected delta field 'level' in output, got:\n%s", out)
	}
}
