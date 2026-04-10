package diff

import (
	"testing"
	"time"
)

func makeEntry(service, level, message string) Entry {
	return Entry{
		Service:   service,
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Raw:       `{"level":"` + level + `","msg":"` + message + `"}`,
	}
}

func TestCompare_Match(t *testing.T) {
	a := makeEntry("svc-a", "info", "user logged in")
	b := makeEntry("svc-b", "info", "user logged in")

	result := Compare(a, b)
	if !result.Match {
		t.Fatalf("expected match, got deltas: %+v", result.Deltas)
	}
	if len(result.Deltas) != 0 {
		t.Fatalf("expected no deltas, got %d", len(result.Deltas))
	}
}

func TestCompare_LevelDiff(t *testing.T) {
	a := makeEntry("svc-a", "info", "startup complete")
	b := makeEntry("svc-b", "warn", "startup complete")

	result := Compare(a, b)
	if result.Match {
		t.Fatal("expected mismatch on level")
	}
	if len(result.Deltas) != 1 || result.Deltas[0].Field != "level" {
		t.Fatalf("expected level delta, got %+v", result.Deltas)
	}
}

func TestCompare_MessageDiff(t *testing.T) {
	a := makeEntry("svc-a", "error", "connection refused")
	b := makeEntry("svc-b", "error", "timeout exceeded")

	result := Compare(a, b)
	if result.Match {
		t.Fatal("expected mismatch on message")
	}
	if len(result.Deltas) != 1 || result.Deltas[0].Field != "message" {
		t.Fatalf("expected message delta, got %+v", result.Deltas)
	}
}

func TestCompare_MultipleDiffs(t *testing.T) {
	a := makeEntry("svc-a", "info", "request received")
	b := makeEntry("svc-b", "debug", "handling request")

	result := Compare(a, b)
	if result.Match {
		t.Fatal("expected mismatch")
	}
	if len(result.Deltas) != 2 {
		t.Fatalf("expected 2 deltas, got %d", len(result.Deltas))
	}
}

func TestFormatResult_Match(t *testing.T) {
	r := Result{Match: true}
	out := FormatResult(r)
	if out != "[diff] entries match" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestFormatResult_Mismatch(t *testing.T) {
	r := Result{
		Match: false,
		Deltas: []Delta{
			{Field: "level", ServiceA: "svc-a", ServiceB: "svc-b", ValueA: "info", ValueB: "warn"},
		},
	}
	out := FormatResult(r)
	if out == "" {
		t.Fatal("expected non-empty output for mismatch")
	}
}
