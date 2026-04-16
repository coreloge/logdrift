package distinct_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/distinct"
)

func makeEntry(svc, level, msg string) diff.Entry {
	return diff.Entry{Service: svc, Level: level, Message: msg}
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

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := distinct.New(distinct.Options{Field: ""})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := distinct.New(distinct.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_FirstOccurrence_ReturnsTrue(t *testing.T) {
	d, _ := distinct.New(distinct.DefaultOptions())
	e := makeEntry("svc", "info", "hello")
	if !d.Allow(e) {
		t.Fatal("expected first occurrence to be allowed")
	}
}

func TestAllow_SecondOccurrence_ReturnsFalse(t *testing.T) {
	d, _ := distinct.New(distinct.DefaultOptions())
	e := makeEntry("svc", "info", "hello")
	d.Allow(e)
	if d.Allow(e) {
		t.Fatal("expected second occurrence to be dropped")
	}
}

func TestAllow_DifferentValues_BothAllowed(t *testing.T) {
	d, _ := distinct.New(distinct.DefaultOptions())
	if !d.Allow(makeEntry("svc", "info", "hello")) {
		t.Fatal("expected first message to pass")
	}
	if !d.Allow(makeEntry("svc", "info", "world")) {
		t.Fatal("expected second distinct message to pass")
	}
}

func TestReset_ClearsState(t *testing.T) {
	d, _ := distinct.New(distinct.DefaultOptions())
	e := makeEntry("svc", "info", "hello")
	d.Allow(e)
	d.Reset()
	if !d.Allow(e) {
		t.Fatal("expected entry to pass after reset")
	}
}

func TestStream_ForwardsDistinctOnly(t *testing.T) {
	d, _ := distinct.New(distinct.DefaultOptions())
	entries := []diff.Entry{
		makeEntry("a", "info", "msg1"),
		makeEntry("b", "info", "msg1"), // duplicate message
		makeEntry("c", "info", "msg2"),
	}
	ctx := context.Background()
	out := distinct.Stream(ctx, feedEntries(entries), d)
	result := drain(out)
	if len(result) != 2 {
		t.Fatalf("expected 2 distinct entries, got %d", len(result))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	d, _ := distinct.New(distinct.DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch := make(chan diff.Entry)
	close(ch)
	outCtx, outCancel := context.WithTimeout(context.Background(), time.Second)
	defer outCancel()
	out := distinct.Stream(ctx, ch, d)
	select {
	case <-out:
	case <-outCtx.Done():
		t.Fatal("stream did not close after context cancel")
	}
}
