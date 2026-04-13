// Package classify provides log entry classification based on field patterns.
// Entries are assigned a category string that downstream stages can use for
// routing, aggregation, or display purposes.
package classify

import (
	"fmt"
	"regexp"

	"github.com/user/logdrift/internal/diff"
)

// Rule maps a compiled pattern against a named field; when the field value
// matches, the entry is tagged with Category.
type Rule struct {
	Field    string
	Pattern  *regexp.Regexp
	Category string
}

// Options configures the Classifier.
type Options struct {
	// Rules are evaluated in order; the first match wins.
	Rules []Rule
	// DefaultCategory is applied when no rule matches. Empty string means
	// no category label is added.
	DefaultCategory string
	// OutputField is the entry field that receives the category value.
	// Defaults to "category".
	OutputField string
}

// DefaultOptions returns a safe zero-value Options.
func DefaultOptions() Options {
	return Options{
		OutputField: "category",
	}
}

// Classifier assigns a category to each log entry.
type Classifier struct {
	opts Options
}

// New constructs a Classifier. Returns an error if OutputField is empty.
func New(opts Options) (*Classifier, error) {
	if opts.OutputField == "" {
		return nil, fmt.Errorf("classify: OutputField must not be empty")
	}
	return &Classifier{opts: opts}, nil
}

// Apply evaluates the entry against all rules and returns a shallow copy with
// the category field set. The original entry is never mutated.
func (c *Classifier) Apply(e diff.Entry) diff.Entry {
	category := c.opts.DefaultCategory

	for _, rule := range c.opts.Rules {
		var val string
		switch rule.Field {
		case "level":
			val = e.Level
		case "message":
			val = e.Message
		default:
			val = e.Fields[rule.Field]
		}
		if rule.Pattern.MatchString(val) {
			category = rule.Category
			break
		}
	}

	if category == "" {
		return e
	}

	// Copy fields map to avoid mutating the original.
	fields := make(map[string]string, len(e.Fields)+1)
	for k, v := range e.Fields {
		fields[k] = v
	}
	fields[c.opts.OutputField] = category

	out := e
	out.Fields = fields
	return out
}
