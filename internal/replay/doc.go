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
package replay
