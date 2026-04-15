// Package parse provides field-extraction parsing for structured log entries
// using user-defined key=value or JSON fragment patterns.
package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the Parser.
type Options struct {
	// Rules is an ordered list of parsing rules applied to each entry.
	Rules []Rule
}

// Rule describes a single field-extraction rule.
type Rule struct {
	// SourceField is the entry field to parse ("message" or an extra key).
	SourceField string
	// Pattern is a named-capture regular expression.
	Pattern string
	// DestField is the extra field written with the captured value.
	// When empty the capture group name is used directly.
	DestField string

	re *regexp.Regexp
}

// DefaultOptions returns an Options with no rules (pass-through).
func DefaultOptions() Options {
	return Options{}
}

// Parser applies field-extraction rules to log entries.
type Parser struct {
	opts Options
}

// New validates opts and returns a ready Parser.
func New(opts Options) (*Parser, error) {
	for i, r := range opts.Rules {
		if strings.TrimSpace(r.SourceField) == "" {
			return nil, fmt.Errorf("rule %d: SourceField must not be empty", i)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("rule %d: invalid pattern: %w", i, err)
		}
		opts.Rules[i].re = re
	}
	return &Parser{opts: opts}, nil
}

// Apply runs all rules against e and returns a new entry with extracted fields.
func (p *Parser) Apply(e diff.Entry) diff.Entry {
	out := copyEntry(e)
	for _, r := range p.opts.Rules {
		src := fieldValue(e, r.SourceField)
		if src == "" {
			continue
		}
		match := r.re.FindStringSubmatch(src)
		if match == nil {
			continue
		}
		for i, name := range r.re.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			dest := r.DestField
			if dest == "" {
				dest = name
			}
			if out.Extra == nil {
				out.Extra = make(map[string]string)
			}
			out.Extra[dest] = match[i]
		}
	}
	return out
}

func fieldValue(e diff.Entry, field string) string {
	switch strings.ToLower(field) {
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

func copyEntry(e diff.Entry) diff.Entry {
	out := e
	if e.Extra != nil {
		out.Extra = make(map[string]string, len(e.Extra))
		for k, v := range e.Extra {
			out.Extra[k] = v
		}
	}
	return out
}
