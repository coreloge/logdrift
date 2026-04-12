// Package redact provides field-level redaction for structured log entries
// before they are rendered or written to output.
package redact

import (
	"regexp"
	"strings"
)

// Rule describes a single redaction rule.
type Rule struct {
	// Field is the JSON key to match (e.g. "password", "token").
	Field string
	// Pattern, if non-nil, only redacts values matching the regexp.
	Pattern *regexp.Regexp
	// Replacement is the string substituted for the redacted value.
	Replacement string
}

// Options configures the Redactor.
type Options struct {
	Rules       []Rule
	// CaseFold makes field matching case-insensitive.
	CaseFold    bool
}

// DefaultOptions returns an Options with a sensible set of built-in rules.
func DefaultOptions() Options {
	return Options{
		CaseFold: true,
		Rules: []Rule{
			{Field: "password", Replacement: "[REDACTED]"},
			{Field: "token", Replacement: "[REDACTED]"},
			{Field: "secret", Replacement: "[REDACTED]"},
			{Field: "api_key", Replacement: "[REDACTED]"},
			{Field: "authorization", Replacement: "[REDACTED]"},
		},
	}
}

// Redactor applies redaction rules to log field maps.
type Redactor struct {
	opts Options
}

// New creates a Redactor from the given Options.
func New(opts Options) *Redactor {
	return &Redactor{opts: opts}
}

// Apply returns a copy of fields with sensitive values replaced.
// The original map is never mutated.
func (r *Redactor) Apply(fields map[string]string) map[string]string {
	out := make(map[string]string, len(fields))
	for k, v := range fields {
		out[k] = v
	}
	for _, rule := range r.opts.Rules {
		for k, v := range out {
			if r.matchField(k, rule.Field) {
				if rule.Pattern == nil || rule.Pattern.MatchString(v) {
					out[k] = rule.Replacement
				}
			}
		}
	}
	return out
}

func (r *Redactor) matchField(key, ruleField string) bool {
	if r.opts.CaseFold {
		return strings.EqualFold(key, ruleField)
	}
	return key == ruleField
}
