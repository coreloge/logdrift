// Package elect implements a lightweight in-process leader election
// mechanism for logdrift.
//
// When multiple goroutines (or service replicas represented as candidates)
// compete to process the same log stream, elect ensures only one holds
// the active lease at a time. Leases are time-bounded and must be
// periodically renewed; if a candidate fails to renew before the TTL
// expires, another candidate may acquire leadership.
//
// Basic usage:
//
//	e, _ := elect.New(elect.DefaultOptions())
//	if e.Acquire("primary") {
//		// process entries
//	}
//
// GuardStream wraps a diff.Entry channel, forwarding entries only while
// the given candidate holds the lease.
package elect
