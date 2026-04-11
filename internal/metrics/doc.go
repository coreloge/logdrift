// Package metrics provides lightweight in-process counters for logdrift.
//
// It tracks two categories of events per service:
//
//   - Entry counts: the total number of log lines received from a service.
//   - Drift counts: the number of times a service produced a log entry that
//     differed from the reference service in level, message, or fields.
//
// All operations on Counter are safe for concurrent use.
//
// Typical usage:
//
//	ctr := metrics.New()
//	ctr.RecordEntry("auth-service")
//	ctr.RecordDrift("auth-service")
//	fmt.Println(ctr.Entries())
//	fmt.Println(ctr.Drifts())
package metrics
