// Package elect provides a lightweight leader-election utility for
// coordinating which service instance is considered the "primary" when
// diffing log streams from replicated services.
package elect

import (
	"errors"
	"sync"
	"time"
)

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		TTL:      5 * time.Second,
		RenewEvery: time.Second,
	}
}

// Options controls election behaviour.
type Options struct {
	// TTL is how long a lease remains valid without renewal.
	TTL time.Duration
	// RenewEvery is how often the current leader renews its lease.
	RenewEvery time.Duration
}

// Elector tracks which candidate currently holds the lease.
type Elector struct {
	mu        sync.Mutex
	opts      Options
	leader    string
	expiry    time.Time
	now       func() time.Time
}

// New creates an Elector with the given options.
func New(opts Options) (*Elector, error) {
	if opts.TTL <= 0 {
		return nil, errors.New("elect: TTL must be positive")
	}
	if opts.RenewEvery <= 0 {
		return nil, errors.New("elect: RenewEvery must be positive")
	}
	return &Elector{opts: opts, now: time.Now}, nil
}

// Acquire attempts to acquire or renew the lease for candidate.
// Returns true if candidate is (or becomes) the leader.
func (e *Elector) Acquire(candidate string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.now()
	if e.leader == "" || now.After(e.expiry) {
		e.leader = candidate
		e.expiry = now.Add(e.opts.TTL)
		return true
	}
	if e.leader == candidate {
		e.expiry = now.Add(e.opts.TTL)
		return true
	}
	return false
}

// Leader returns the current leader and whether the lease is still valid.
func (e *Elector) Leader() (string, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leader == "" || e.now().After(e.expiry) {
		return "", false
	}
	return e.leader, true
}

// Revoke forcibly clears the current lease.
func (e *Elector) Revoke() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.leader = ""
	e.expiry = time.Time{}
}
