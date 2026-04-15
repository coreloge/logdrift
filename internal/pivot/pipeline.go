package pivot

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads entries from in, records each one into p, and forwards the
// entry unchanged to the returned channel. The returned channel is closed when
// ctx is cancelled or in is closed.
func Stream(ctx context.Context, p *Pivoter, in <-chan diff.Entry) <-chan diff.Entry {
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
				p.Record(e)
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
