// Package alert provides threshold-based alerting for log drift detection.
// It watches diff results from snapshot comparisons and emits alerts when
// the number of differing fields across services exceeds a configured threshold.
package alert

import (
	"fmt"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelCrit  Level = "CRIT"
)

// Alert is emitted when a diff result crosses a configured threshold.
type Alert struct {
	Service   string
	Level     Level
	DeltaCount int
	Message   string
	At        time.Time
}

// Config holds thresholds for alert levels.
type Config struct {
	// WarnThreshold triggers a WARN alert when delta count >= value.
	WarnThreshold int
	// CritThreshold triggers a CRIT alert when delta count >= value.
	CritThreshold int
}

// DefaultConfig returns sensible default thresholds.
func DefaultConfig() Config {
	return Config{
		WarnThreshold: 1,
		CritThreshold: 3,
	}
}

// Evaluate inspects a diff.Result and returns an Alert if thresholds are
// breached, or nil if the result is within acceptable bounds.
func Evaluate(service string, result diff.Result, cfg Config) *Alert {
	n := len(result.Deltas)
	if n == 0 {
		return nil
	}

	level := LevelWarn
	if n >= cfg.CritThreshold {
		level = LevelCrit
	} else if n < cfg.WarnThreshold {
		return nil
	}

	return &Alert{
		Service:    service,
		Level:      level,
		DeltaCount: n,
		Message:    fmt.Sprintf("%s: %d field(s) drifted", service, n),
		At:         time.Now(),
	}
}
