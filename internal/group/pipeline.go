package group

import (
	"context"

	"github.com/logdrift/logdrift/internal/diff"
)

// Stream reads entries from in, records each one in g, and forwards every
// entry unchanged to the returned channel. The output channel is closed when
// ctx is cancelled or in is closed.
func Stream(ctx context.Context, g *Grouper, in <-chan diff.Entry) <-chan diff.Entry {
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
				g.Record(entry)
				select {
				case out <- entry:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
