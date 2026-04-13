package audit

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/diff"
)

// Entry is the minimal interface consumed by the audit pipeline.
type Entry struct {
	Service string
	Level   string
	Message string
}

// DriftResult pairs an entry with any diff deltas detected.
type DriftResult struct {
	Entry  Entry
	Deltas []diff.Delta
}

// Stream reads DriftResult values from in, records them in the Auditor,
// and forwards each Entry to the returned channel.
// The returned channel is closed when ctx is cancelled or in is closed.
func Stream(ctx context.Context, a *Auditor, in <-chan DriftResult) <-chan Entry {
	out := make(chan Entry)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case dr, ok := <-in:
				if !ok {
					return
				}
				if len(dr.Deltas) > 0 {
					a.RecordDrift(dr.Entry.Service, dr.Deltas)
				} else {
					a.RecordEntry(dr.Entry.Service, dr.Entry.Level, dr.Entry.Message)
				}
				select {
				case out <- dr.Entry:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
