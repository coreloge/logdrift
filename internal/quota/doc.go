// Package quota provides a rolling-window per-service entry quota for log
// streams. It is intended to be composed into a processing pipeline to shed
// load from noisy services without dropping traffic from well-behaved ones.
//
// Usage:
//
//	q, err := quota.New(quota.Options{Max: 100, Window: time.Minute})
//	if err != nil { ... }
//	filtered := quota.Stream(ctx, q, entries)
//
// Entries arriving after the per-service limit has been reached within the
// current window are silently dropped. The window resets automatically once
// the configured duration has elapsed.
package quota
