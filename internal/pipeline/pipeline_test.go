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

func TestPipeline_RunUntilContextCancelled(t *testing.T) {
	mx := newStub() // empty — pipeline should exit via ctx cancel
	ctr := metrics.New()
	snap := snapshot.New()
	coll := snapshot.NewCollector(snap)
	rdr := render.New(nil) // discard output

	p := pipeline.New(pipeline.Config{
		Multiplexer: mx,
		Collector:   coll,
		Renderer:    rdr,
		Counter:     ctr,
		FilterOpts:  stream.FilterOptions{},
	})

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

	ctx := context.Background()
	_ = p.Run(ctx)

	if got := ctr.Entries()["api"]; got != 2 {
		t.Fatalf("expected 2 entries for api, got %d", got)
	}
}
