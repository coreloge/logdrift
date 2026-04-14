// Package offset tracks per-service byte offsets for log streams,
// enabling resumable tailing after restarts or crashes.
package offset

import (
	"errors"
	"sync"
)

// ErrNotFound is returned when no offset exists for the requested service.
var ErrNotFound = errors.New("offset: service not found")

// Tracker holds the last committed byte offset for each service.
type Tracker struct {
	mu      sync.RWMutex
	offsets map[string]int64
}

// New returns an initialised Tracker with no recorded offsets.
func New() *Tracker {
	return &Tracker{
		offsets: make(map[string]int64),
	}
}

// Set records the current offset for the named service.
func (t *Tracker) Set(service string, offset int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.offsets[service] = offset
}

// Get returns the stored offset for service, or ErrNotFound if none exists.
func (t *Tracker) Get(service string) (int64, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	v, ok := t.offsets[service]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

// Delete removes the offset entry for service. It is a no-op if the service
// was never recorded.
func (t *Tracker) Delete(service string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.offsets, service)
}

// Services returns a snapshot of all service names currently tracked.
func (t *Tracker) Services() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]string, 0, len(t.offsets))
	for k := range t.offsets {
		out = append(out, k)
	}
	return out
}

// Reset removes all stored offsets.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.offsets = make(map[string]int64)
}
