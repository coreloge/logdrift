// Package reorder implements a timestamp-based reordering buffer for log entry
// streams.
//
// Log entries arriving out of chronological order — common when tailing
// multiple services with clock skew or buffered writes — are held for a
// configurable HoldWindow before being flushed downstream in ascending
// timestamp order.
//
// Usage:
//
//	r, err := reorder.New(reorder.DefaultOptions())
//	if err != nil {
//		log.Fatal(err)
//	}
//	ordered := r.Stream(ctx, rawEntries)
//
// A forced flush occurs when the internal buffer reaches MaxBuffer entries,
// preventing unbounded memory growth under high-throughput conditions.
package reorder
