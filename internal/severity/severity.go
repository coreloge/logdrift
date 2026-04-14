// Package severity provides utilities for ranking and comparing log level
// severity values. It allows callers to determine whether one level is more
// or less severe than another and to normalise free-form level strings into
// canonical ranks.
package severity

import "strings"

// Level represents a canonical severity rank. Higher values are more severe.
type Level int

const (
	Unknown Level = iota
	Debug
	Info
	Warn
	Error
	Fatal
)

// ranks maps lower-cased level strings to their canonical Level.
var ranks = map[string]Level{
	"debug":   Debug,
	"trace":   Debug,
	"verbose": Debug,
	"info":    Info,
	"notice":  Info,
	"warn":    Warn,
	"warning": Warn,
	"error":   Error,
	"err":     Error,
	"fatal":   Fatal,
	"panic":   Fatal,
	"crit":    Fatal,
	"critical":Fatal,
}

// Parse converts a free-form level string into a Level. Unrecognised strings
// return Unknown.
func Parse(s string) Level {
	if l, ok := ranks[strings.ToLower(strings.TrimSpace(s))]; ok {
		return l
	}
	return Unknown
}

// String returns the canonical name for a Level.
func (l Level) String() string {
	switch l {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	case Fatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// AtLeast reports whether l is at least as severe as min.
func (l Level) AtLeast(min Level) bool {
	return l >= min
}

// Compare returns -1, 0, or 1 depending on whether a is less than, equal to,
// or greater than b.
func Compare(a, b Level) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
