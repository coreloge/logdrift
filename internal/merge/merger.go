// Package merge provides ordered merging of multiple log entry streams
// into a single chronologically sorted output channel.
package merge

import (
	"container/heap"
	"context"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the Merger.
type Options struct {
	// BufferSize is the capacity of the output channel.
	BufferSize int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{BufferSize: 64}
}

// Merger merges multiple entry streams ordered by timestamp.
type Merger struct {
	opts Options
}

// New creates a Merger with the given options.
func New(opts Options) *Merger {
	if opts.BufferSize <= 0 {
		opts.BufferSize = DefaultOptions().BufferSize
	}
	return &Merger{opts: opts}
}

// item wraps an entry with its source channel index for heap ordering.
type item struct {
	entry diff.Entry
	src   int
}

// entryHeap implements heap.Interface ordered by entry timestamp.
type entryHeap []item

func (h entryHeap) Len() int            { return len(h) }
func (h entryHeap) Less(i, j int) bool  { return h[i].entry.Timestamp.Before(h[j].entry.Timestamp) }
func (h entryHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *entryHeap) Push(x interface{}) { *h = append(*h, x.(item)) }
func (h *entryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// Stream merges sources into a single chronologically ordered channel.
// It returns when all sources are drained or ctx is cancelled.
func (m *Merger) Stream(ctx context.Context, sources []<-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry, m.opts.BufferSize)
	go func() {
		defer close(out)
		h := &entryHeap{}
		heap.Init(h)

		// Seed heap with first entry from each source.
		active := make([]<-chan diff.Entry, len(sources))
		copy(active, sources)
		for i, ch := range active {
			if e, ok := <-ch; ok {
				heap.Push(h, item{entry: e, src: i})
			} else {
				active[i] = nil
			}
		}

		for h.Len() > 0 {
			select {
			case <-ctx.Done():
				return
			default:
			}
			it := heap.Pop(h).(item)
			select {
			case <-ctx.Done():
				return
			case out <- it.entry:
			}
			if ch := active[it.src]; ch != nil {
				if e, ok := <-ch; ok {
					heap.Push(h, item{entry: e, src: it.src})
				} else {
					active[it.src] = nil
				}
			}
		}
	}()
	return out
}

// ensure diff.Entry has a Timestamp field — used only for compilation reference.
var _ = time.Time{}
