package tee

import (
	"context"
	"io"

	"github.com/user/logdrift/internal/diff"
)

// Pipe is a convenience wrapper that constructs a Tee from an io.Writer and
// immediately starts streaming. It panics if New returns an error (it never
// does for a non-nil writer).
func Pipe(ctx context.Context, in <-chan diff.Entry, w io.Writer, format string) <-chan diff.Entry {
	t, err := New(Options{Writer: w, Format: format})
	if err != nil {
		panic("tee.Pipe: " + err.Error())
	}
	return Stream(ctx, in, t)
}
