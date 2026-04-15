// Package clone provides deep-copy utilities for log entries,
// ensuring pipeline stages cannot mutate shared state.
package clone

import "github.com/logdrift/logdrift/internal/diff"

// Entry returns a deep copy of the given log entry. The Extra map
// is duplicated so that downstream mutations do not affect the original.
func Entry(e diff.Entry) diff.Entry {
	out := diff.Entry{
		Service:   e.Service,
		Level:     e.Level,
		Message:   e.Message,
		Timestamp: e.Timestamp,
	}
	if len(e.Extra) > 0 {
		out.Extra = make(map[string]string, len(e.Extra))
		for k, v := range e.Extra {
			out.Extra[k] = v
		}
	}
	return out
}

// Stream reads entries from in, emits a deep copy of each to the returned
// channel, and closes it when in is drained or ctx is done.
func Stream(ctx interface{ Done() <-chan struct{} }, in <-chan diff.Entry) <-chan diff.Entry {
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
				case out <- Entry(e):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
