package expire

import (
	"context"

	"github.com/robmorgan/logdrift/internal/diff"
)

// Stream reads entries from in, drops any that have expired according to e,
// and forwards the rest to the returned channel.
// The output channel is closed when ctx is cancelled or in is closed.
func Stream(ctx context.Context, e *Expirer, in <-chan diff.Entry) <-chan diff.Entry {
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
				if e.Allow(entry) {
					select {
					case out <- entry:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
