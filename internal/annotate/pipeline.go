package annotate

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads entries from in, applies the annotator to each, and forwards
// the annotated entry to the returned channel. The output channel is closed
// when ctx is cancelled or in is closed.
func Stream(ctx context.Context, a *Annotator, in <-chan diff.Entry) <-chan diff.Entry {
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
				case out <- a.Apply(entry):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
