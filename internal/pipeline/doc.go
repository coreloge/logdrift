// Package pipeline provides the top-level orchestration layer for logdrift.
//
// A Pipeline connects the stream multiplexer, entry filter, snapshot
// collector, diff engine, renderer, metrics counter and alert watcher
// into a single cohesive processing loop.
//
// Typical usage:
//
//	p := pipeline.New(pipeline.Config{
//		Multiplexer:  mx,
//		Collector:    coll,
//		Renderer:     rdr,
//		Counter:      ctr,
//		AlertWatcher: aw,
//		FilterOpts:   opts,
//	})
//	if err := p.Run(ctx); err != nil {
//		log.Fatal(err)
//	}
package pipeline
