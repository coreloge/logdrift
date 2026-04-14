package severity

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// FilterOptions controls which entries are forwarded by Filter.
type FilterOptions struct {
	// MinLevel is the minimum severity an entry must have to pass through.
	// Entries whose level cannot be parsed are treated as Unknown and will be
	// dropped when MinLevel > Unknown.
	MinLevel Level
}

// DefaultFilterOptions returns options that pass every entry through.
func DefaultFilterOptions() FilterOptions {
	return FilterOptions{MinLevel: Unknown}
}

// Filter reads entries from in, drops those whose level is below opts.MinLevel,
// and forwards the rest to the returned channel. The channel is closed when ctx
// is cancelled or in is closed.
func Filter(ctx context.Context, in <-chan diff.Entry, opts FilterOptions) <-chan diff.Entry {
	out := make(chan diff.Entry, 64)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-in:
				if !ok {
					return
				}
				if Parse(e.Level).AtLeast(opts.MinLevel) {
					select {
					case out <- e:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
