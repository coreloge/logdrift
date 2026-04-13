package topology_test

import (
	"sort"
	"testing"

	"github.com/logdrift/internal/diff"
	"github.com/logdrift/internal/topology"
)

func makeEntry(service, caller string) diff.Entry {
	fields := map[string]string{}
	if caller != "" {
		fields["caller"] = caller
	}
	return diff.Entry{
		Service: service,
		Level:   "info",
		Message: "test message",
		Fields:  fields,
	}
}

func TestNew_StartsEmpty(t *testing.T) {
	g := topology.New()
	if got := g.Edges(); len(got) != 0 {
		t.Fatalf("expected empty graph, got %d edges", len(got))
	}
}

func TestRecord_AddsEdge(t *testing.T) {
	g := topology.New()
	g.Record(makeEntry("api", "db"))
	edges := g.Edges()
	key := topology.Edge{From: "api", To: "db"}
	if edges[key] != 1 {
		t.Fatalf("expected count 1 for edge api->db, got %d", edges[key])
	}
}

func TestRecord_IncrementsDuplicateEdge(t *testing.T) {
	g := topology.New()
	g.Record(makeEntry("api", "db"))
	g.Record(makeEntry("api", "db"))
	key := topology.Edge{From: "api", To: "db"}
	if got := g.Edges()[key]; got != 2 {
		t.Fatalf("expected count 2, got %d", got)
	}
}

func TestRecord_IgnoresMissingCaller(t *testing.T) {
	g := topology.New()
	g.Record(makeEntry("api", ""))
	if len(g.Edges()) != 0 {
		t.Fatal("expected no edges when caller is absent")
	}
}

func TestRecord_IgnoresSelfLoop(t *testing.T) {
	g := topology.New()
	g.Record(makeEntry("api", "api"))
	if len(g.Edges()) != 0 {
		t.Fatal("expected no self-loop edge")
	}
}

func TestNeighbours_ReturnsOutboundServices(t *testing.T) {
	g := topology.New()
	g.Record(makeEntry("api", "db"))
	g.Record(makeEntry("api", "cache"))
	g.Record(makeEntry("worker", "db"))

	got := g.Neighbours("api")
	sort.Strings(got)
	if len(got) != 2 || got[0] != "cache" || got[1] != "db" {
		t.Fatalf("unexpected neighbours: %v", got)
	}
}

func TestReset_ClearsGraph(t *testing.T) {
	g := topology.New()
	g.Record(makeEntry("api", "db"))
	g.Reset()
	if len(g.Edges()) != 0 {
		t.Fatal("expected empty graph after reset")
	}
}
