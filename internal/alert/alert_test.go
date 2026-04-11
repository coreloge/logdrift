package alert

import (
	"testing"

	"github.com/user/logdrift/internal/diff"
)

func makeResult(deltas ...diff.Delta) diff.Result {
	return diff.Result{Equal: len(deltas) == 0, Deltas: deltas}
}

func makeDelta(field, a, b string) diff.Delta {
	return diff.Delta{Field: field, A: a, B: b}
}

func TestEvaluate_NoDelta_ReturnsNil(t *testing.T) {
	cfg := DefaultConfig()
	result := makeResult()
	if got := Evaluate("svc", result, cfg); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestEvaluate_BelowWarn_ReturnsNil(t *testing.T) {
	cfg := Config{WarnThreshold: 2, CritThreshold: 4}
	result := makeResult(makeDelta("level", "info", "warn"))
	if got := Evaluate("svc", result, cfg); got != nil {
		t.Fatalf("expected nil for count below warn threshold, got %+v", got)
	}
}

func TestEvaluate_AtWarnThreshold(t *testing.T) {
	cfg := DefaultConfig() // warn=1, crit=3
	result := makeResult(makeDelta("level", "info", "warn"))
	a := Evaluate("api", result, cfg)
	if a == nil {
		t.Fatal("expected alert, got nil")
	}
	if a.Level != LevelWarn {
		t.Errorf("expected WARN, got %s", a.Level)
	}
	if a.Service != "api" {
		t.Errorf("expected service 'api', got %s", a.Service)
	}
	if a.DeltaCount != 1 {
		t.Errorf("expected delta count 1, got %d", a.DeltaCount)
	}
}

func TestEvaluate_AtCritThreshold(t *testing.T) {
	cfg := DefaultConfig() // crit=3
	result := makeResult(
		makeDelta("level", "info", "error"),
		makeDelta("msg", "ok", "fail"),
		makeDelta("code", "200", "500"),
	)
	a := Evaluate("db", result, cfg)
	if a == nil {
		t.Fatal("expected alert, got nil")
	}
	if a.Level != LevelCrit {
		t.Errorf("expected CRIT, got %s", a.Level)
	}
}

func TestEvaluate_MessageContainsService(t *testing.T) {
	cfg := DefaultConfig()
	result := makeResult(makeDelta("level", "info", "warn"))
	a := Evaluate("gateway", result, cfg)
	if a == nil {
		t.Fatal("expected alert")
	}
	if a.Message == "" {
		t.Error("expected non-empty message")
	}
}
