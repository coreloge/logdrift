package stamp

import (
	"fmt"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
)

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Field:  "stamped_at",
		Format: time.RFC3339,
	}
}

// Options controls stamper behaviour.
type Options struct {
	// Field is the extra field key to write the timestamp into.
	Field string
	// Format is a Go time layout string used to format the stamp value.
	Format string
	// ReferenceNow, when non-nil, overrides time.Now for testing.
	ReferenceNow func() time.Time
}

// Stamper appends a wall-clock timestamp to every log entry.
type Stamper struct {
	opts Options
	now  func() time.Time
}

// New returns a Stamper or an error if options are invalid.
func New(opts Options) (*Stamper, error) {
	if opts.Field == "" {
		return nil, fmt.Errorf("stamp: field must not be empty")
	}
	if opts.Format == "" {
		return nil, fmt.Errorf("stamp: format must not be empty")
	}
	now := time.Now
	if opts.ReferenceNow != nil {
		now = opts.ReferenceNow
	}
	return &Stamper{opts: opts, now: now}, nil
}

// Apply returns a copy of e with the timestamp field added.
func (s *Stamper) Apply(e diff.Entry) diff.Entry {
	out := copyEntry(e)
	if out.Extra == nil {
		out.Extra = make(map[string]string)
	}
	out.Extra[s.opts.Field] = s.now().Format(s.opts.Format)
	return out
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
