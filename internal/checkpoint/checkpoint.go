// Package checkpoint persists and restores log tail offsets so that
// logdrift can resume from where it left off after a restart.
package checkpoint

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

// ErrNotFound is returned when no checkpoint exists for a given service.
var ErrNotFound = errors.New("checkpoint: no entry found")

// Store holds per-service byte offsets on disk.
type Store struct {
	mu   sync.RWMutex
	path string
	data map[string]int64
}

// New opens (or creates) a checkpoint file at path.
func New(path string) (*Store, error) {
	s := &Store{path: path, data: make(map[string]int64)}
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return s, nil
}

// Set records the current offset for service.
func (s *Store) Set(service string, offset int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[service] = offset
	return s.flush()
}

// Get returns the stored offset for service, or ErrNotFound.
func (s *Store) Get(service string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[service]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

// Delete removes the checkpoint entry for service.
func (s *Store) Delete(service string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, service)
	return s.flush()
}

// Services returns all service names that have a stored checkpoint.
func (s *Store) Services() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &s.data)
}

func (s *Store) flush() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}
