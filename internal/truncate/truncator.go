// Package truncate provides utilities for truncating long field values
// in log entries before they are rendered or written to output.
package truncate

import (
	"context"
	"unicode/utf8"

	"logdrift/internal/diff"
)

// DefaultOptions returns a sane set of truncation defaults.
func DefaultOptions() Options {
	return Options{
		MaxMessageLen: 256,
		MaxFieldLen:   128,
		Ellipsis:      "…",
	}
}

// Options controls how field values are truncated.
type Options struct {
	// MaxMessageLen is the maximum rune length of the message field.
	// Zero means no limit.
	MaxMessageLen int
	// MaxFieldLen is the maximum rune length for every other string field.
	// Zero means no limit.
	MaxFieldLen int
	// Ellipsis is appended when a value is truncated.
	Ellipsis string
}

// Truncator applies length limits to log entry fields.
type Truncator struct {
	opts Options
}

// New creates a Truncator with the provided options.
func New(opts Options) *Truncator {
	return &Truncator{opts: opts}
}

// Apply returns a copy of entry with fields truncated according to the
// configured limits. The original entry is never mutated.
func (t *Truncator) Apply(e diff.Entry) diff.Entry {
	out := e
	out.Fields = make(map[string]string, len(e.Fields))
	for k, v := range e.Fields {
		out.Fields[k] = v
	}

	if t.opts.MaxMessageLen > 0 {
		out.Message = truncate(e.Message, t.opts.MaxMessageLen, t.opts.Ellipsis)
	}
	for k, v := range out.Fields {
		if t.opts.MaxFieldLen > 0 {
			out.Fields[k] = truncate(v, t.opts.MaxFieldLen, t.opts.Ellipsis)
		}
	}
	return out
}

// Stream reads entries from in, applies truncation, and forwards results to
// the returned channel. The channel is closed when ctx is cancelled or in
// is closed.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) <-chan diff.Entry {
	out := make(chan diff.Entry)
	t := New(opts)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- t.Apply(e):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

func truncate(s string, max int, ellipsis string) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + ellipsis
}
