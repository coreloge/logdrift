// Package keep filters a log entry stream, forwarding only entries that
// match at least one of the configured rules. It is the logical inverse of
// the drop package.
package keep

import (
	"context"
	"fmt"
	"regexp"

	"github.com/yourorg/logdrift/internal/diff"
)

// Rule describes a single keep condition.
type Rule struct {
	Field   string // empty string matches the message field
	Pattern string
}

// Options configures the Keeper.
type Options struct {
	Rules []Rule
}

// DefaultOptions returns an Options that keeps everything.
func DefaultOptions() Options { return Options{} }

// Keeper forwards only entries that satisfy at least one rule.
type Keeper struct {
	patterns []compiled
}

type compiled struct {
	field string
	re    *regexp.Regexp
}

// New creates a Keeper from opts. Returns an error if any pattern is invalid.
func New(opts Options) (*Keeper, error) {
	var cs []compiled
	for _, r := range opts.Rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("keep: empty pattern for field %q", r.Field)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("keep: invalid pattern %q: %w", r.Pattern, err)
		}
		cs = append(cs, compiled{field: r.Field, re: re})
	}
	return &Keeper{patterns: cs}, nil
}

// ShouldKeep returns true when the entry matches at least one rule, or when
// no rules are configured.
func (k *Keeper) ShouldKeep(e diff.Entry) bool {
	if len(k.patterns) == 0 {
		return true
	}
	for _, c := range k.patterns {
		var val string
		if c.field == "" || c.field == "message" {
			val = e.Message
		} else if c.field == "level" {
			val = e.Level
		} else if c.field == "service" {
			val = e.Service
		} else {
			val, _ = e.Extra[c.field].(string)
		}
		if c.re.MatchString(val) {
			return true
		}
	}
	return false
}

// Stream reads from in, forwards matching entries to the returned channel, and
// closes it when ctx is done or in is closed.
func Stream(ctx context.Context, in <-chan diff.Entry, k *Keeper) <-chan diff.Entry {
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
				if k.ShouldKeep(e) {
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
