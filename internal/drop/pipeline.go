package drop

import (
	"context"

	"github.com/logdrift/logdrift/internal/diff"
)

// Stream filters out entries that match any drop rule, forwarding the rest.
func Stream(ctx context.Context, in <-chan diff.Entry, d *Dropper) <-chan diff.Entry {
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
				if !d.ShouldDrop(e) {
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
