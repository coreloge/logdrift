// Package replay provides a Replayer that reads structured log entries from a
// static file and emits them into a diff.Entry channel at a configurable rate.
//
// This is useful for:
//   - Integration testing pipelines without live services.
//   - Demonstrating logdrift features against captured log snapshots.
//   - Benchmarking downstream consumers (filters, renderers, alerters).
//
// Basic usage:
//
//	opts := replay.Options{DelayPerLine: 10 * time.Millisecond, Service: "api"}
//	r := replay.New("captured.log", opts)
//	ch, err := r.Run(ctx)
//	if err != nil { ... }
//	for entry := range ch {
//	    // process entry
//	}
//
// Replay speed:
//
// Setting DelayPerLine to 0 causes the Replayer to emit all entries as fast as
// possible, which is useful for benchmarks. For realistic simulations, a value
// of 50–200ms approximates a moderately busy service.
//
// Context cancellation:
//
// Cancelling the context passed to Run causes the Replayer to stop reading and
// close the output channel cleanly, without dropping any already-queued entries.
package replay
