// Package suppress implements a cooldown-based suppression stage for the
// logdrift pipeline.
//
// Entries whose designated field matches a configured regular expression are
// forwarded at most once per cooldown window per (service, value) pair.
// Subsequent matching entries within the window are silently dropped, reducing
// noise from high-frequency repeated log lines without losing the first
// occurrence.
//
// Usage:
//
//	supp, err := suppress.New(suppress.Options{
//		Field:    "message",
//		Pattern:  `connection refused`,
//		Cooldown: 10 * time.Second,
//	})
//	if err != nil { ... }
//	out := suppress.Stream(ctx, supp, in)
package suppress
