// Package window implements a sliding time-window aggregator for structured
// log entries.
//
// # Overview
//
// A [Window] partitions incoming [diff.Entry] values into fixed-duration
// buckets (e.g. 30 s). Buckets are created on demand and the oldest ones
// are evicted once [Options.MaxBuckets] is exceeded, keeping memory usage
// bounded regardless of stream volume.
//
// # Usage
//
//	w := window.New(window.DefaultOptions())
//	out := window.Stream(ctx, w, entryChan)
//	// consume out …
//	buckets := w.Buckets()  // inspect aggregated data at any time
//
// # Thread Safety
//
// Window is safe for concurrent use.
package window
