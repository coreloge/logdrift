package parse

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream reads entries from in, applies p to each one, and forwards the
// enriched entry to the returned channel. The output channel is closed when
// in is closed or ctx is cancelled.
func Stream(ctx context.Context, p *Parser, in <-chan diff.Entry) <-chan diff.Entry {
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
				select {
				case out <- p.Apply(e):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
