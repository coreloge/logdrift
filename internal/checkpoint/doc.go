// Package checkpoint provides durable offset tracking for log tail sources.
//
// A [Store] maps service names to byte offsets and persists the mapping as
// JSON on disk.  On startup logdrift can query the store to seek each file
// tail to its last known position, avoiding re-processing previously seen
// log lines after a restart or crash.
//
// # Usage
//
//	store, err := checkpoint.New("/var/lib/logdrift/checkpoints.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Wrap a live entry stream so offsets are updated automatically.
//	tracked := checkpoint.Stream(ctx, upstream, store)
//
// Offsets are approximate: they are derived from the byte length of the
// formatted entry rather than the underlying file descriptor position.
// This is sufficient for crash-recovery heuristics but should not be
// treated as an exact seek position.
package checkpoint
