// Package coalesce provides a Coalescer that groups log entries from different
// services by a shared correlation field (e.g. request_id) and emits them as a
// unified batch once a minimum number of contributing services is reached or a
// configurable time window expires.
//
// Typical usage:
//
//	c := coalesce.New(coalesce.DefaultOptions())
//	if merged := c.Record(entry); merged != nil {
//		// process merged group
//	}
//	// periodically call c.Flush() to drain stale groups
package coalesce
