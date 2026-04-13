package throttle

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream filters entries from in, dropping those that exceed their per-level
// rate limit, and forwards the rest to the returned channel.
func Stream(ctx context.Context, in <-chan diff.Entry, l *Limiter) <-chan diff.Entry {
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
				if l.Allow(entry) {
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
