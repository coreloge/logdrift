// Package counter implements a streaming per-field occurrence counter for
// logdrift log entries.
//
// A Counter tracks how many times each distinct value of a chosen field
// (e.g. "level" or "service") has been seen. Each call to Record returns a
// copy of the entry annotated with the running count in a configurable Extra
// field (default: "_count").
//
// Use Stream to wire the counter into a channel-based pipeline:
//
//	c, _ := counter.New(counter.DefaultOptions())
//	annotated := counter.Stream(ctx, c, upstream)
//
// Counts are safe for concurrent use.
package counter
