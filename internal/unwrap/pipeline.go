package unwrap

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads entries from in, applies the Unwrapper, and forwards results.
// It closes the returned channel when in is drained or ctx is cancelled.
func Stream(ctx context.Context, u *Unwrapper, in <-chan diff.Entry) <-chan diff.Entry {
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
				case out <- u.Apply(e):
				case <-ctx.Done():
					return
				}
			}
		}
	
}
