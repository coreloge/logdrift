// Package annotate attaches structured annotations to log entries based on
// configurable rules. Annotations are written into the entry's Extra map
// under a caller-specified key prefix.
package annotate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yourorg/logdrift/internal/diff"
)

// Rule describes a single annotation rule.
type Rule struct {
	// Field is the entry field to match against ("message", "level", or an Extra key).
	Field string
	// Pattern is the regular expression that must match the field value.
	Pattern string
	// Annotation is the key=value pair written into Extra when the rule matches.
	AnnotationKey   string
	AnnotationValue string

	re *regexp.Regexp
}

// Options configures the Annotator.
type Options struct {
	// Rules is the ordered list of annotation rules to evaluate.
	Rules []Rule
	// Prefix is prepended to every AnnotationKey written into Extra.
	Prefix string
}

// DefaultOptions returns a safe zero-value Options.
func DefaultOptions() Options {
	return Options{
		Prefix: "annotation.",
	}
}

// Annotator applies annotation rules to log entries.
type Annotator struct {
	opts  Options
	rules []Rule
}

// New compiles all rule patterns and returns an Annotator.
// Returns an error if any pattern fails to compile.
func New(opts Options) (*Annotator, error) {
	compiled := make([]Rule, len(opts.Rules))
	for i, r := range opts.Rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("annotate: rule %d invalid pattern %q: %w", i, r.Pattern, err)
		}
		compiled[i] = r
		compiled[i].re = re
	}
	return &Annotator{opts: opts, rules: compiled}, nil
}

// Apply evaluates every rule against entry and writes matching annotations
// into a copy of the entry's Extra map. The original entry is never mutated.
func (a *Annotator) Apply(entry diff.Entry) diff.Entry {
	out := entry
	out.Extra = copyExtra(entry.Extra)

	for _, r := range a.rules {
		val := fieldValue(entry, r.Field)
		if r.re.MatchString(val) {
			key := strings.TrimRight(a.opts.Prefix, ".") + "." + r.AnnotationKey
			out.Extra[key] = r.AnnotationValue
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
	default:
		if e.Extra != nil {
			return e.Extra[field]
		}
		return ""
	}
}

func copyExtra(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src)+4)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
