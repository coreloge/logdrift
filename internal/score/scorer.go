package score

import (
	"errors"
	"math"

	"github.com/logdrift/logdrift/internal/diff"
)

// DefaultOptions returns a Scorer with sensible defaults.
func DefaultOptions() Options {
	return Options{
		LevelWeight:   0.4,
		MessageWeight: 0.4,
		FieldWeight:   0.2,
		MaxScore:      1.0,
	}
}

// Options controls how drift scores are calculated.
type Options struct {
	// LevelWeight is the contribution of a level mismatch to the score.
	LevelWeight float64
	// MessageWeight is the contribution of a message mismatch.
	MessageWeight float64
	// FieldWeight is the contribution of extra-field mismatches.
	FieldWeight float64
	// MaxScore is the ceiling returned for any single comparison.
	MaxScore float64
}

// Scorer computes a numeric drift score from a diff.Result.
type Scorer struct {
	opts Options
}

// New creates a Scorer with the provided Options.
// It returns an error if any weight is negative or weights sum to zero.
func New(opts Options) (*Scorer, error) {
	if opts.LevelWeight < 0 || opts.MessageWeight < 0 || opts.FieldWeight < 0 {
		return nil, errors.New("score: weights must be non-negative")
	}
	if opts.LevelWeight+opts.MessageWeight+opts.FieldWeight == 0 {
		return nil, errors.New("score: at least one weight must be non-zero")
	}
	if opts.MaxScore <= 0 {
		return nil, errors.New("score: MaxScore must be positive")
	}
	return &Scorer{opts: opts}, nil
}

// Compute returns a drift score in the range [0, MaxScore] for the given result.
// A score of 0 means no drift; higher values indicate more divergence.
func (s *Scorer) Compute(r diff.Result) float64 {
	var raw float64
	for _, d := range r.Deltas {
		switch d.Field {
		case "level":
			raw += s.opts.LevelWeight
		case "message":
			raw += s.opts.MessageWeight
		default:
			raw += s.opts.FieldWeight
		}
	}
	return math.Min(raw, s.opts.MaxScore)
}
