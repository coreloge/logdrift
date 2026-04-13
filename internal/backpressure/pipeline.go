package backpressure

import (
	"context"

	"github.com/logdrift/logdrift/internal/diff"
)

// Stream is a convenience wrapper that constructs a Backpressure stage with
// default options and immediately starts streaming from in.
// It returns the output channel and the stage so callers can inspect Dropped().
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) (<-chan diff.Entry, *Backpressure) {
	bp := New(opts)
	out := bp.Stream(ctx, in)
	return out, bp
}
