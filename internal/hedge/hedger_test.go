package hedge_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/hedge"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{Service: svc, Message: msg, Level: "info"}
}

func feedChan(entries []diff.Entry) <-chan diff.Entry {
	ch := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drain(ctx context.Context, ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for {
		select {
		case e, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, e)
		case <-ctx.Done():
			return out
		}
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := hedge.DefaultOptions()
	if opts.MaxSources != 0 {
		t.Fatalf("expected MaxSources 0, got %d", opts.MaxSources)
	}
}

func TestHedger_ReceivesAllEntries(t *testing.T) {
	src1 := feedChan([]diff.Entry{makeEntry("svc-a", "hello")})
	src2 := feedChan([]diff.Entry{makeEntry("svc-b", "world")})

	h := hedge.New(hedge.DefaultOptions(), src1, src2)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	entries := drain(ctx, h.Stream(ctx))
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestHedger_MaxSources_Limits(t *testing.T) {
	src1 := feedChan([]diff.Entry{makeEntry("svc-a", "one")})
	src2 := feedChan([]diff.Entry{makeEntry("svc-b", "two")})
	src3 := feedChan([]diff.Entry{makeEntry("svc-c", "three")})

	opts := hedge.Options{MaxSources: 2}
	h := hedge.New(opts, src1, src2, src3)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	entries := drain(ctx, h.Stream(ctx))
	// Only 2 sources are used; expect exactly 2 entries.
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries with MaxSources=2, got %d", len(entries))
	}
}

func TestHedger_StopsOnContextCancel(t *testing.T) {
	blocking := make(chan diff.Entry) // never sends

	h := hedge.New(hedge.DefaultOptions(), blocking)
	ctx, cancel := context.WithCancel(context.Background())
	out := h.Stream(ctx)
	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close after cancel")
	}
}

func TestHedger_EmptySources(t *testing.T) {
	h := hedge.New(hedge.DefaultOptions())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	entries := drain(ctx, h.Stream(ctx))
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries from empty hedger, got %d", len(entries))
	}
}
