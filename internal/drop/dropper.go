package drop

import (
	"errors"
	"regexp"

	"github.com/logdrift/logdrift/internal/diff"
)

// Rule describes a single drop condition.
type Rule struct {
	Field   string // empty string means match against message
	Pattern string
}

// Options configures the Dropper.
type Options struct {
	Rules []Rule
}

// DefaultOptions returns an Options with no rules (passes everything).
func DefaultOptions() Options {
	return Options{}
}

// compiled holds a compiled rule.
type compiled struct {
	field string
	re    *regexp.Regexp
}

// Dropper drops log entries whose fields match configured patterns.
type Dropper struct {
	rules []compiled
}

// New creates a Dropper from opts. Returns an error if any pattern fails to compile.
func New(opts Options) (*Dropper, error) {
	if len(opts.Rules) == 0 {
		return &Dropper{}, nil
	}
	cs := make([]compiled, 0, len(opts.Rules))
	for _, r := range opts.Rules {
		if r.Pattern == "" {
			return nil, errors.New("drop: empty pattern in rule")
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, err
		}
		cs = append(cs, compiled{field: r.Field, re: re})
	}
	return &Dropper{rules: cs}, nil
}

// ShouldDrop returns true if the entry matches any configured rule.
func (d *Dropper) ShouldDrop(e diff.Entry) bool {
	for _, c := range d.rules {
		var val string
		if c.field == "" || c.field == "message" {
			val = e.Message
		} else if c.field == "level" {
			val = e.Level
		} else if c.field == "service" {
			val = e.Service
		} else if v, ok := e.Extra[c.field]; ok {
			val = v
		}
		if c.re.MatchString(val) {
			return true
		}
	}
	return false
}
