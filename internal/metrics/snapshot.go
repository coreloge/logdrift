package metrics

import (
	"sync"
	"time"
)

// SnapshotRecord holds a point-in-time capture of counter values.
type SnapshotRecord struct {
	Timestamp  time.Time
	Entries    map[string]int64
	Drifts     map[string]int64
}

// SnapshotStore retains a bounded history of metric snapshots.
type SnapshotStore struct {
	mu      sync.RWMutex
	records []SnapshotRecord
	maxLen  int
}

// NewSnapshotStore creates a SnapshotStore that keeps at most maxLen records.
// If maxLen is <= 0 it defaults to 60.
func NewSnapshotStore(maxLen int) *SnapshotStore {
	if maxLen <= 0 {
		maxLen = 60
	}
	return &SnapshotStore{maxLen: maxLen}
}

// Capture takes a snapshot of the given Counter and appends it to the store.
func (s *SnapshotStore) Capture(c *Counter) {
	rec := SnapshotRecord{
		Timestamp: time.Now().UTC(),
		Entries:   c.Entries(),
		Drifts:    c.Drifts(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.records) >= s.maxLen {
		s.records = s.records[1:]
	}
	s.records = append(s.records, rec)
}

// All returns a copy of all stored snapshot records in chronological order.
func (s *SnapshotStore) All() []SnapshotRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SnapshotRecord, len(s.records))
	copy(out, s.records)
	return out
}

// Latest returns the most recent snapshot, and false if the store is empty.
func (s *SnapshotStore) Latest() (SnapshotRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.records) == 0 {
		return SnapshotRecord{}, false
	}
	return s.records[len(s.records)-1], true
}

// Len returns the number of snapshots currently held.
func (s *SnapshotStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}
