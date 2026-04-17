// Package shard provides a content-based sharding stage for log entry streams.
//
// A Sharder hashes a configurable field (e.g. "service") and routes each entry
// to one of N output channels. This allows downstream consumers to process
// entries in parallel without reordering concerns within a single shard.
//
// Basic usage:
//
//	s, _ := shard.New(shard.Options{Field: "service", Shards: 4})
//	shard.Stream(ctx, input, s)
//	for _, out := range s.Outputs() {
//	    go consume(out)
//	}
package shard
