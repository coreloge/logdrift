package quota

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

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

func makeEntry(svc string) diff.Entry {
	return diff.Entry{Service: svc, Message: "msg", Level: "info"}
}

func TestStream_ForwardsAllEntries_WhenUnderQuota(t *testing.T) {
	q, _ := New(Options{Max: 10, Window: time.Minute})
	in := feedEntries([]diff.Entry{makeEntry("a"), makeEntry("b"), makeEntry("a")})
	out := drainStream(Stream(context.Background(), q, in))
	if len(out) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(out))
	}
}

func TestStream_DropsEntriesOverQuota(t *testing.T) {
	q, _ := New(Options{Max: 2, Window: time.Minute})
	entries := []diff.Entry{makeEntry("svc"), makeEntry("svc"), makeEntry("svc"), makeEntry("svc")}
	in := feedEntries(entries)
	out := drainStream(Stream(context.Background(), q, in))
	if len(out) != 2 {
		t.Fatalf("expected 2 entries after quota, got %d", len(out))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	q, _ := New(Options{Max: 100, Window: time.Minute})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := make(chan diff.Entry)
	close(in)
	out := Stream(ctx, q, in)
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected closed channel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
