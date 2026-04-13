// Package flatten provides utilities for flattening nested log entry fields
// into a single-level map using a configurable separator.
package flatten

import (
	"fmt"
	"strings"

	"github.com/yourorg/logdrift/internal/diff"
)

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Separator: ".",
		MaxDepth:  8,
	}
}

// Options controls flattening behaviour.
type Options struct {
	// Separator is placed between parent and child key segments.
	Separator string
	// MaxDepth limits recursive expansion; 0 means unlimited.
	MaxDepth int
}

// Flattener expands nested fields inside a log entry's Extra map.
type Flattener struct {
	opts Options
}

// New creates a Flattener with the provided options.
func New(opts Options) (*Flattener, error) {
	if opts.Separator == "" {
		return nil, fmt.Errorf("flatten: separator must not be empty")
	}
	return &Flattener{opts: opts}, nil
}

// Apply returns a copy of entry with all nested Extra values flattened.
func (f *Flattener) Apply(entry diff.Entry) diff.Entry {
	out := entry
	flat := make(map[string]string, len(entry.Extra))
	for k, v := range entry.Extra {
		flattenValue(flat, k, v, f.opts.Separator, 0, f.opts.MaxDepth)
	}
	out.Extra = flat
	return out
}

// flattenValue recursively walks v and writes leaf strings into dst.
func flattenValue(dst map[string]string, prefix, value, sep string, depth, maxDepth int) {
	if maxDepth > 0 && depth >= maxDepth {
		dst[prefix] = value
		return
	}
	// Attempt to detect simple "key=value" pairs encoded as a single string.
	if strings.Contains(value, "{") {
		// Treat as opaque — store as-is.
		dst[prefix] = value
		return
	}
	parts := strings.SplitN(value, sep, 2)
	if len(parts) == 2 && strings.Contains(parts[0], "=") {
		for _, pair := range strings.Split(value, " ") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				key := prefix + sep + strings.TrimSpace(kv[0])
				flattenValue(dst, key, strings.TrimSpace(kv[1]), sep, depth+1, maxDepth)
			}
		}
		return
	}
	dst[prefix] = value
}

// Stream applies the flattener to every entry received from in and forwards
// the result to the returned channel. It closes the output when in is closed
// or ctx is done.
func Stream(ctx interface{ Done() <-chan struct{} }, in <-chan diff.Entry, f *Flattener) <-chan diff.Entry {
	out := make(chan diff.Entry)
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
				select {
				case out <- f.Apply(e):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
