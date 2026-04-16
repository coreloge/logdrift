// Package group partitions a stream of log entries into named buckets keyed
// by an arbitrary field value (e.g. service, level, or any extra field).
//
// Basic usage:
//
//	g, err := group.New(group.DefaultOptions())
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Record entries directly …
//	g.Record(entry)
//
//	// … or wire into a pipeline.
//	out := group.Stream(ctx, g, in)
//
// After processing, retrieve all entries for a specific key:
//
//	entries := g.Get("auth-service")
package group
