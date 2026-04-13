package partition_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/partition"
)

func makeEntry(service, msg string, fields map[string]string) diff.Entry {
	return diff.Entry{
		Service: service,
		Message: msg,
		Level:   "info",
		Fields:  fields,
	}
}

func feedEntries(entries []diff.Entry) <-chan diff.Entry {
	ch := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func TestDefaultOptions(t *testing.T) {
	opts := partition.DefaultOptions()
	if opts.Field != "service" {
		t.Fatalf("expected field 'service', got %q", opts.Field)
	}
	if opts.BufferSize != 64 {
		t.Fatalf("expected buffer 64, got %d", opts.BufferSize)
	}
}

func TestStream_PartitionsByService(t *testing.T) {
	p := partition.New(partition.DefaultOptions())
	entries := []diff.Entry{
		makeEntry("api", "req", nil),
		makeEntry("worker", "job", nil),
		makeEntry("api", "resp", nil),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p.Stream(ctx, feedEntries(entries))

	apiCh := p.Bucket("api")
	if len(apiCh) != 2 {
		t.Fatalf("expected 2 entries in api bucket, got %d", len(apiCh))
	}
	workerCh := p.Bucket("worker")
	if len(workerCh) != 1 {
		t.Fatalf("expected 1 entry in worker bucket, got %d", len(workerCh))
	}
}

func TestStream_PartitionsByCustomField(t *testing.T) {
	opts := partition.DefaultOptions()
	opts.Field = "tenant"
	p := partition.New(opts)

	entries := []diff.Entry{
		makeEntry("svc", "a", map[string]string{"tenant": "acme"}),
		makeEntry("svc", "b", map[string]string{"tenant": "globex"}),
		makeEntry("svc", "c", map[string]string{"tenant": "acme"}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p.Stream(ctx, feedEntries(entries))

	if len(p.Bucket("acme")) != 2 {
		t.Fatalf("expected 2 entries for acme")
	}
	if len(p.Bucket("globex")) != 1 {
		t.Fatalf("expected 1 entry for globex")
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	p := partition.New(partition.DefaultOptions())
	blocking := make(chan diff.Entry) // never closed

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		p.Stream(ctx, blocking)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stream did not stop after context cancellation")
	}
}

func TestKeys_ReturnsAllPartitions(t *testing.T) {
	p := partition.New(partition.DefaultOptions())
	entries := []diff.Entry{
		makeEntry("alpha", "x", nil),
		makeEntry("beta", "y", nil),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p.Stream(ctx, feedEntries(entries))

	keys := p.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}
