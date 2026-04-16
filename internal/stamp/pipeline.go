package stamp

import (
	"context"

	"github.com/logdrift/logdrift/internal/diff"
)

// Stream reads entries from in, stamps each one, and forwards them to the
// returned channel. The output channel is closed when ctx is cancelled or in
// is closed.
func Stream(ctx context.Context, s *Stamper, in <-chan diff.Entry) <-chan diff.Entry {
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
				case out <- s.Apply(e):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
