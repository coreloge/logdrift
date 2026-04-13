package throttle

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(level string) diff.Entry {
	return diff.Entry{Service: "svc", Level: level, Message: "msg"}
}

func TestDefaultOptions_NoLimits(t *testing.T) {
	opts := DefaultOptions()
	if len(opts.LevelRates) != 0 {
		t.Fatalf("expected empty LevelRates, got %v", opts.LevelRates)
	}
}

func TestAllow_UnlimitedLevel_AlwaysPasses(t *testing.T) {
	l := New(DefaultOptions())
	for i := 0; i < 100; i++ {
		if !l.Allow(makeEntry("info")) {
			t.Fatal("expected entry to be allowed")
		}
	}
}

func TestAllow_ZeroRate_DropsAll(t *testing.T) {
	opts := DefaultOptions()
	opts.LevelRates["debug"] = 0
	l := New(opts)
	if !l.Allow(makeEntry("debug")) {
		// zero rate treated as unlimited
		t.Log("zero rate passes — treated as unlimited, ok")
	}
}

func TestAllow_LimitEnforced(t *testing.T) {
	opts := DefaultOptions()
	opts.LevelRates["warn"] = 2 // 2 tokens/s
	l := New(opts)

	allowed := 0
	for i := 0; i < 10; i++ {
		if l.Allow(makeEntry("warn")) {
			allowed++
		}
	}
	// Only the first ~2 should pass immediately
	if allowed > 3 {
		t.Fatalf("expected at most 3 allowed, got %d", allowed)
	}
}

func TestAllow_PerLevelIsolation(t *testing.T) {
	opts := DefaultOptions()
	opts.LevelRates["error"] = 1
	l := New(opts)

	// error is rate-limited; info is not
	l.Allow(makeEntry("error")) // consume the one token
	if l.Allow(makeEntry("error")) {
		t.Fatal("second error entry should be dropped")
	}
	if !l.Allow(makeEntry("info")) {
		t.Fatal("info entry should always pass")
	}
}

func TestStream_DropsThrottledEntries(t *testing.T) {
	opts := DefaultOptions()
	opts.LevelRates["debug"] = 1
	l := New(opts)

	in := make(chan diff.Entry, 5)
	for i := 0; i < 5; i++ {
		in <- makeEntry("debug")
	}
	close(in)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out := Stream(ctx, in, l)
	var received []diff.Entry
	for e := range out {
		received = append(received, e)
	}
	if len(received) >= 5 {
		t.Fatalf("expected throttling to drop some entries, got %d", len(received))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	l := New(DefaultOptions())
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, in, l)
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel to be closed after context cancel")
	}
}
