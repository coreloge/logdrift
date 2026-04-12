package redact_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/redact"
)

func feedEntries(entries []diff.LogEntry) <-chan diff.LogEntry {
	ch := make(chan diff.LogEntry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drainStream(ch <-chan diff.LogEntry) []diff.LogEntry {
	var out []diff.LogEntry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestStream_RedactsFields(t *testing.T) {
	r := redact.New(redact.Options{
		Rules: []redact.Rule{
			{Field: "secret", Replacement: "[REDACTED]"},
		},
	})
	in := feedEntries([]diff.LogEntry{
		{Service: "svc", Message: "hello", Fields: map[string]string{"secret": "abc", "user": "bob"}},
	})
	out := redact.Stream(context.Background(), r, in)
	entries := drainStream(out)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Fields["secret"] != "[REDACTED]" {
		t.Errorf("expected secret redacted, got %q", entries[0].Fields["secret"])
	}
	if entries[0].Fields["user"] != "bob" {
		t.Errorf("expected user unchanged, got %q", entries[0].Fields["user"])
	}
}

func TestStream_PassesThroughUnaffectedEntries(t *testing.T) {
	r := redact.New(redact.Options{})
	in := feedEntries([]diff.LogEntry{
		{Service: "a", Message: "msg1", Level: "info"},
		{Service: "b", Message: "msg2", Level: "warn"},
	})
	out := redact.Stream(context.Background(), r, in)
	entries := drainStream(out)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	r := redact.New(redact.DefaultOptions())
	blocking := make(chan diff.LogEntry)
	ctx, cancel := context.WithCancel(context.Background())
	out := redact.Stream(ctx, r, blocking)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to close")
	}
}
