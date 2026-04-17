package shard_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
	"github.com/logdrift/logdrift/internal/shard"
)

func makeEntry(service, level, msg string) diff.Entry {
	return diff.Entry{Service: service, Level: level, Message: msg}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := shard.New(shard.Options{Field: "", Shards: 2})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_ZeroShards_ReturnsError(t *testing.T) {
	_, err := shard.New(shard.Options{Field: "service", Shards: 0})
	if err == nil {
		t.Fatal("expected error for zero shards")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	s, err := shard.New(shard.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Outputs()) != 4 {
		t.Fatalf("expected 4 outputs, got %d", len(s.Outputs()))
	}
}

func TestAssign_DistributesEntries(t *testing.T) {
	s, _ := shard.New(shard.Options{Field: "service", Shards: 2})
	services := []string{"alpha", "beta", "gamma", "alpha", "beta"}
	for _, svc := range services {
		s.Assign(makeEntry(svc, "info", "msg"))
	}
	s.Close()

	total := 0
	for _, out := range s.Outputs() {
		for range out {
			total++
		}
	}
	if total != len(services) {
		t.Fatalf("expected %d entries total, got %d", len(services), total)
	}
}

func TestAssign_SameServiceSameShard(t *testing.T) {
	s, _ := shard.New(shard.Options{Field: "service", Shards: 4})
	e1 := makeEntry("svc-x", "info", "a")
	e2 := makeEntry("svc-x", "warn", "b")
	s.Assign(e1)
	s.Assign(e2)
	s.Close()

	counts := make([]int, 4)
	for i, out := range s.Outputs() {
		for range out {
			counts[i]++
		}
	}
	for i, c := range counts {
		if c == 2 {
			return // both ended up in the same shard
			_ = i
		}
	}
	t.Fatal("expected both entries in the same shard")
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	s, _ := shard.New(shard.Options{Field: "service", Shards: 2})
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	shard.Stream(ctx, in, s)
	cancel()
	time.Sleep(20 * time.Millisecond)
	// outputs should be closed; draining should not block
	for _, out := range s.Outputs() {
		for range out {
		}
	}
}
