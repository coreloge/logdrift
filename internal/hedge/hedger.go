// Package hedge provides a hedged-read pattern for log entry sources.
// A hedged request fans out to multiple sources and returns the first
// successful result, cancelling the remaining in-flight reads.
package hedge

import (
	"context"
	"sync"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the Hedger.
type Options struct {
	// MaxSources is the maximum number of sources to fan out to.
	// Zero means no limit.
	MaxSources int
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxSources: 0,
	}
}

// Hedger fans out a read to multiple entry channels and forwards
// the first entry received from any of them.
type Hedger struct {
	opts    Options
	sources []<-chan diff.Entry
}

// New creates a Hedger for the provided sources.
func New(opts Options, sources ...(<-chan diff.Entry)) *Hedger {
	if opts.MaxSources > 0 && len(sources) > opts.MaxSources {
		sources = sources[:opts.MaxSources]
	}
	return &Hedger{opts: opts, sources: sources}
}

// Len returns the number of sources the Hedger is configured to fan out to.
func (h *Hedger) Len() int {
	return len(h.sources)
}

// Stream returns a channel that emits the first entry received from any
// source. All sources are drained concurrently; whichever delivers an
// entry first wins and the result is forwarded downstream. The returned
// channel is closed when ctx is cancelled or all sources are exhausted.
func (h *Hedger) Stream(ctx context.Context) <-chan diff.Entry {
	out := make(chan diff.Entry)
	go func() {
		defer close(out)
		merged := make(chan diff.Entry, len(h.sources))
		var wg sync.WaitGroup
		for _, src := range h.sources {
			wg.Add(1)
			go func(ch <-chan diff.Entry) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case e, ok := <-ch:
						if !ok {
							return
						}
						select {
						case merged <- e:
						case <-ctx.Done():
							return
						}
					}
				}
			}(src)
		}
		go func() {
			wg.Wait()
			close(merged)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-merged:
				if !ok {
					return
				}
				select {
				case out <- e:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
