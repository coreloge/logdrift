package cursor

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream reads log entries from in, updates the cursor Store for each entry's
// service, and forwards every entry unchanged to the returned channel.
// The returned channel is closed when ctx is cancelled or in is closed.
func Stream(ctx context.Context, store *Store, in <-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry, cap(in))
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
				// Each raw line is assumed to consume len(entry.Message)+1 bytes
				// (message + newline). Real byte offsets would come from the tailer.
				store.Advance(entry.Service, int64(len(entry.Message)+1), 1)
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
