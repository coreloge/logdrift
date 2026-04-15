package sequence

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads entries from in, stamps each with a sequence number via s, and
// forwards them to the returned channel. The output channel is closed when ctx
// is cancelled or in is closed.
func Stream(ctx context.Context, s *Sequencer, in <-chan diff.Entry) <-chan diff.Entry {
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
				stamped := s.Apply(e)
				select {
				case out <- stamped:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
