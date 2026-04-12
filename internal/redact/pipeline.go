package redact

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream reads LogEntry values from in, applies redaction to each entry's
// Fields map, and forwards the result to the returned channel.
// The returned channel is closed when ctx is cancelled or in is closed.
func Stream(ctx context.Context, r *Redactor, in <-chan diff.LogEntry) <-chan diff.LogEntry {
	out := make(chan diff.LogEntry, cap(in))
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
				entry.Fields = r.Apply(entry.Fields)
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
