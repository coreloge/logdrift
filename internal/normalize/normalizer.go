// Package normalize provides field-level normalization for log entries,
// allowing values to be coerced into canonical forms (e.g. lower-casing
// log levels, trimming whitespace, collapsing repeated spaces).
package normalize

import (
	"strings"

	"github.com/yourorg/logdrift/internal/diff"
)

// Op is a normalization operation applied to a field value.
type Op string

const (
	OpLower    Op = "lower"    // convert to lower-case
	OpUpper    Op = "upper"    // convert to upper-case
	OpTrim     Op = "trim"     // strip leading/trailing whitespace
	OpCollapse Op = "collapse" // collapse internal runs of whitespace to a single space
)

// Rule pairs a field name with the operation to apply to it.
type Rule struct {
	Field string
	Op    Op
}

// Options configures the Normalizer.
type Options struct {
	Rules []Rule
}

// DefaultOptions returns an Options that normalises the standard "level"
// field to lower-case and trims whitespace from "message".
func DefaultOptions() Options {
	return Options{
		Rules: []Rule{
			{Field: "level", Op: OpLower},
			{Field: "message", Op: OpTrim},
		},
	}
}

// Normalizer applies a set of Rules to log entries.
type Normalizer struct {
	opts Options
}

// New creates a Normalizer with the given Options.
func New(opts Options) *Normalizer {
	return &Normalizer{opts: opts}
}

// Apply returns a copy of entry with all configured rules applied.
func (n *Normalizer) Apply(entry diff.Entry) diff.Entry {
	out := diff.Entry{
		Service: entry.Service,
		Timestamp: entry.Timestamp,
		Level:   entry.Level,
		Message: entry.Message,
		Fields:  make(map[string]string, len(entry.Fields)),
	}
	for k, v := range entry.Fields {
		out.Fields[k] = v
	}

	for _, r := range n.opts.Rules {
		switch r.Field {
		case "level":
			out.Level = applyOp(out.Level, r.Op)
		case "message":
			out.Message = applyOp(out.Message, r.Op)
		default:
			if v, ok := out.Fields[r.Field]; ok {
				out.Fields[r.Field] = applyOp(v, r.Op)
			}
		}
	}
	return out
}

func applyOp(s string, op Op) string {
	switch op {
	case OpLower:
		return strings.ToLower(s)
	case OpUpper:
		return strings.ToUpper(s)
	case OpTrim:
		return strings.TrimSpace(s)
	case OpCollapse:
		return strings.Join(strings.Fields(s), " ")
	}
	return s
}
