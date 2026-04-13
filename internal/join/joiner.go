// Package join provides a stream joiner that merges multiple log entry
// channels into a single ordered output channel, tagging each entry with
// its originating source label.
package join

import (
	"context"
	"sync"

	"github.com/logdrift/logdrift/internal/diff"
)

// Source pairs a named label with its corresponding entry channel.
type Source struct {
	Label string
	Ch    <-chan diff.Entry
}

// Options controls Joiner behaviour.
type Options struct {
	// BufferSize is the capacity of the merged output channel.
	BufferSize int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{BufferSize: 64}
}

// Joiner fans-in multiple Source channels into one.
type Joiner struct {
	opts    Options
	sources []Source
}

// New creates a Joiner for the given sources.
func New(sources []Source, opts Options) *Joiner {
	if opts.BufferSize <= 0 {
		opts.BufferSize = DefaultOptions().BufferSize
	}
	return &Joiner{opts: opts, sources: sources}
}

// Stream starts one goroutine per source and forwards every entry to the
// returned channel. The channel is closed once all sources are exhausted or
// ctx is cancelled.
func (j *Joiner) Stream(ctx context.Context) <-chan diff.Entry {
	out := make(chan diff.Entry, j.opts.BufferSize)

	var wg sync.WaitGroup
	for _, src := range j.sources {
		wg.Add(1)
		go func(s Source) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case e, ok := <-s.Ch:
					if !ok {
						return
					}
					if e.Service == "" {
						e.Service = s.Label
					}
					select {
					case out <- e:
					case <-ctx.Done():
						return
					}
				}
			}
		}(src)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
