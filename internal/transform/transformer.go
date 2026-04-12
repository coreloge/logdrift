// Package transform provides field-level transformation of log entries
// before they are passed downstream in a pipeline.
package transform

import (
	"fmt"
	"strings"

	"logdrift/internal/diff"
)

// Op is the type of transformation to apply.
type Op string

const (
	OpUppercase Op = "uppercase"
	OpLowercase Op = "lowercase"
	OpPrefix    Op = "prefix"
	OpSuffix    Op = "suffix"
	OpTruncate  Op = "truncate"
)

// Rule describes a single transformation applied to a named field.
type Rule struct {
	Field  string
	Op     Op
	// Arg is used by Prefix, Suffix, and Truncate operations.
	Arg    string
}

// Options configures the Transformer.
type Options struct {
	Rules []Rule
}

// DefaultOptions returns an Options with no rules.
func DefaultOptions() Options {
	return Options{}
}

// Transformer applies a set of field transformation rules to log entries.
type Transformer struct {
	opts Options
}

// New creates a Transformer with the provided options.
func New(opts Options) *Transformer {
	return &Transformer{opts: opts}
}

// Apply returns a copy of entry with all matching rules applied.
func (t *Transformer) Apply(entry diff.Entry) diff.Entry {
	out := diff.Entry{
		Service:  entry.Service,
		Level:    entry.Level,
		Message:  entry.Message,
		Fields:   make(map[string]string, len(entry.Fields)),
	}
	for k, v := range entry.Fields {
		out.Fields[k] = v
	}

	for _, rule := range t.opts.Rules {
		switch rule.Field {
		case "message":
			out.Message = applyOp(out.Message, rule)
		case "level":
			out.Level = applyOp(out.Level, rule)
		default:
			if v, ok := out.Fields[rule.Field]; ok {
				out.Fields[rule.Field] = applyOp(v, rule)
			}
		}
	}
	return out
}

func applyOp(value string, rule Rule) string {
	switch rule.Op {
	case OpUppercase:
		return strings.ToUpper(value)
	case OpLowercase:
		return strings.ToLower(value)
	case OpPrefix:
		return rule.Arg + value
	case OpSuffix:
		return value + rule.Arg
	case OpTruncate:
		n := 0
		fmt.Sscanf(rule.Arg, "%d", &n)
		if n > 0 && len(value) > n {
			return value[:n]
		}
		return value
	}
	return value
}
