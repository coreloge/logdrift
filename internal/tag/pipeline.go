package tag

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads entries from in, applies the tagger to each entry, and forwards
// the result to the returned channel. The output channel is closed when ctx is
// cancelled or in is closed.
func Stream(ctx context.Context, t *Tagger, in <-chan diff.LogEntry) <-chan diff.LogEntry {
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
				tagged := t.Apply(entry)
				select {
				case out <- tagged:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
