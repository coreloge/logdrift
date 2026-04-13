package route

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// Stream wraps New and Run into a convenient helper that starts the router in
// a background goroutine and returns the router so callers can subscribe to
// named routes before entries arrive.
//
// The router stops when ctx is cancelled or in is closed.
func Stream(ctx context.Context, in <-chan diff.Entry, opts Options) *Router {
	r := New(opts)
	go r.Run(ctx, in)
	return r
}
