package score

import (
	"testing"

	"github.com/logdrift/logdrift/internal/diff"
)

func makeResult(fields ...string) diff.Result {
	var deltas []diff.Delta
	for _, f := range fields {
		deltas = append(deltas, diff.Delta{Field: f, A: "x", B: "y"})
	}
	return diff.Result{Deltas: deltas}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.LevelWeight <= 0 || opts.MessageWeight <= 0 || opts.FieldWeight <= 0 {
		t.Fatal("expected positive default weights")
	}
	if opts.MaxScore <= 0 {
		t.Fatal("expected positive MaxScore")
	}
}

func TestNew_NegativeWeight_ReturnsError(t *testing.T) {
	opts := DefaultOptions()
	opts.LevelWeight = -1
	_, err := New(opts)
	if err == nil {
		t.Fatal("expected error for negative weight")
	}
}

func TestNew_AllZeroWeights_ReturnsError(t *testing.T) {
	_, err := New(Options{MaxScore: 1.0})
	if err == nil {
		t.Fatal("expected error when all weights are zero")
	}
}

func TestNew_ZeroMaxScore_ReturnsError(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxScore = 0
	_, err := New(opts)
	if err == nil {
		t.Fatal("expected error for zero MaxScore")
	}
}

func TestCompute_NoDelta_ReturnsZero(t *testing.T) {
	s, _ := New(DefaultOptions())
	if got := s.Compute(makeResult()); got != 0 {
		t.Fatalf("expected 0, got %f", got)
	}
}

func TestCompute_LevelDelta_ReturnsLevelWeight(t *testing.T) {
	opts := DefaultOptions()
	s, _ := New(opts)
	got := s.Compute(makeResult("level"))
	if got != opts.LevelWeight {
		t.Fatalf("expected %f, got %f", opts.LevelWeight, got)
	}
}

func TestCompute_CappedAtMaxScore(t *testing.T) {
	opts := Options{LevelWeight: 5, MessageWeight: 5, FieldWeight: 5, MaxScore: 1.0}
	s, _ := New(opts)
	got := s.Compute(makeResult("level", "message", "service"))
	if got != opts.MaxScore {
		t.Fatalf("expected max %f, got %f", opts.MaxScore, got)
	}
}

func TestCompute_MultipleDifferentFields(t *testing.T) {
	opts := Options{LevelWeight: 0.3, MessageWeight: 0.4, FieldWeight: 0.1, MaxScore: 1.0}
	s, _ := New(opts)
	got := s.Compute(makeResult("level", "message"))
	want := opts.LevelWeight + opts.MessageWeight
	if got != want {
		t.Fatalf("expected %f, got %f", want, got)
	}
}
