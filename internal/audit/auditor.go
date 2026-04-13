// Package audit provides a structured audit trail for log entries,
// recording when entries were observed, by which service, and any
// drift events that were detected during processing.
package audit

import (
	"sync"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/diff"
)

// Record represents a single audit event.
type Record struct {
	Timestamp time.Time
	Service   string
	Message   string
	Level     string
	Drift     bool
	Deltas    []diff.Delta
}

// Options configures the Auditor.
type Options struct {
	// MaxRecords is the maximum number of records retained in memory.
	// Defaults to 1000.
	MaxRecords int
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{MaxRecords: 1000}
}

// Auditor accumulates audit records for observed log entries and drift events.
type Auditor struct {
	mu      sync.Mutex
	opts    Options
	records []Record
}

// New creates a new Auditor with the provided options.
func New(opts Options) *Auditor {
	if opts.MaxRecords <= 0 {
		opts.MaxRecords = DefaultOptions().MaxRecords
	}
	return &Auditor{opts: opts}
}

// RecordEntry appends an audit record for a plain log entry.
func (a *Auditor) RecordEntry(service, level, message string) {
	a.append(Record{
		Timestamp: time.Now(),
		Service:   service,
		Level:     level,
		Message:   message,
	})
}

// RecordDrift appends an audit record for a detected drift event.
func (a *Auditor) RecordDrift(service string, deltas []diff.Delta) {
	a.append(Record{
		Timestamp: time.Now(),
		Service:   service,
		Drift:     true,
		Deltas:    deltas,
	})
}

// All returns a copy of all retained audit records in insertion order.
func (a *Auditor) All() []Record {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]Record, len(a.records))
	copy(out, a.records)
	return out
}

// Reset clears all retained audit records.
func (a *Auditor) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.records = a.records[:0]
}

func (a *Auditor) append(r Record) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.records) >= a.opts.MaxRecords {
		// drop oldest
		a.records = a.records[1:]
	}
	a.records = append(a.records, r)
}
