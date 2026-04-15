// Package skip implements a pipeline stage that discards the first N entries
// from a structured log stream.
//
// # Overview
//
// When tailing log files that contain a fixed preamble — such as startup
// banners, version headers, or initialisation noise — it is useful to skip
// those entries before any downstream processing (diffing, alerting, etc.).
//
// # Usage
//
//	opts := skip.Options{
//	    Count:      5,      // drop the first 5 entries globally
//	    PerService: false,  // share one counter across all services
//	}
//	out, err := skip.Stream(ctx, in, opts)
//
// When PerService is true each service label maintains its own independent
// counter, so every service individually skips its first Count entries.
//
// # Zero-skip
//
// Setting Count to 0 (the default) is a no-op; all entries are forwarded
// without overhead.
package skip
