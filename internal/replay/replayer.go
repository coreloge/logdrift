// Package replay provides functionality for replaying historical log entries
// from a file, emitting them into a stream at a controlled rate.
package replay

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

// Options configures the replay behaviour.
type Options struct {
	// DelayPerLine is the pause between emitting each log entry.
	// Zero means emit as fast as possible.
	DelayPerLine time.Duration

	// Service is the service label to attach to replayed entries.
	Service string
}

// DefaultOptions returns sensible replay defaults.
func DefaultOptions() Options {
	return Options{
		DelayPerLine: 10 * time.Millisecond,
		Service:      "replay",
	}
}

// Replayer reads log lines from a file and emits parsed entries.
type Replayer struct {
	path string
	opts Options
}

// New creates a Replayer for the given file path.
func New(path string, opts Options) *Replayer {
	if opts.Service == "" {
		opts.Service = "replay"
	}
	return &Replayer{path: path, opts: opts}
}

// Run opens the file and emits parsed log entries to out until EOF or ctx
// is cancelled. The returned channel is closed when Run finishes.
func (r *Replayer) Run(ctx context.Context) (<-chan diff.Entry, error) {
	f, err := os.Open(r.path)
	if err != nil {
		return nil, err
	}

	out := make(chan diff.Entry, 64)

	go func() {
		defer close(out)
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			entry, err := diff.ParseLine(scanner.Text())
			if err != nil {
				continue
			}
			entry.Service = r.opts.Service

			if r.opts.DelayPerLine > 0 {
				select {
				case <-time.After(r.opts.DelayPerLine):
				case <-ctx.Done():
					return
				}
			}

			select {
			case out <- entry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
