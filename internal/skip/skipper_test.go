package skip_test

import (
	"context"
	"testing"
	"time"

	"logdrift/internal/diff"
	"logdrift/internal/skip"
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

func drain(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestDefaultOptions(t *testing.T) {
	opts := skip.DefaultOptions()
	if opts.Count != 0 {
		t.Fatalf("expected Count=0, got %d", opts.Count)
	}
}

func TestNew_NegativeCount_ReturnsError(t *testing.T) {
	_, err := skip.New(skip.Options{Count: -1})
	if err == nil {
		t.Fatal("expected error for negative Count")
	}
}

func TestNew_ZeroCount_NoError(t *testing.T) {
	_, err := skip.New(skip.Options{Count: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_ZeroCount_PassesAll(t *testing.T) {
	sk, _ := skip.New(skip.Options{Count: 0})
	for i := 0; i < 5; i++ {
		if !sk.Allow(makeEntry("svc", "msg")) {
			t.Fatal("expected all entries to pass with Count=0")
		}
	}
}

func TestAllow_SkipsFirstN(t *testing.T) {
	sk, _ := skip.New(skip.Options{Count: 3})
	results := make([]bool, 6)
	for i := range results {
		results[i] = sk.Allow(makeEntry("svc", "msg"))
	}
	expected := []bool{false, false, false, true, true, true}
	for i, v := range expected {
		if results[i] != v {
			t.Fatalf("index %d: expected %v got %v", i, v, results[i])
		}
	}
}

func TestAllow_PerService_IndependentCounters(t *testing.T) {
	sk, _ := skip.New(skip.Options{Count: 2, PerService: true})
	// First two of svcA should be dropped.
	if sk.Allow(makeEntry("svcA", "1")) {
		t.Fatal("svcA entry 1 should be skipped")
	}
	// svcB counter is independent; first entry of svcB should also be dropped.
	if sk.Allow(makeEntry("svcB", "1")) {
		t.Fatal("svcB entry 1 should be skipped")
	}
	if sk.Allow(makeEntry("svcA", "2")) {
		t.Fatal("svcA entry 2 should be skipped")
	}
	// Now svcA counter is exhausted.
	if !sk.Allow(makeEntry("svcA", "3")) {
		t.Fatal("svcA entry 3 should pass")
	}
}

func TestStream_SkipsFirstN(t *testing.T) {
	entries := []diff.Entry{
		makeEntry("svc", "a"),
		makeEntry("svc", "b"),
		makeEntry("svc", "c"),
		makeEntry("svc", "d"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := skip.Stream(ctx, feedEntries(entries), skip.Options{Count: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := drain(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Message != "c" || got[1].Message != "d" {
		t.Fatalf("unexpected messages: %v", got)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan diff.Entry)
	out, err := skip.Stream(ctx, in, skip.Options{Count: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
