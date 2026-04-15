// Package extract provides field extraction from log entries into
// structured key-value pairs based on regex capture groups.
package extract

import (
	"fmt"
	"regexp"

	"github.com/user/logdrift/internal/diff"
)

// Rule defines a single extraction rule: a named regex applied to a
// source field, with capture groups written into the entry's Extra map.
type Rule struct {
	// SourceField is the entry field to match against ("message", or an Extra key).
	SourceField string
	// Pattern is the regular expression; named groups become output fields.
	Pattern string
	// OutputPrefix is prepended to every captured group name.
	OutputPrefix string

	re *regexp.Regexp
}

// Options configures the Extractor.
type Options struct {
	Rules []Rule
}

// DefaultOptions returns an Options with no rules.
func DefaultOptions() Options {
	return Options{}
}

// Extractor applies extraction rules to log entries.
type Extractor struct {
	opts Options
}

// New compiles all rule patterns and returns an Extractor.
// Returns an error if any pattern fails to compile.
func New(opts Options) (*Extractor, error) {
	for i, r := range opts.Rules {
		if r.SourceField == "" {
			return nil, fmt.Errorf("rule %d: source_field must not be empty", i)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("rule %d: invalid pattern %q: %w", i, r.Pattern, err)
		}
		opts.Rules[i].re = re
	}
	return &Extractor{opts: opts}, nil
}

// Apply runs all extraction rules against entry and returns a new entry
// with any captured fields merged into Extra. The original is not mutated.
func (e *Extractor) Apply(entry diff.LogEntry) diff.LogEntry {
	out := copyEntry(entry)
	for _, rule := range e.opts.Rules {
		src := fieldValue(entry, rule.SourceField)
		if src == "" {
			continue
		}
		match := rule.re.FindStringSubmatch(src)
		if match == nil {
			continue
		}
		for j, name := range rule.re.SubexpNames() {
			if j == 0 || name == "" {
				continue
			}
			key := rule.OutputPrefix + name
			out.Extra[key] = match[j]
		}
	}
	return out
}

func fieldValue(e diff.LogEntry, field string) string {
	switch field {
	case "message":
		return e.Message
	case "level":
		return e.Level
	case "service":
		return e.Service
	default:
		return e.Extra[field]
	}
}

func copyEntry(e diff.LogEntry) diff.LogEntry {
	out := e
	out.Extra = make(map[string]string, len(e.Extra))
	for k, v := range e.Extra {
		out.Extra[k] = v
	}
	return out
}
