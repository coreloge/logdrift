package elect

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{Service: svc, Message: msg, Level: "info"}
}

func feedEntries(entries []diff.Entry) <-chan diff.Entry {
	ch := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drainStream(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestGuardStream_LeaderForwardsEntries(t *testing.T) {
	e, _ := New(Options{TTL: 5 * time.Second, RenewEvery: time.Second})
	entries := []diff.Entry{makeEntry("svc", "hello"), makeEntry("svc", "world")}
	in := feedEntries(entries)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out := GuardStream(ctx, in, e, "primary")
	got := drainStream(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestGuardStream_NonLeader_DropsEntries(t *testing.T) {
	e, _ := New(Options{TTL: 5 * time.Second, RenewEvery: time.Second})
	e.Acquire("other") // other holds the lease
	entries := []diff.Entry{makeEntry("svc", "hello")}
	in := feedEntries(entries)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out := GuardStream(ctx, in, e, "replica")
	got := drainStream(out)
	if len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}

func TestGuardStream_StopsOnContextCancel(t *testing.T) {
	e, _ := New(DefaultOptions())
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := GuardStream(ctx, in, e, "primary")
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected closed channel")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for channel close")
	}
}
