package counter

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads entries from in, records each one through c, and forwards the
// annotated copy to the returned channel. The output channel is closed when in
// is closed or ctx is cancelled.
func Stream(ctx context.Context, c *Counter, in <-chan diff.Entry) <-chan diff.Entry {
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
				annotated := c.Record(e)
				select {
				case out <- annotated:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
