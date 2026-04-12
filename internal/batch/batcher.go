// Package batch provides time- and size-based batching of log entries.
// Entries are accumulated and flushed either when the batch reaches a
// configured size or when a configurable timeout elapses.
package batch

import (
	"context"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// DefaultOptions returns a BatchOptions with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxSize: 50,
		FlushInterval: 2 * time.Second,
	}
}

// Options configures the Batcher behaviour.
type Options struct {
	// MaxSize is the maximum number of entries before a flush is forced.
	MaxSize int
	// FlushInterval is the maximum time to wait before flushing a partial batch.
	FlushInterval time.Duration
}

// Batcher accumulates LogEntry values and emits slices on a regular cadence.
type Batcher struct {
	opts Options
}

// New creates a Batcher with the provided Options.
func New(opts Options) *Batcher {
	if opts.MaxSize <= 0 {
		opts.MaxSize = DefaultOptions().MaxSize
	}
	if opts.FlushInterval <= 0 {
		opts.FlushInterval = DefaultOptions().FlushInterval
	}
	return &Batcher{opts: opts}
}

// Stream reads from in and writes batches to the returned channel.
// The returned channel is closed when ctx is cancelled or in is closed.
func (b *Batcher) Stream(ctx context.Context, in <-chan diff.LogEntry) <-chan []diff.LogEntry {
	out := make(chan []diff.LogEntry)
	go func() {
		defer close(out)
		buf := make([]diff.LogEntry, 0, b.opts.MaxSize)
		ticker := time.NewTicker(b.opts.FlushInterval)
		defer ticker.Stop()
		flush := func() {
			if len(buf) == 0 {
				return
			}
			batch := make([]diff.LogEntry, len(buf))
			copy(batch, buf)
			select {
			case out <- batch:
			case <-ctx.Done():
				return
			}
			buf = buf[:0]
		}
		for {
			select {
			case <-ctx.Done():
				flush()
				return
			case entry, ok := <-in:
				if !ok {
					flush()
					return
				}
				buf = append(buf, entry)
				if len(buf) >= b.opts.MaxSize {
					flush()
					ticker.Reset(b.opts.FlushInterval)
				}
			case <-ticker.C:
				flush()
			}
		}
	}()
	return out
}
