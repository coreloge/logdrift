// Package rotate provides log rotation detection and automatic re-open
// logic for file-based log sources.
package rotate

import (
	"context"
	"os"
	"time"

	"github.com/user/logdrift/internal/watch"
)

// Options controls rotation detection behaviour.
type Options struct {
	// PollInterval is how often the file is checked for rotation.
	PollInterval time.Duration
	// MaxReopens is the maximum number of consecutive reopen attempts
	// before the rotator gives up (0 = unlimited).
	MaxReopens int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		PollInterval: 500 * time.Millisecond,
		MaxReopens:   10,
	}
}

// Event is emitted on the Events channel when a rotation is detected.
type Event struct {
	Path    string
	Reopens int
}

// Rotator watches a file path for rotation (truncation or inode change)
// and emits an Event whenever the file should be re-opened.
type Rotator struct {
	path    string
	opts    Options
	watcher *watch.Watcher
	Events  <-chan Event
}

// New creates and starts a Rotator for the given file path.
// The returned Rotator emits on its Events channel until ctx is cancelled.
func New(ctx context.Context, path string, opts Options) (*Rotator, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	w, err := watch.New(path, opts.PollInterval)
	if err != nil {
		return nil, err
	}

	evCh := make(chan Event, 8)
	r := &Rotator{
		path:    path,
		opts:    opts,
		watcher: w,
		Events:  evCh,
	}

	go r.run(ctx, evCh)
	return r, nil
}

func (r *Rotator) run(ctx context.Context, out chan<- Event) {
	defer close(out)
	reopens := 0
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-r.watcher.Events:
			if !ok {
				return
			}
			reopens++
			if r.opts.MaxReopens > 0 && reopens > r.opts.MaxReopens {
				return
			}
			select {
			case out <- Event{Path: r.path, Reopens: reopens}:
			case <-ctx.Done():
				return
			}
		}
	}
}
