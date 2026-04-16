package pin

import (
	"context"

	"github.com/logdrift/logdrift/internal/diff"
)

// Stream forwards every entry from in to the returned channel while
// also recording each entry in p. The returned channel is closed when
// ctx is cancelled or in is closed.
func Stream(ctx context.Context, p *Pinner, in <-chan diff.Entry) <-chan diff.Entry {
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
				p.Pin(e)
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
