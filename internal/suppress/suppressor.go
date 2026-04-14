// Package suppress provides a pipeline stage that suppresses repeated log
// entries matching a given pattern or field value for a configurable cooldown
// period. Once suppressed, matching entries are dropped until the cooldown
// expires, at which point the next matching entry is forwarded again.
package suppress

import (
	"regexp"
	"sync"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// Options configures the Suppressor.
type Options struct {
	// Field is the entry field to match against ("message", or any Extra key).
	// Defaults to "message".
	Field string
	// Pattern is a regular expression matched against the field value.
	// If empty, all entries are candidates for suppression.
	Pattern string
	// Cooldown is the minimum duration between forwarded matching entries.
	// Defaults to 5 seconds.
	Cooldown time.Duration
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Field:    "message",
		Cooldown: 5 * time.Second,
	}
}

// Suppressor drops repeated matching entries within a cooldown window.
type Suppressor struct {
	opts    Options
	pattern *regexp.Regexp
	mu      sync.Mutex
	lastSeen map[string]time.Time
}

// New creates a Suppressor from opts. Returns an error if Pattern is set but
// cannot be compiled.
func New(opts Options) (*Suppressor, error) {
	if opts.Field == "" {
		opts.Field = "message"
	}
	if opts.Cooldown <= 0 {
		opts.Cooldown = 5 * time.Second
	}
	var re *regexp.Regexp
	if opts.Pattern != "" {
		var err error
		re, err = regexp.Compile(opts.Pattern)
		if err != nil {
			return nil, err
		}
	}
	return &Suppressor{
		opts:     opts,
		pattern:  re,
		lastSeen: make(map[string]time.Time),
	}, nil
}

// Allow returns true if the entry should be forwarded, false if it should be
// suppressed. It is safe for concurrent use.
func (s *Suppressor) Allow(e diff.LogEntry) bool {
	val := fieldValue(e, s.opts.Field)
	if s.pattern != nil && !s.pattern.MatchString(val) {
		return true
	}
	key := e.Service + "\x00" + val
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if last, ok := s.lastSeen[key]; ok && now.Sub(last) < s.opts.Cooldown {
		return false
	}
	s.lastSeen[key] = now
	return true
}

func fieldValue(e diff.LogEntry, field string) string {
	switch field {
	case "message":
		return e.Message
	case "level":
		return e.Level
	default:
		if e.Extra != nil {
			if v, ok := e.Extra[field]; ok {
				return v
			}
		}
		return ""
	}
}
