// Package merge implements chronological merging of multiple structured log
// entry streams into a single ordered output stream.
//
// It uses a min-heap keyed on entry timestamps so that entries from N
// concurrent sources are emitted in wall-clock order without buffering the
// entire dataset in memory.
//
// Basic usage:
//
//	// Variadic convenience wrapper.
//	out := merge.Stream(ctx, chA, chB, chC)
//
//	// Or with custom buffer size.
//	m := merge.New(merge.Options{BufferSize: 128})
//	out := m.Stream(ctx, []<-chan diff.Entry{chA, chB})
//
// The output channel is closed once all source channels are drained or the
// context is cancelled, whichever comes first.
package merge
