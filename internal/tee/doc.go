// Package tee implements a transparent pipeline tap for log entry streams.
//
// A Tee stage forwards every entry downstream unchanged while simultaneously
// writing a formatted copy to an io.Writer side-channel. This is useful for
// capturing a raw dump to a file or network sink without interrupting the
// main processing pipeline.
//
// Example usage:
//
//	t, _ := tee.New(tee.Options{
//		Writer: os.Stderr,
//		Format: "text",
//	})
//	out := tee.Stream(ctx, in, t)
package tee
