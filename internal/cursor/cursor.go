// Package cursor tracks read positions within log streams, enabling
// resumable tailing across restarts without re-processing old entries.
package cursor

import (
	"errors"
	"sync"
)

// ErrNotFound is returned when no cursor exists for a given key.
var ErrNotFound = errors.New("cursor: key not found")

// Position holds the byte offset and line count for a single stream.
type Position struct {
	Offset int64
	Line   int64
}

// Store holds cursor positions keyed by service or file path.
type Store struct {
	mu      sync.RWMutex
	records map[string]Position
}

// New returns an initialised, empty Store.
func New() *Store {
	return &Store{records: make(map[string]Position)}
}

// Set stores or overwrites the Position for key.
func (s *Store) Set(key string, pos Position) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[key] = pos
}

// Get retrieves the Position for key. Returns ErrNotFound if absent.
func (s *Store) Get(key string) (Position, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.records[key]
	if !ok {
		return Position{}, ErrNotFound
	}
	return p, nil
}

// Delete removes the cursor entry for key. No-op if absent.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, key)
}

// Keys returns all registered keys in an unspecified order.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.records))
	for k := range s.records {
		keys = append(keys, k)
	}
	return keys
}

// Advance increments the stored Position for key by the given delta values.
// If no record exists, it creates one starting from zero.
func (s *Store) Advance(key string, deltaOffset, deltaLines int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.records[key]
	p	p.Line += deltaLines
	s.records[key] = p
}
