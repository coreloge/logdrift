package ceiling_test

import (
	"context"
	"testing"
	"time"

	"logdrift/internal/ceiling"
	"logdrift/internal/diff"
)

func makeEntry(service string) diff.Entry {
	return diff.Entry{
		Service:   service,
		Level:     "info",
		Message:   "test message",
		Timestamp: time.Now(),
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

func drain(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestDefaultOptions_NoLimit(t *testing.T) {
	opts := ceiling.DefaultOptions()
	if opts.MaxPerWindow != 0 {
		t.Fatalf("expected MaxPerWindow=0, got %d", opts.MaxPerWindow)
	}
}

func TestNew_NegativeWindow_ReturnsError(t *testing.T) {
	_, err := ceiling.New(ceiling.Options{Window: -time.Second, MaxPerWindow: 5})
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}

func TestNew_ZeroWindow_ReturnsError(t *testing.T) {
	_, err := ceiling.New(ceiling.Options{Window: 0, MaxPerWindow: 5})
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestAllow_ZeroMax_AlwaysPasses(t *testing.T) {
	c, err := ceiling.New(ceiling.Options{Window: time.Minute, MaxPerWindow: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 100; i++ {
		if !c.Allow(makeEntry("svc")) {
			t.Fatal("expected Allow=true for unlimited ceiling")
		}
	}
}

func TestAllow_LimitEnforced(t *testing.T) {
	c, err := ceiling.New(ceiling.Options{Window: time.Minute, MaxPerWindow: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 3; i++ {
		if !c.Allow(makeEntry("svc")) {
			t.Fatalf("entry %d should be allowed", i)
		}
	}
	if c.Allow(makeEntry("svc")) {
		t.Fatal("4th entry should be blocked")
	}
}

func TestAllow_PerServiceIsolation(t *testing.T) {
	c, err := ceiling.New(ceiling.Options{Window: time.Minute, MaxPerWindow: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// fill up svcA
	c.Allow(makeEntry("svcA"))
	c.Allow(makeEntry("svcA"))
	// svcB should still pass
	if !c.Allow(makeEntry("svcB")) {
		t.Fatal("svcB should not be affected by svcA ceiling")
	}
}

func TestStream_ForwardsAllEntries_WhenUnderCeiling(t *testing.T) {
	c, _ := ceiling.New(ceiling.Options{Window: time.Minute, MaxPerWindow: 10})
	entries := []diff.Entry{makeEntry("a"), makeEntry("a"), makeEntry("b")}
	out := ceiling.Stream(context.Background(), feedEntries(entries), c)
	got := drain(out)
	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
}

func TestStream_DropsEntriesOverCeiling(t *testing.T) {
	c, _ := ceiling.New(ceiling.Options{Window: time.Minute, MaxPerWindow: 2})
	entries := []diff.Entry{makeEntry("svc"), makeEntry("svc"), makeEntry("svc"), makeEntry("svc")}
	out := ceiling.Stream(context.Background(), feedEntries(entries), c)
	got := drain(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries after ceiling, got %d", len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c, _ := ceiling.New(ceiling.Options{Window: time.Minute, MaxPerWindow: 10})
	in := make(chan diff.Entry)
	close(in)
	out := ceiling.Stream(ctx, in, c)
	drain(out) // should not block
}
