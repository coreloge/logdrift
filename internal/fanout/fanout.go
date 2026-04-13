// Package fanout provides a mechanism for broadcasting log entries
// from a single input channel to multiple output channels.
package fanout

import (
	"context"
	"sync"

	"github.com/logdrift/internal/diff"
)

// Fanout reads entries from a single source channel and writes each entry
// to all registered subscriber channels.
type Fanout struct {
	mu   sync.RWMutex
	subs []chan diff.Entry
}

// New creates a Fanout with the given buffer size for each subscriber channel.
func New(bufSize int) *Fanout {
	return &Fanout{}
}

// Subscribe returns a new channel that will receive every entry broadcast
// through the Fanout. The channel is buffered with the provided bufSize.
func (f *Fanout) Subscribe(bufSize int) <-chan diff.Entry {
	ch := make(chan diff.Entry, bufSize)
	f.mu.Lock()
	f.subs = append(f.subs, ch)
	f.mu.Unlock()
	return ch
}

// Run reads from src until it is closed or ctx is cancelled, broadcasting
// each entry to all current subscribers. Subscriber channels are closed
// when Run returns.
func (f *Fanout) Run(ctx context.Context, src <-chan diff.Entry) {
	defer f.closeAll()
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-src:
			if !ok {
				return
			}
			f.broadcast(entry)
		}
	}
}

func (f *Fanout) broadcast(entry diff.Entry) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, ch := range f.subs {
		select {
		case ch <- entry:
		default:
			// drop if subscriber is full to avoid blocking the pipeline
		}
	}
}

func (f *Fanout) closeAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, ch := range f.subs {
		close(ch)
	}
	f.subs = nil
}
