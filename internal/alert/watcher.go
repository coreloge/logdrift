package alert

import (
	"context"

	"github.com/user/logdrift/internal/diff"
)

// ResultStream carries labelled diff results from a snapshot comparison pass.
type ResultStream struct {
	Service string
	Result  diff.Result
}

// Watcher consumes a channel of ResultStream values, evaluates each against
// the provided Config, and forwards any triggered Alerts to the Out channel.
type Watcher struct {
	Out <-chan Alert
	cfg Config
	out chan Alert
}

// NewWatcher creates a Watcher that reads from src until it is closed or ctx
// is cancelled. Alerts are available on the returned Watcher.Out channel.
func NewWatcher(ctx context.Context, src <-chan ResultStream, cfg Config) *Watcher {
	ch := make(chan Alert, 16)
	w := &Watcher{
		Out: ch,
		cfg: cfg,
		out: ch,
	}
	go w.run(ctx, src)
	return w
}

func (w *Watcher) run(ctx context.Context, src <-chan ResultStream) {
	defer close(w.out)
	for {
		select {
		case <-ctx.Done():
			return
		case rs, ok := <-src:
			if !ok {
				return
			}
			if a := Evaluate(rs.Service, rs.Result, w.cfg); a != nil {
				select {
				case w.out <- *a:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
