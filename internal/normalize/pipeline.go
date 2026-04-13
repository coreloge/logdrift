package normalize

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads entries from in, applies the Normalizer, and forwards the
// normalised entries to the returned channel. The output channel is closed
// when in is closed or ctx is cancelled.
func Stream(ctx context.Context, n *Normalizer, in <-chan diff.Entry) <-chan diff.Entry {
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
				normalised := n.Apply(entry)
				select {
				case out <- normalised:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
