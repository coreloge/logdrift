package mask

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads entries from in, applies masking, and forwards results to the
// returned channel. The output channel is closed when ctx is cancelled or in
// is closed.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) <-chan diff.Entry {
	out := make(chan diff.Entry)
	m := New(opts)

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
				case out <- m.Apply(entry):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}
