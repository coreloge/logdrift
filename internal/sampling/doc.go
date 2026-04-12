// Package sampling provides pluggable strategies for reducing the volume of
// log entries flowing through a logdrift pipeline.
//
// Three strategies are available:
//
//   - StrategyNone    – every entry is forwarded unchanged (default).
//   - StrategyRandom  – each entry is kept with probability Rate (0.0–1.0).
//   - StrategyRateLimit – at most Rate entries per second are forwarded for
//     each distinct service label; excess entries within the same second are
//     silently dropped.
//
// Typical usage:
//
//	sampler := sampling.New(sampling.Options{
//		Strategy: sampling.StrategyRateLimit,
//		Rate:     100, // max 100 entries/s per service
//	})
//	defer sampler.Stop()
//
//	for entry := range stream {
//		if sampler.Accept(entry) {
//			downstream <- entry
//		}
//	}
package sampling
