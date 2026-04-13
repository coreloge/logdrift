// Package correlate provides a time-bounded index that groups structured log
// entries from multiple services by a shared correlation field (e.g.
// "correlation_id", "request_id").
//
// # Usage
//
//	corr := correlate.New(correlate.DefaultOptions())
//
//	// Feed entries from the multiplexer stream.
//	for e := range stream {
//		corr.Record(e)
//	}
//
//	// Later, retrieve all entries that share a correlation ID.
//	entries := corr.Get("abc-123")
//
// Groups are automatically evicted once their last-updated timestamp exceeds
// the configured Window, keeping memory usage bounded during long-running
// sessions.
package correlate
