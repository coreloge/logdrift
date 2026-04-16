// Package proxy provides a pass-through pipeline stage that observes
// entries without modifying them, useful for side-effects such as logging
// counts or triggering hooks mid-pipeline.
package proxy

import (
	"context"
	"sync"

	"github.com/logdrift/logdrift/internal/diff"
)

// HookFunc is called for every entry that passes through the proxy.
type HookFunc func(entry diff.Entry)

// Options configures the Proxy.
type Options struct {
	// Hook is called synchronously for each entry. Must be non-nil.
	Hook HookFunc
}

// DefaultOptions returns an Options with a no-op hook.
func DefaultOptions() Options {
	return Options{
		Hook: func(diff.Entry) {},
	}
}

// Proxy observes entries and forwards them unchanged.
type Proxy struct {
	opts Options
	mu   sync.Mutex
	seen int
}

// New creates a new Proxy. Returns an error if Hook is nil.
func New(opts Options) (*Proxy, error) {
	if opts.Hook == nil {
		return nil, errNilHook
	}
	return &Proxy{opts: opts}, nil
}

// Apply calls the hook and returns the entry unchanged.
func (p *Proxy) Apply(e diff.Entry) diff.Entry {
	p.mu.Lock()
	p.seen++
	p.mu.Unlock()
	p.opts.Hook(e)
	return e
}

// Seen returns the total number of entries observed.
func (p *Proxy) Seen() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.seen
}

// Stream reads from in, applies the proxy hook, and forwards each entry to
// the returned channel. The output channel is closed when ctx is done or in
// is closed.
func Stream(ctx context.Context, p *Proxy, in <-chan diff.Entry) <-chan diff.Entry {
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
				select {
				case out <- p.Apply(e):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
