// Package tag provides a pipeline stage that attaches dynamic tags to log
// entries based on field value pattern matching.
package tag

import (
	"fmt"
	"regexp"

	"github.com/yourorg/logdrift/internal/diff"
)

// Rule associates a compiled pattern with the tag value to apply when the
// pattern matches the configured source field.
type Rule struct {
	Field   string
	Pattern *regexp.Regexp
	Tag     string
}

// Options configures the Tagger.
type Options struct {
	// OutputField is the entry field that receives the matched tag.
	// Defaults to "tag".
	OutputField string

	// Rules is the ordered list of tagging rules. The first matching rule wins.
	Rules []Rule
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		OutputField: "tag",
	}
}

// Tagger applies tag rules to log entries.
type Tagger struct {
	opts Options
}

// New validates opts and returns a ready-to-use Tagger.
func New(opts Options) (*Tagger, error) {
	if opts.OutputField == "" {
		return nil, fmt.Errorf("tag: OutputField must not be empty")
	}
	return &Tagger{opts: opts}, nil
}

// Apply evaluates each rule against entry and writes the first matching tag
// into the configured output field. If no rule matches the entry is returned
// unchanged.
func (t *Tagger) Apply(entry diff.LogEntry) diff.LogEntry {
	for _, r := range t.opts.Rules {
		var src string
		switch r.Field {
		case "message":
			src = entry.Message
		case "level":
			src = entry.Level
		case "service":
			src = entry.Service
		default:
			if entry.Extra != nil {
				src, _ = entry.Extra[r.Field].(string)
			}
		}
		if r.Pattern.MatchString(src) {
			out := copyEntry(entry)
			out.Extra[t.opts.OutputField] = r.Tag
			return out
		}
	}
	return entry
}

func copyEntry(e diff.LogEntry) diff.LogEntry {
	extra := make(map[string]interface{}, len(e.Extra)+1)
	for k, v := range e.Extra {
		extra[k] = v
	}
	e.Extra = extra
	return e
}
