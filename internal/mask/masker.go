// Package mask provides field-level value masking for log entries,
// replacing sensitive values with a fixed placeholder or a custom string.
package mask

import (
	"strings"

	"github.com/yourorg/logdrift/internal/diff"
)

const defaultPlaceholder = "***"

// Options controls which fields are masked and how.
type Options struct {
	// Fields is the list of field names (case-insensitive) whose values will be masked.
	Fields []string
	// Placeholder replaces the original value. Defaults to "***".
	Placeholder string
}

// DefaultOptions returns an Options with no fields masked.
func DefaultOptions() Options {
	return Options{
		Placeholder: defaultPlaceholder,
	}
}

// Masker applies value masking to log entries.
type Masker struct {
	opts   Options
	fields map[string]struct{}
}

// New creates a Masker from the given Options.
func New(opts Options) *Masker {
	if opts.Placeholder == "" {
		opts.Placeholder = defaultPlaceholder
	}
	fields := make(map[string]struct{}, len(opts.Fields))
	for _, f := range opts.Fields {
		fields[strings.ToLower(f)] = struct{}{}
	}
	return &Masker{opts: opts, fields: fields}
}

// Apply returns a copy of entry with configured fields masked.
// The original entry is never mutated.
func (m *Masker) Apply(entry diff.Entry) diff.Entry {
	if len(m.fields) == 0 {
		return entry
	}

	_, maskMsg := m.fields["message"]
	_, maskLvl := m.fields["level"]

	out := diff.Entry{
		Service: entry.Service,
		Level:   entry.Level,
		Message: entry.Message,
		Fields:  make(map[string]string, len(entry.Fields)),
	}

	if maskMsg {
		out.Message = m.opts.Placeholder
	}
	if maskLvl {
		out.Level = m.opts.Placeholder
	}

	for k, v := range entry.Fields {
		if _, ok := m.fields[strings.ToLower(k)]; ok {
			out.Fields[k] = m.opts.Placeholder
		} else {
			out.Fields[k] = v
		}
	}
	return out
}
