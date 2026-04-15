// Package pivot groups log entries by a chosen field and counts occurrences
// per service, producing a cross-tabulation useful for spotting drift patterns.
package pivot

import (
	"errors"
	"sync"

	"github.com/yourorg/logdrift/internal/diff"
)

// Options controls pivot behaviour.
type Options struct {
	// Field is the entry field to pivot on (e.g. "level", "status").
	Field string
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{Field: "level"}
}

// Cell holds the count for a (service, fieldValue) pair.
type Cell struct {
	Service string
	Value   string
	Count   int
}

// Table is the result of a pivot operation.
type Table struct {
	// Rows maps fieldValue -> service -> count.
	Rows map[string]map[string]int
}

// Services returns a sorted-stable slice of all service names seen.
func (t *Table) Services() []string {
	seen := map[string]struct{}{}
	for _, row := range t.Rows {
		for svc := range row {
			seen[svc] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	return out
}

// Pivoter accumulates entries and builds pivot tables on demand.
type Pivoter struct {
	opts Options
	mu   sync.Mutex
	data map[string]map[string]int // fieldValue -> service -> count
}

// New creates a Pivoter. Returns an error if Field is empty.
func New(opts Options) (*Pivoter, error) {
	if opts.Field == "" {
		return nil, errors.New("pivot: Field must not be empty")
	}
	return &Pivoter{
		opts: opts,
		data: make(map[string]map[string]int),
	}, nil
}

// Record ingests a single log entry.
func (p *Pivoter) Record(e diff.Entry) {
	var val string
	switch p.opts.Field {
	case "level":
		val = e.Level
	case "message":
		val = e.Message
	default:
		v, ok := e.Extra[p.opts.Field]
		if !ok {
			return
		}
		val = v
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.data[val] == nil {
		p.data[val] = make(map[string]int)
	}
	p.data[val][e.Service]++
}

// Table returns a snapshot of the current pivot table.
func (p *Pivoter) Table() Table {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := Table{Rows: make(map[string]map[string]int, len(p.data))}
	for val, row := range p.data {
		copy := make(map[string]int, len(row))
		for svc, cnt := range row {
			copy[svc] = cnt
		}
		out.Rows[val] = copy
	}
	return out
}

// Reset clears all accumulated data.
func (p *Pivoter) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = make(map[string]map[string]int)
}
