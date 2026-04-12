// Package pipeline wires together the core logdrift components into a
// single, runnable processing pipeline.
package pipeline

import (
	"context"

	"github.com/yourorg/logdrift/internal/alert"
	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/metrics"
	"github.com/yourorg/logdrift/internal/render"
	"github.com/yourorg/logdrift/internal/snapshot"
	"github.com/yourorg/logdrift/internal/stream"
)

// Config holds all dependencies required to construct a Pipeline.
type Config struct {
	Multiplexer  *stream.Multiplexer
	Collector    *snapshot.Collector
	Renderer     *render.Renderer
	Counter      *metrics.Counter
	AlertWatcher *alert.Watcher
	FilterOpts   stream.FilterOptions
}

// Pipeline orchestrates filtering, snapshotting, diffing and rendering.
type Pipeline struct {
	cfg Config
}

// New creates a Pipeline from the provided Config.
func New(cfg Config) *Pipeline {
	return &Pipeline{cfg: cfg}
}

// Run starts the pipeline and blocks until ctx is cancelled.
// Each log entry received from the multiplexer is:
//  1. Filtered according to FilterOpts.
//  2. Recorded in the metrics counter.
//  3. Accumulated by the collector.
//  4. Compared against the current snapshot and rendered.
func (p *Pipeline) Run(ctx context.Context) error {
	filtered := stream.Filter(p.cfg.Multiplexer.Out(), p.cfg.FilterOpts)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case entry, ok := <-filtered:
			if !ok {
				return nil
			}
			p.cfg.Counter.RecordEntry(entry.Service)

			p.cfg.Collector.Add(entry)
			snap := p.cfg.Collector.Snapshot()

			result := snap.Compare(diff.Compare)
			if len(result.Deltas) > 0 {
				p.cfg.Counter.RecordDrift(entry.Service)
			}

			p.cfg.Renderer.RenderEntry(entry)
			p.cfg.Renderer.RenderDiff(result)
		}
	}
}
