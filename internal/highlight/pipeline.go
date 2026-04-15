package highlight

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream reads entries from in, applies highlighting, and forwards each
// coloured entry to the returned channel. The output channel is closed when
// ctx is cancelled or in is closed.
func Stream(ctx context.Context, h *Highlighter, in <-chan diff.LogEntry) <-chan diff.LogEntry {
	out := make(chan diff.LogEntry)
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
				case out <- h.Apply(entry):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
