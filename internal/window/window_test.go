package window_test

import (
	"testing"
	"time"

	"logdrift/internal/diff"
	"logdrift/internal/window"
)

func makeEntry(service string, ts time.Time) diff.Entry {
	return diff.Entry{
		Service:   service,
		Message:   "test message",
		Level:     "info",
		Timestamp: ts,
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := window.DefaultOptions()
	if opts.Width <= 0 {
		t.Fatalf("expected positive Width, got %v", opts.Width)
	}
	if opts.MaxBuckets <= 0 {
		t.Fatalf("expected positive MaxBuckets, got %d", opts.MaxBuckets)
	}
}

func TestAdd_CreatesBucket(t *testing.T) {
	w := window.New(window.Options{Width: time.Minute, MaxBuckets: 5})
	now := time.Now()
	w.Add(makeEntry("svc-a", now))

	buckets := w.Buckets()
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if len(buckets[0].Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(buckets[0].Entries))
	}
}

func TestAdd_SameBucket(t *testing.T) {
	w := window.New(window.Options{Width: time.Minute, MaxBuckets: 5})
	now := time.Now().Truncate(time.Minute)
	w.Add(makeEntry("svc-a", now))
	w.Add(makeEntry("svc-b", now.Add(5*time.Second)))

	buckets := w.Buckets()
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if len(buckets[0].Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(buckets[0].Entries))
	}
}

func TestAdd_EvictsOldestBuckets(t *testing.T) {
	w := window.New(window.Options{Width: time.Minute, MaxBuckets: 3})
	base := time.Now().Truncate(time.Minute)
	for i := 0; i < 5; i++ {
		w.Add(makeEntry("svc", base.Add(time.Duration(i)*time.Minute)))
	}

	buckets := w.Buckets()
	if len(buckets) != 3 {
		t.Fatalf("expected 3 buckets after eviction, got %d", len(buckets))
	}
}

func TestReset_ClearsBuckets(t *testing.T) {
	w := window.New(window.DefaultOptions())
	w.Add(makeEntry("svc", time.Now()))
	w.Reset()

	if len(w.Buckets()) != 0 {
		t.Fatal("expected empty buckets after Reset")
	}
}
