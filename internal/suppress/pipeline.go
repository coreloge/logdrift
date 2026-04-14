package suppress

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream reads entries from in, forwards those that pass the Suppressor, and
// closes the returned channel when ctx is done or in is closed.
func Stream(ctx context.Context, s *Suppressor, in <-chan diff.LogEntry) <-chan diff.LogEntry {
	out := make(chan diff.LogEntry)
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
				if s.Allow(e) {
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
