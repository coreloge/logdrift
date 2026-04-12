// Package ratelimit implements per-service token-bucket rate limiting for
// logdrift entry streams.
//
// # Overview
//
// When tailing many services simultaneously it is easy for a noisy service to
// flood the output or skew snapshot comparisons. The Throttle type limits how
// many entries are forwarded per service per second, dropping the remainder.
//
// # Usage
//
//	th := ratelimit.New(ratelimit.Options{MaxPerSecond: 100})
//	filtered := th.Apply(entryChan)
//
// Setting MaxPerSecond to 0 (the default returned by DefaultOptions) disables
// rate limiting entirely so existing pipelines are unaffected unless an
// explicit limit is configured.
package ratelimit
