// Package backpressure implements a pipeline stage that manages downstream
// congestion in logdrift entry streams.
//
// When a downstream consumer processes entries more slowly than they arrive,
// the backpressure stage can either drop excess entries immediately (Drop
// strategy) or wait up to a configurable timeout before discarding them
// (Block strategy).
//
// Usage:
//
//	opts := backpressure.DefaultOptions()
//	opts.Strategy = backpressure.Block
//	opts.Timeout = 50 * time.Millisecond
//
//	out, bp := backpressure.Stream(ctx, in, opts)
//	for entry := range out {
//		// process entry
//	}
//	fmt.Println("dropped:", bp.Dropped())
package backpressure
