package elect

import (
	"context"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// GuardStream forwards entries from in only while candidate holds the
// leader lease, dropping entries when another candidate is elected.
// The lease is renewed on every received entry.
func GuardStream(
	ctx context.Context,
	in <-chan diff.Entry,
	e *Elector,
	candidate string,
) <-chan diff.Entry {
	out := make(chan diff.Entry, 64)
	go func() {
		defer close(out)
		renewTick := time.NewTicker(e.opts.RenewEvery)
		defer renewTick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-renewTick.C:
				e.Acquire(candidate)
			case entry, ok := <-in:
				if !ok {
					return
				}
				if e.Acquire(candidate) {
					select {
					case out <- entry:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
