package score

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
)

func feedResults(results []diff.Result) <-chan diff.Result {
	ch := make(chan diff.Result, len(results))
	for _, r := range results {
		ch <- r
	}
	close(ch)
	return ch
}

func drainScored(ch <-chan ScoredResult) []ScoredResult {
	var out []ScoredResult
	for sr := range ch {
		out = append(out, sr)
	}
	return out
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	s, _ := New(DefaultOptions())
	in := feedResults([]diff.Result{makeResult(), makeResult("level")})
	out := Stream(context.Background(), s, in)
	got := drainScored(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
}

func TestStream_ScoresAreCorrect(t *testing.T) {
	opts := DefaultOptions()
	s, _ := New(opts)
	in := feedResults([]diff.Result{makeResult("level")})
	out := Stream(context.Background(), s, in)
	got := drainScored(out)
	if got[0].Score != opts.LevelWeight {
		t.Fatalf("expected score %f, got %f", opts.LevelWeight, got[0].Score)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	s, _ := New(DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	blocking := make(chan diff.Result)
	out := Stream(ctx, s, blocking)
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected closed channel after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("stream did not stop after context cancel")
	}
}
