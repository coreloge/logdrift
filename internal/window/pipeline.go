package window

import (
	"context"

	"logdrift/internal/diff"
)

// Stream reads entries from in, feeds them into w, and forwards every entry
// unchanged to the returned channel. The returned channel is closed when ctx
// is cancelled or in is closed.
func Stream(ctx context.Context, w *Window, in <-chan diff.Entry) <-chan diff.Entry {
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
				w.Add(e)
				select {
				case out <- e:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
