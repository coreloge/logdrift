package score

import (
	"context"

	"github.com/logdrift/logdrift/internal/diff"
)

// ScoredResult pairs a diff.Result with its computed drift score.
type ScoredResult struct {
	Result diff.Result
	Score  float64
}

// Stream reads diff.Results from in, annotates each with a drift score, and
// forwards ScoredResults to the returned channel. The channel is closed when
// ctx is cancelled or in is drained.
func Stream(ctx context.Context, s *Scorer, in <-chan diff.Result) <-chan ScoredResult {
	out := make(chan ScoredResult)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-in:
				if !ok {
					return
				}
				sr := ScoredResult{
					Result: r,
					Score:  s.Compute(r),
				}
				select {
				case out <- sr:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
