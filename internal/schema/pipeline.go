package schema

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream reads entries from in, validates each one, and forwards valid entries
// to the returned channel. When DropInvalid is true, invalid entries are
// silently discarded; otherwise they are forwarded unchanged (validation errors
// are non-fatal in stream mode — use Validator.Validate directly for strict
// rejection). The returned channel is closed when ctx is done or in is closed.
func Stream(ctx context.Context, v *Validator, in <-chan diff.Entry) <-chan diff.Entry {
	out := make(chan diff.Entry)
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
				if err := v.Validate(entry); err != nil {
					if v.opts.DropInvalid {
						continue
					}
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
