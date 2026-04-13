// Package topology tracks the dependency relationships between services
// observed in the log stream, inferring edges from correlation IDs and
// service labels that appear together within the same log entries.
package topology

import (
	"sync"

	"github.com/logdrift/internal/diff"
)

// Edge represents a directed dependency between two services.
type Edge struct {
	From string
	To   string
}

// Graph holds the inferred service topology.
type Graph struct {
	mu    sync.RWMutex
	edges map[Edge]int // edge -> observation count
}

// New returns an empty topology Graph.
func New() *Graph {
	return &Graph{
		edges: make(map[Edge]int),
	}
}

// Record inspects a log entry and records an edge when the entry carries
// both a "service" label and a non-empty "caller" field that names a
// different service.
func (g *Graph) Record(entry diff.Entry) {
	caller, ok := entry.Fields["caller"]
	if !ok || caller == "" {
		return
	}
	if entry.Service == "" || caller == entry.Service {
		return
	}
	e := Edge{From: entry.Service, To: caller}
	g.mu.Lock()
	g.edges[e]++
	g.mu.Unlock()
}

// Edges returns a snapshot of all observed edges and their counts.
func (g *Graph) Edges() map[Edge]int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make(map[Edge]int, len(g.edges))
	for e, c := range g.edges {
		out[e] = c
	}
	return out
}

// Neighbours returns the set of services that "svc" has been observed
// calling (outbound edges).
func (g *Graph) Neighbours(svc string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	seen := map[string]struct{}{}
	for e := range g.edges {
		if e.From == svc {
			seen[e.To] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	return out
}

// Reset clears all recorded edges.
func (g *Graph) Reset() {
	g.mu.Lock()
	g.edges = make(map[Edge]int)
	g.mu.Unlock()
}
