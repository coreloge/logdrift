package lookup

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream reads entries from in, applies the lookup enrichment, and forwards
// every entry (resolved or unchanged) to the returned channel.
// The output channel is closed when ctx is cancelled or in is closed.
func Stream(ctx context.Context, l *Lookup, in <-chan diff.Entry) <-chan diff.Entry {
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
				case out <- l.Apply(e):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
