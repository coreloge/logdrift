// Package unwrap extracts a nested JSON object from a named field and
// promotes its keys to the top-level entry, optionally removing the
// source field afterwards.
package unwrap

import (
	"errors"
	"fmt"

	"github.com/yourorg/logdrift/internal/diff"
)

// Options controls Unwrapper behaviour.
type Options struct {
	// Field is the Extra key whose value is a map[string]any to unwrap.
	Field string
	// Prefix is prepended to each promoted key (may be empty).
	Prefix string
	// Remove deletes the source field after unwrapping.
	Remove bool
	// Overwrite allows promoted keys to overwrite existing Extra keys.
	Overwrite bool
}

// DefaultOptions returns safe defaults.
func DefaultOptions() Options {
	return Options{
		Remove:    true,
		Overwrite: false,
	}
}

// Unwrapper promotes nested map fields to the top-level entry Extra map.
type Unwrapper structpts Options
}

// New creates an Unwrapper. Returns an error if Field is empty.
func New(opts Options) (*Unwrapper, error) {
	if opts.Field == "" {
		return nil, errors.New("unwrap: Field must not be empty")
	}
	return &Unwrapper{opts: opts}, nil
}

// Apply returns a copy of e with the nested field unwrapped.
// If the field is absent or not a map the entry is returned unchanged.
func (u *Unwrapper) Apply(e diff.Entry) diff.Entry {
	out := copyEntry(e)

	raw, ok := out.Extra[u.opts.Field]
	if !ok {
		return out
	}
	nested, ok := raw.(map[string]any)
	if !ok {
		return out
	}

	for k, v := range nested {
		dest := fmt.Sprintf("%s%s", u.opts.Prefix, k)
		if _, exists := out.Extra[dest]; exists && !u.opts.Overwrite {
			continue
		}
		out.Extra[dest] = v
	}

	if u.opts.Remove {
		delete(out.Extra, u.opts.Field)
	}
	return out
}

func copyEntry(e diff.Entry) diff.Entry {
	out := e
	out.Extra = make(map[string]any, len(e.Extra))
	for k, v := range e.Extra {
		out.Extra[k] = v
	}
	return out
}
