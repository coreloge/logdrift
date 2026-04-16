package proxy_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/logdrift/logdrift/internal/diff"
	"github.com/logdrift/logdrift/internal/proxy"
)

func TestStream_HookCalledForEachEntry(t *testing.T) {
	var count int64
	hook := func(_ diff.Entry) { atomic.AddInt64(&count, 1) }

	p, err := proxy.New(proxy.Options{Hook: hook})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := []diff.Entry{
		{Service: "a", Level: "info", Message: "one"},
		{Service: "b", Level: "warn", Message: "two"},
	}

	in := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		in <- e
	}
	close(in)

	out := proxy.Stream(context.Background(), p, in)
	var collected []diff.Entry
	for e := range out {
		collected = append(collected, e)
	}

	if int(count) != len(entries) {
		t.Fatalf("hook called %d times, want %d", count, len(entries))
	}
	if len(collected) != len(entries) {
		t.Fatalf("collected %d entries, want %d", len(collected), len(entries))
	}
}

func TestStream_SeenMatchesProcessed(t *testing.T) {
	p, _ := proxy.New(proxy.DefaultOptions())

	in := make(chan diff.Entry, 3)
	for i := 0; i < 3; i++ {
		in <- diff.Entry{Message: "msg"}
	}
	close(in)

	out := proxy.Stream(context.Background(), p, in)
	for range out {
	}

	if p.Seen() != 3 {
		t.Fatalf("expected Seen()=3, got %d", p.Seen())
	}
}
