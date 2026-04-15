// Package reorder provides a log entry reordering buffer that emits entries
// in timestamp order after collecting them within a configurable hold window.
package reorder

import (
	"context"
	"sort"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// DefaultOptions returns a sensible default configuration.
func DefaultOptions() Options {
	return Options{
		HoldWindow: 200 * time.Millisecond,
		MaxBuffer:  512,
	}
}

// Options controls reorder behaviour.
type Options struct {
	// HoldWindow is how long entries are buffered before being flushed in order.
	HoldWindow time.Duration
	// MaxBuffer is the maximum number of entries held before a forced flush.
	MaxBuffer int
}

// Reorderer buffers incoming log entries and emits them in timestamp order.
type Reorderer struct {
	opts Options
}

// New creates a Reorderer with the given options.
func New(opts Options) (*Reorderer, error) {
	if opts.HoldWindow <= 0 {
		opts.HoldWindow = DefaultOptions().HoldWindow
	}
	if opts.MaxBuffer <= 0 {
		opts.MaxBuffer = DefaultOptions().MaxBuffer
	}
	return &Reorderer{opts: opts}, nil
}

// Stream reads from in, buffers entries for HoldWindow, then emits them sorted
// by timestamp. The output channel is closed when ctx is cancelled.
func (r *Reorderer) Stream(ctx context.Context, in <-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry, r.opts.MaxBuffer)
	go func() {
		defer close(out)
		buf := make([]diff.Entry, 0, r.opts.MaxBuffer)
		ticker := time.NewTicker(r.opts.HoldWindow)
		defer ticker.Stop()

		flush := func() {
			if len(buf) == 0 {
				return
			}
			sort.Slice(buf, func(i, j int) bool {
				return buf[i].Timestamp.Before(buf[j].Timestamp)
			})
			for _, e := range buf {
				select {
				case out <- e:
				case <-ctx.Done():
					return
				}
			}
			buf = buf[:0]
		}

		for {
			select {
			case <-ctx.Done():
				flush()
				return
			case e, ok := <-in:
				if !ok {
					flush()
					return
				}
				buf = append(buf, e)
				if len(buf) >= r.opts.MaxBuffer {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()
	return out
}
