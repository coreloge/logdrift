// Package throttle implements per-level token-bucket rate limiting for
// structured log entry streams.
//
// Each log level ("debug", "info", "warn", "error", etc.) can be assigned an
// independent maximum throughput expressed in entries per second. Entries
// arriving faster than the configured rate are silently dropped.
//
// Usage:
//
//	opts := throttle.DefaultOptions()
//	opts.LevelRates["debug"] = 10  // allow 10 debug entries/s
//	opts.LevelRates["error"] = 100 // allow 100 error entries/s
//
//	limiter := throttle.New(opts)
//	out := throttle.Stream(ctx, in, limiter)
package throttle
