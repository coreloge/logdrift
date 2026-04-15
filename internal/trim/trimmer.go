// Package trim provides a pipeline stage that removes leading and trailing
// whitespace from log entry fields before they reach downstream consumers.
package trim

import (
	"context"
	"strings"

	"logdrift/internal/diff"
)

// Options controls which fields are trimmed.
type Options struct {
	// Fields lists the entry fields to trim. If empty, only Message is trimmed.
	// Use "message", "level", "service", or any key in Extra.
	Fields []string
	// TrimExtra trims all keys in the Extra map when true.
	TrimExtra bool
}

// DefaultOptions returns an Options that trims only the Message field.
func DefaultOptions() Options {
	return Options{
		Fields:    []string{"message"},
		TrimExtra: false,
	}
}

// Trimmer applies whitespace trimming to log entries.
type Trimmer struct {
	opts Options
}

// New creates a Trimmer with the provided Options.
func New(opts Options) *Trimmer {
	return &Trimmer{opts: opts}
}

// Apply returns a copy of entry with the configured fields trimmed.
func (t *Trimmer) Apply(entry diff.Entry) diff.Entry {
	out := diff.Entry{
		Service:   entry.Service,
		Timestamp: entry.Timestamp,
		Level:     entry.Level,
		Message:   entry.Message,
		Extra:     copyExtra(entry.Extra),
	}

	for _, f := range t.opts.Fields {
		switch strings.ToLower(f) {
		case "message":
			out.Message = strings.TrimSpace(out.Message)
		case "level":
			out.Level = strings.TrimSpace(out.Level)
		case "service":
			out.Service = strings.TrimSpace(out.Service)
		default:
			if v, ok := out.Extra[f]; ok {
				out.Extra[f] = strings.TrimSpace(v)
			}
		}
	}

	if t.opts.TrimExtra {
		for k, v := range out.Extra {
			out.Extra[k] = strings.TrimSpace(v)
		}
	}

	return out
}

// Stream reads entries from in, trims them, and forwards to the returned channel.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) <-chan diff.Entry {
	out := make(chan diff.Entry)
	t := New(opts)
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
				out <- t.Apply(entry)
			}
		}
	}()
	return out
}

func copyExtra(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
