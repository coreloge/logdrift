// Package snapshot provides functionality for capturing and comparing
// point-in-time snapshots of log entry streams across services.
package snapshot

import (
	"sync"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// Snapshot holds a captured set of log entries keyed by service name.
type Snapshot struct {
	CapturedAt time.Time
	Entries    map[string][]diff.Entry
	mu         sync.RWMutex
}

// New creates an empty Snapshot with the current timestamp.
func New() *Snapshot {
	return &Snapshot{
		CapturedAt: time.Now(),
		Entries:    make(map[string][]diff.Entry),
	}
}

// Add appends a log entry under the given service key.
func (s *Snapshot) Add(service string, entry diff.Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Entries[service] = append(s.Entries[service], entry)
}

// Services returns a sorted list of service names present in the snapshot.
func (s *Snapshot) Services() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.Entries))
	for k := range s.Entries {
		names = append(names, k)
	}
	return names
}

// Get returns the entries for the given service, or nil if not present.
func (s *Snapshot) Get(service string) []diff.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Entries[service]
}

// Compare diffs two snapshots for a given service and returns all deltas.
func Compare(baseline, current *Snapshot, service string) []diff.DiffResult {
	baseEntries := baseline.Get(service)
	currEntries := current.Get(service)

	results := make([]diff.DiffResult, 0)
	limit := len(baseEntries)
	if len(currEntries) < limit {
		limit = len(currEntries)
	}
	for i := 0; i < limit; i++ {
		results = append(results, diff.Compare(baseEntries[i], currEntries[i]))
	}
	return results
}
