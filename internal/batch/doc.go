// Package batch implements time- and size-based batching for log entry streams.
//
// A Batcher reads individual diff.LogEntry values from an input channel and
// groups them into slices that are emitted on an output channel. A batch is
// flushed when either:
//
//   - the number of buffered entries reaches Options.MaxSize, or
//   - Options.FlushInterval elapses since the last flush.
//
// This is useful for downstream consumers that benefit from processing entries
// in bulk — for example, writing to a database, computing aggregate statistics,
// or forwarding over a network connection with reduced overhead.
//
// Basic usage:
//
//	b := batch.New(batch.DefaultOptions())
//	batches := b.Stream(ctx, entryCh)
//	for batch := range batches {
//		// process []diff.LogEntry
//	}
package batch
