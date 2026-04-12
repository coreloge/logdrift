package enrich

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream reads entries from in, applies enrichment, and forwards them to
// the returned channel. The output channel is closed when ctx is cancelled
// or in is closed.
func Stream(ctx context.Context, e *Enricher, in <-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry)
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
				select {
				case out <- e.Apply(entry):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
