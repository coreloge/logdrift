package dedupe

import (
	"context"

	"github.com/logdrift/internal/diff"
)

// Stream wraps an input channel and returns a new channel that omits duplicate
// entries as determined by the provided Deduper. The output channel is closed
// when ctx is cancelled or the input channel is closed.
func Stream(ctx context.Context, in <-chan diff.LogEntry, d *Deduper) <-chan diff.LogEntry {
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
				if d.IsDuplicate(entry) {
					continue
				}
				select {
				case out <- entry:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
