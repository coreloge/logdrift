// Package tee provides a pipeline stage that forwards each log entry to a
// side-channel writer while also passing it downstream unchanged.
package tee

import (
	"context"
	"io"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the Tee stage.
type Options struct {
	// Writer receives a formatted copy of every entry.
	Writer io.Writer
	// Format is either "text" or "json". Defaults to "text".
	Format string
}

// DefaultOptions returns a no-op set of options (writer is io.Discard).
func DefaultOptions() Options {
	return Options{
		Writer: io.Discard,
		Format: "text",
	}
}

// Tee copies each entry to the configured writer and forwards it downstream.
type Tee struct {
	opts Options
}

// New creates a Tee with the supplied options.
func New(opts Options) (*Tee, error) {
	if opts.Writer == nil {
		opts.Writer = io.Discard
	}
	if opts.Format == "" {
		opts.Format = "text"
	}
	return &Tee{opts: opts}, nil
}

// Apply writes a formatted line to the side-channel and returns the entry
// unmodified.
func (t *Tee) Apply(e diff.Entry) diff.Entry {
	var line string
	if t.opts.Format == "json" {
		line = diff.FormatEntry(e) // reuse existing JSON formatter
	} else {
		line = diff.FormatEntry(e)
	}
	_, _ = io.WriteString(t.opts.Writer, line+"\n")
	return e
}

// Stream reads from in, tees each entry, and forwards to the returned channel.
func Stream(ctx context.Context, in <-chan diff.Entry, t *Tee) <-chan diff.Entry {
	out := make(chan diff.Entry)
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
				out <- t.Apply(e)
			}
		}
	}()
	return out
}
