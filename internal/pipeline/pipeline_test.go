package pipeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/metrics"
	"github.com/yourorg/logdrift/internal/pipeline"
	"github.com/yourorg/logdrift/internal/render"
	"github.com/yourorg/logdrift/internal/snapshot"
	"github.com/yourorg/logdrift/internal/stream"
)

// stubMultiplexer satisfies the stream.Multiplexer interface for testing.
type stubMultiplexer struct {
	ch chan diff.Entry
}

// newStub creates a stubMultiplexer pre-loaded with the given entries.
// The channel is closed immediately so the pipeline drains and exits naturally.
func newStub(entries ...diff.Entry) *stubMultiplexer {
	ch := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return &stubMultiplexer{ch: ch}
}

func (s *stubMultiplexer) Out() <-chan diff.Entry { return s.ch }
func (s *stubMultiplexer) Stop()                 {}

// newTestPipeline is a helper that wires up a Pipeline with the given
// multiplexer, using a fresh counter, snapshot, collector, and a
// discard renderer. It reduces boilerplate across test cases.
func newTestPipeline(mx *stubMultiplexer) (*pipeline.Pipeline, *metrics.Counter) {
	ctr := metrics.New()
	snap := snapshot.New()
	coll := snapshot.NewCollector(snap)
	rdr := render.New(nil)
	p := pipeline.New(pipeline.Config{
		Multiplexer: mx,
		Collector:   coll,
		Renderer:    rdr,
		Counter:     ctr,
		FilterOpts:  stream.FilterOptions{},
	})
	return p, ctr
}

func TestPipeline_RunUntilContextCancelled(t *testing.T) {
	mx := newStub() // empty — pipeline should exit via ctx cancel
	p, _ := newTestPipeline(mx)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := p.Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipeline_RecordsEntries(t *testing.T) {
	entries := []diff.Entry{
		{Service: "api", Level: "info", Message: "started"},
		{Service: "api", Level: "info", Message: "ready"},
	}
	mx := newStub(entries...)
	p, ctr := newTestPipeline(mx)

	ctx := context.Background()
	_ = p.Run(ctx)

	if got := ctr.Entries()["api"]; got != 2 {
		t.Fatalf("expected 2 entries for api, got %d", got)
	}
}
