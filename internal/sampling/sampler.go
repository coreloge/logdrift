// Package sampling provides log entry rate-limiting and sampling strategies
// for reducing noise in high-volume log streams.
package sampling

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// Strategy defines how entries are sampled from a stream.
type Strategy int

const (
	// StrategyNone passes every entry through unchanged.
	StrategyNone Strategy = iota
	// StrategyRandom drops entries with probability (1 - rate).
	StrategyRandom
	// StrategyRateLimit passes at most N entries per second per service.
	StrategyRateLimit
)

// Options configures the sampler behaviour.
type Options struct {
	Strategy  Strategy
	// Rate is the fraction of entries to keep (0.0–1.0) for StrategyRandom,
	// or the maximum entries-per-second for StrategyRateLimit.
	Rate      float64
}

// DefaultOptions returns a pass-through sampler configuration.
func DefaultOptions() Options {
	return Options{Strategy: StrategyNone, Rate: 1.0}
}

// Sampler filters a stream of log entries according to a chosen strategy.
type Sampler struct {
	opts     Options
	counters map[string]*atomic.Int64
	ticker   *time.Ticker
	rng      *rand.Rand
}

// New creates a Sampler with the given options.
func New(opts Options) *Sampler {
	s := &Sampler{
		opts:     opts,
		counters: make(map[string]*atomic.Int64),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	if opts.Strategy == StrategyRateLimit && opts.Rate > 0 {
		s.ticker = time.NewTicker(time.Second)
		go s.resetLoop()
	}
	return s
}

// Accept reports whether the given entry should be forwarded downstream.
func (s *Sampler) Accept(entry diff.Entry) bool {
	switch s.opts.Strategy {
	case StrategyRandom:
		return s.rng.Float64() < s.opts.Rate
	case StrategyRateLimit:
		ctr := s.counterFor(entry.Service)
		limit := int64(s.opts.Rate)
		if limit <= 0 {
			return false
		}
		if ctr.Add(1) > limit {
			return false
		}
		return true
	default:
		return true
	}
}

func (s *Sampler) counterFor(service string) *atomic.Int64 {
	if c, ok := s.counters[service]; ok {
		return c
	}
	c := &atomic.Int64{}
	s.counters[service] = c
	return c
}

func (s *Sampler) resetLoop() {
	for range s.ticker.C {
		for _, c := range s.counters {
			c.Store(0)
		}
	}
}

// Stop releases resources held by the sampler.
func (s *Sampler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
