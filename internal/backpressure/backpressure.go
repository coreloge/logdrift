// Package backpressure provides a pipeline stage that applies backpressure
// to an entry stream when downstream consumers are slow, dropping or blocking
// entries according to the configured strategy.
package backpressure

import (
	"context"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
)

// Strategy controls how the stage behaves when the output channel is full.
type Strategy string

const (
	// Drop silently discards entries that cannot be forwarded immediately.
	Drop Strategy = "drop"
	// Block waits up to Timeout for space in the output channel.
	Block Strategy = "block"
)

// Options configures the Backpressure stage.
type Options struct {
	// Strategy determines what happens when the output buffer is full.
	Strategy Strategy
	// BufferSize is the capacity of the internal output channel.
	BufferSize int
	// Timeout is the maximum time to wait when Strategy is Block.
	Timeout time.Duration
}

// DefaultOptions returns sensible defaults: drop strategy, buffer of 64, no timeout.
func DefaultOptions() Options {
	return Options{
		Strategy:   Drop,
		BufferSize: 64,
		Timeout:    100 * time.Millisecond,
	}
}

// Backpressure holds runtime state for the stage.
type Backpressure struct {
	opts    Options
	dropped uint64
}

// New creates a new Backpressure stage with the given options.
func New(opts Options) *Backpressure {
	if opts.BufferSize <= 0 {
		opts.BufferSize = DefaultOptions().BufferSize
	}
	return &Backpressure{opts: opts}
}

// Dropped returns the number of entries dropped since the stage was created.
func (b *Backpressure) Dropped() uint64 { return b.dropped }

// Stream reads from in and forwards entries to the returned channel, applying
// the configured backpressure strategy when the buffer is full.
func (b *Backpressure) Stream(ctx context.Context, in <-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry, b.opts.BufferSize)
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
				b.forward(ctx, out, entry)
			}
		}
	}()
	return out
}

func (b *Backpressure) forward(ctx context.Context, out chan<- diff.Entry, e diff.Entry) {
	if b.opts.Strategy == Block {
		timer := time.NewTimer(b.opts.Timeout)
		defer timer.Stop()
		select {
		case out <- e:
		case <-timer.C:
			b.dropped++
		case <-ctx.Done():
		}
		return
	}
	// Drop strategy: non-blocking send.
	select {
	case out <- e:
	default:
		b.dropped++
	}
}
