// Package label provides field-based labelling for log entries,
// allowing static or derived tags to be attached to entries as they
// pass through the pipeline.
package label

import (
	"context"
	"strings"

	"github.com/yourorg/logdrift/internal/diff"
)

// Options controls how the Labeler behaves.
type Options struct {
	// Labels is a map of field name → static value applied to every entry.
	Labels map[string]string

	// Prefix is prepended to every label key before it is written.
	Prefix string

	// Overwrite controls whether existing fields are replaced.
	Overwrite bool
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Labels:    map[string]string{},
		Prefix:    "",
		Overwrite: false,
	}
}

// Labeler applies a fixed set of labels to log entries.
type Labeler struct {
	opts Options
}

// New creates a Labeler with the provided options.
func New(opts Options) *Labeler {
	return &Labeler{opts: opts}
}

// Apply returns a copy of entry with the configured labels attached.
// Fields that already exist are only overwritten when opts.Overwrite is true.
func (l *Labeler) Apply(entry diff.Entry) diff.Entry {
	if len(l.opts.Labels) == 0 {
		return entry
	}

	out := diff.Entry{
		Service:   entry.Service,
		Level:     entry.Level,
		Message:   entry.Message,
		Timestamp: entry.Timestamp,
		Fields:    make(map[string]string, len(entry.Fields)+len(l.opts.Labels)),
	}
	for k, v := range entry.Fields {
		out.Fields[k] = v
	}

	for k, v := range l.opts.Labels {
		key := k
		if l.opts.Prefix != "" {
			key = strings.TrimRight(l.opts.Prefix, ".") + "." + k
		}
		if _, exists := out.Fields[key]; exists && !l.opts.Overwrite {
			continue
		}
		out.Fields[key] = v
	}
	return out
}

// Stream reads entries from in, applies labels, and forwards them to the
// returned channel. It closes the output channel when ctx is cancelled or
// in is closed.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) <-chan diff.Entry {
	out := make(chan diff.Entry)
	l := New(opts)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-in:
				if !ok {
					return
				}
				out <- l.Apply(entry)
			}
		}
	}()
	return out
}
