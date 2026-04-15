// Package highlight applies ANSI colour highlighting to log entries based on
// configurable field-value rules. It is useful when rendering to a terminal
// and wanting to draw attention to specific log levels or service names.
package highlight

import (
	"fmt"
	"strings"

	"github.com/user/logdrift/internal/diff"
)

// Colour is an ANSI escape colour code.
type Colour string

const (
	Red     Colour = "\033[31m"
	Yellow  Colour = "\033[33m"
	Green   Colour = "\033[32m"
	Blue    Colour = "\033[34m"
	Magenta Colour = "\033[35m"
	Cyan    Colour = "\033[36m"
	Reset   Colour = "\033[0m"
)

// Rule maps a field and its expected value to a display colour.
type Rule struct {
	Field  string
	Value  string
	Colour Colour
}

// Options configures the Highlighter.
type Options struct {
	Rules     []Rule
	CaseFold  bool
}

// DefaultOptions returns sensible defaults that colour common log levels.
func DefaultOptions() Options {
	return Options{
		Rules: []Rule{
			{Field: "level", Value: "error", Colour: Red},
			{Field: "level", Value: "warn", Colour: Yellow},
			{Field: "level", Value: "info", Colour: Green},
			{Field: "level", Value: "debug", Colour: Cyan},
		},
		CaseFold: true,
	}
}

// Highlighter wraps log entries with ANSI colour codes.
type Highlighter struct {
	opts Options
}

// New creates a Highlighter with the given options.
func New(opts Options) *Highlighter {
	return &Highlighter{opts: opts}
}

// Apply returns a copy of entry with its Message wrapped in the first matching
// rule's colour codes. If no rule matches the entry is returned unchanged.
func (h *Highlighter) Apply(entry diff.LogEntry) diff.LogEntry {
	colour := h.match(entry)
	if colour == "" {
		return entry
	}
	out := entry
	out.Message = fmt.Sprintf("%s%s%s", colour, entry.Message, Reset)
	return out
}

func (h *Highlighter) match(entry diff.LogEntry) Colour {
	for _, r := range h.opts.Rules {
		v := fieldValue(entry, r.Field)
		if h.opts.CaseFold {
			if strings.EqualFold(v, r.Value) {
				return r.Colour
			}
		} else if v == r.Value {
			return r.Colour
		}
	}
	return ""
}

func fieldValue(entry diff.LogEntry, field string) string {
	switch field {
	case "level":
		return entry.Level
	case "service":
		return entry.Service
	case "message":
		return entry.Message
	default:
		if entry.Extra != nil {
			if v, ok := entry.Extra[field]; ok {
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return ""
}
