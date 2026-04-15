package merge

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream is a convenience wrapper that constructs a Merger with default
// options and immediately merges the provided sources.
//
// Example:
//
//	out := merge.Stream(ctx, ch1, ch2, ch3)
func Stream(ctx context.Context, sources ...<-chan diff.Entry) <-chan diff.Entry {
	m := New(DefaultOptions())
	return m.Stream(ctx, sources)
}
