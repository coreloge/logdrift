package quota

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream reads entries from in, drops those that exceed the per-service quota,
// and forwards the rest to the returned channel. It closes the output channel
// when ctx is cancelled or in is closed.
func Stream(ctx context.Context, q *Quota, in <-chan diff.Entry) <-chan diff.Entry {
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
				if q.Allow(e.Service) {
					select {
					case out <- e:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
