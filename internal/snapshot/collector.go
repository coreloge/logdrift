package snapshot

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// StreamEntry pairs a parsed log entry with its originating service name.
type StreamEntry struct {
	Service string
	Entry   diff.Entry
}

// Collector reads from a channel of StreamEntry values and accumulates
// them into a Snapshot until the context is cancelled or the channel closes.
type Collector struct {
	snapshot *Snapshot
}

// NewCollector returns a Collector backed by a fresh Snapshot.
func NewCollector() *Collector {
	return &Collector{snapshot: New()}
}

// Collect reads entries from ch until it is closed or ctx is done.
// It is safe to call Snapshot() after Collect returns.
func (c *Collector) Collect(ctx context.Context, ch <-chan StreamEntry) {
	for {
		select {
		case <-ctx.Done():
			return
		case se, ok := <-ch:
			if !ok {
				return
			}
			c.snapshot.Add(se.Service, se.Entry)
		}
	}
}

// Snapshot returns the accumulated Snapshot.
func (c *Collector) Snapshot() *Snapshot {
	return c.snapshot
}
