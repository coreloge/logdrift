// Package route provides entry routing — directing log entries to named
// output channels based on configurable field-matching rules.
package route

import (
	"context"
	"strings"

	"github.com/user/logdrift/internal/diff"
)

// Rule describes a single routing rule.
type Rule struct {
	// Name is a human-readable label for the route.
	Name string
	// Field is the entry field to match against (e.g. "level", "service").
	Field string
	// Values is the set of accepted values for Field (case-insensitive).
	Values []string
}

// Options holds configuration for a Router.
type Options struct {
	Rules []Rule
	// DefaultRoute receives entries that match no rule. If empty, unmatched
	// entries are dropped.
	DefaultRoute string
}

// DefaultOptions returns a zero-value Options that drops unmatched entries.
func DefaultOptions() Options { return Options{} }

// Router dispatches log entries to named output channels.
type Router struct {
	opts   Options
	routes map[string]chan diff.Entry
}

// New creates a Router. Callers must call Run to start dispatching.
func New(opts Options) *Router {
	return &Router{
		opts:   opts,
		routes: make(map[string]chan diff.Entry),
	}
}

// Subscribe returns a channel that receives entries routed to name.
// The channel is buffered with capacity 64.
func (r *Router) Subscribe(name string) <-chan diff.Entry {
	ch := make(chan diff.Entry, 64)
	r.routes[name] = ch
	return ch
}

// Run reads from in and dispatches entries until ctx is cancelled or in is
// closed.
func (r *Router) Run(ctx context.Context, in <-chan diff.Entry) {
	defer r.closeAll()
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-in:
			if !ok {
				return
			}
			r.dispatch(e)
		}
	}
}

func (r *Router) dispatch(e diff.Entry) {
	for _, rule := range r.opts.Rules {
		if r.matches(e, rule) {
			r.send(rule.Name, e)
			return
		}
	}
	if r.opts.DefaultRoute != "" {
		r.send(r.opts.DefaultRoute, e)
	}
}

func (r *Router) matches(e diff.Entry, rule Rule) bool {
	var val string
	switch strings.ToLower(rule.Field) {
	case "level":
		val = e.Level
	case "service":
		val = e.Service
	default:
		val = e.Fields[rule.Field]
	}
	val = strings.ToLower(val)
	for _, v := range rule.Values {
		if strings.ToLower(v) == val {
			return true
		}
	}
	return false
}

func (r *Router) send(name string, e diff.Entry) {
	if ch, ok := r.routes[name]; ok {
		select {
		case ch <- e:
		default:
		}
	}
}

func (r *Router) closeAll() {
	for _, ch := range r.routes {
		close(ch)
	}
}
