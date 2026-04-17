package shard

import (
	"context"

	"github.com/logdrift/logdrift/internal/diff"
)

// Stream reads from in, assigns each entry to a shard, and closes the sharder
// when the context is cancelled or the input channel is drained.
func Stream(ctx context.Context, in <-chan diff.Entry, s *Sharder) {
	go func() {
		defer s.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-in:
				if !ok {
					return
				}
				s.Assign(e)
			}
		}
	}()
}
