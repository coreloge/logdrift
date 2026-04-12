// Package buffer provides a fixed-size ring buffer for storing recent log entries.
// It is safe for concurrent use and is intended for use in snapshot and replay
// pipelines where only the N most recent entries per service are needed.
package buffer

import (
	"sync"

	"github.com/logdrift/internal/diff"
)

// DefaultCapacity is the number of entries retained when none is specified.
const DefaultCapacity = 256

// Ring is a thread-safe circular buffer of log entries.
type Ring struct {
	mu       sync.Mutex
	buf      []diff.Entry
	cap      int
	head     int // next write position
	size     int // number of valid entries
}

// New returns a Ring with the given capacity.
// If cap is <= 0, DefaultCapacity is used.
func New(cap int) *Ring {
	if cap <= 0 {
		cap = DefaultCapacity
	}
	return &Ring{
		buf: make([]diff.Entry, cap),
		cap: cap,
	}
}

// Push adds an entry to the ring, overwriting the oldest entry when full.
func (r *Ring) Push(e diff.Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = e
	r.head = (r.head + 1) % r.cap
	if r.size < r.cap {
		r.size++
	}
}

// Snapshot returns a slice of the buffered entries in chronological order
// (oldest first). The returned slice is a copy.
func (r *Ring) Snapshot() []diff.Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.size == 0 {
		return nil
	}
	out := make([]diff.Entry, r.size)
	start := (r.head - r.size + r.cap) % r.cap
	for i := 0; i < r.size; i++ {
		out[i] = r.buf[(start+i)%r.cap]
	}
	return out
}

// Len returns the number of entries currently held in the buffer.
func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.size
}

// Reset clears all entries without releasing the underlying memory.
func (r *Ring) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.head = 0
	r.size = 0
}
