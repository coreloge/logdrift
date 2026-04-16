// Package distinct provides a pipeline stage that filters a stream of log
// entries, forwarding only those whose configured field value has not been
// seen before.
//
// This is useful for suppressing repetitive log lines while still capturing
// the first occurrence of each unique message, level, or custom field value.
//
// Example usage:
//
//	d, err := distinct.New(distinct.Options{Field: "message"})
//	if err != nil { ... }
//	out := distinct.Stream(ctx, in, d)
package distinct
