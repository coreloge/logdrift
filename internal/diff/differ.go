package diff

import (
	"fmt"
	"strings"
	"time"
)

// Entry represents a parsed structured log entry from a service.
type Entry struct {
	Service   string
	Timestamp time.Time
	Level     string
	Message   string
	Raw       string
}

// Delta describes a difference between two log entries.
type Delta struct {
	Field    string
	ServiceA string
	ServiceB string
	ValueA   string
	ValueB   string
}

// Result holds the comparison outcome between two entries.
type Result struct {
	Match  bool
	Deltas []Delta
}

// Compare compares two log entries and returns a Result describing any differences.
func Compare(a, b Entry) Result {
	var deltas []Delta

	if !strings.EqualFold(a.Level, b.Level) {
		deltas = append(deltas, Delta{
			Field:    "level",
			ServiceA: a.Service,
			ServiceB: b.Service,
			ValueA:   a.Level,
			ValueB:   b.Level,
		})
	}

	if a.Message != b.Message {
		deltas = append(deltas, Delta{
			Field:    "message",
			ServiceA: a.Service,
			ServiceB: b.Service,
			ValueA:   a.Message,
			ValueB:   b.Message,
		})
	}

	return Result{
		Match:  len(deltas) == 0,
		Deltas: deltas,
	}
}

// FormatDelta returns a human-readable string for a single Delta.
func FormatDelta(d Delta) string {
	return fmt.Sprintf(
		"[diff] field=%q  %s=%q  %s=%q",
		d.Field, d.ServiceA, d.ValueA, d.ServiceB, d.ValueB,
	)
}

// FormatResult returns a human-readable summary for a Result.
func FormatResult(r Result) string {
	if r.Match {
		return "[diff] entries match"
	}
	lines := make([]string, 0, len(r.Deltas))
	for _, d := range r.Deltas {
		lines = append(lines, FormatDelta(d))
	}
	return strings.Join(lines, "\n")
}
