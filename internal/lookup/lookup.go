// Package lookup provides a field-value lookup enrichment stage that resolves
// a named field against a static map and writes the result to an output field.
package lookup

import (
	"errors"
	"fmt"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the Lookup enricher.
type Options struct {
	// SourceField is the entry field whose value is used as the lookup key.
	SourceField string
	// OutputField is the field written with the resolved value.
	OutputField string
	// Table maps lookup keys to replacement values.
	Table map[string]string
	// Default is written to OutputField when no match is found.
	// If empty and no match is found, the entry is forwarded unchanged.
	Default string
}

// DefaultOptions returns a safe zero-value Options.
func DefaultOptions() Options {
	return Options{
		SourceField: "service",
		OutputField: "team",
		Table:       map[string]string{},
	}
}

// Lookup resolves a field value through a static table.
type Lookup struct {
	opts Options
}

// New validates opts and returns a Lookup.
func New(opts Options) (*Lookup, error) {
	if opts.SourceField == "" {
		return nil, errors.New("lookup: SourceField must not be empty")
	}
	if opts.OutputField == "" {
		return nil, errors.New("lookup: OutputField must not be empty")
	}
	if opts.Table == nil {
		return nil, errors.New("lookup: Table must not be nil")
	}
	return &Lookup{opts: opts}, nil
}

// Apply resolves the source field and writes the result to the output field.
// A copy of the entry is always returned; the original is never mutated.
func (l *Lookup) Apply(e diff.Entry) diff.Entry {
	out := copyEntry(e)
	key := fieldValue(e, l.opts.SourceField)
	if resolved, ok := l.opts.Table[key]; ok {
		if out.Extra == nil {
			out.Extra = map[string]string{}
		}
		out.Extra[l.opts.OutputField] = resolved
		return out
	}
	if l.opts.Default != "" {
		if out.Extra == nil {
			out.Extra = map[string]string{}
		}
		out.Extra[l.opts.OutputField] = l.opts.Default
	}
	return out
}

func fieldValue(e diff.Entry, field string) string {
	switch field {
	case "service":
		return e.Service
	case "level":
		return e.Level
	case "message":
		return e.Message
	default:
		if e.Extra != nil {
			return e.Extra[field]
		}
		return ""
	}
}

func copyEntry(e diff.Entry) diff.Entry {
	out := e
	if e.Extra != nil {
		out.Extra = make(map[string]string, len(e.Extra))
		for k, v := range e.Extra {
			out.Extra[k] = fmt.Sprintf("%s", v)
		}
	}
	return out
}
