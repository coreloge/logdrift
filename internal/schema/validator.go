// Package schema provides JSON schema validation for structured log entries,
// ensuring that incoming log lines conform to an expected field contract before
// they are processed further in the pipeline.
package schema

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/user/logdrift/internal/diff"
)

// FieldRule describes a single field constraint.
type FieldRule struct {
	// Required indicates the field must be present and non-empty.
	Required bool
	// Pattern is an optional regular expression the field value must match.
	Pattern string

	compiled *regexp.Regexp
}

// Options configures the Validator.
type Options struct {
	// Rules maps field names to their constraints.
	Rules map[string]FieldRule
	// DropInvalid silently drops entries that fail validation instead of
	// returning an error.
	DropInvalid bool
}

// DefaultOptions returns an Options with no rules (pass-through).
func DefaultOptions() Options {
	return Options{
		Rules:       make(map[string]FieldRule),
		DropInvalid: false,
	}
}

// Validator checks log entries against a set of field rules.
type Validator struct {
	opts Options
}

// New creates a Validator, pre-compiling any regex patterns in the rules.
// Returns an error if any pattern fails to compile.
func New(opts Options) (*Validator, error) {
	for name, rule := range opts.Rules {
		if rule.Pattern != "" {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, fmt.Errorf("schema: invalid pattern for field %q: %w", name, err)
			}
			rule.compiled = re
			opts.Rules[name] = rule
		}
	}
	return &Validator{opts: opts}, nil
}

// Validate checks entry against the configured rules.
// Returns a non-nil error describing the first violation found.
func (v *Validator) Validate(entry diff.Entry) error {
	for field, rule := range v.opts.Rules {
		var val string
		switch field {
		case "level":
			val = entry.Level
		case "message":
			val = entry.Message
		default:
			val = entry.Fields[field]
		}

		if rule.Required && val == "" {
			return fmt.Errorf("schema: required field %q is missing or empty", field)
		}
		if rule.compiled != nil && val != "" && !rule.compiled.MatchString(val) {
			return fmt.Errorf("schema: field %q value %q does not match pattern %q", field, val, rule.Pattern)
		}
	}
	return nil
}

// ErrDropped is returned by Stream when DropInvalid is false and validation fails.
var ErrDropped = errors.New("schema: entry dropped due to validation failure")
