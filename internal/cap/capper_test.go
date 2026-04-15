package cap_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/cap"
	"github.com/user/logdrift/internal/diff"
)

func makeEntry(msg string) diff.Entry {
	return diff.Entry{Service: "svc", Level: "info", Message: msg}
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
	opts := cap.DefaultOptions()
	if opts.MaxEntries != 0 {
		t.Fatalf("expected 0, got %d", opts.MaxEntries)
	}
}

func TestNew_NegativeMax_ReturnsError(t *testing.T) {
	_, err := cap.New(cap.Options{MaxEntries: -1})
	if err == nil {
		t.Fatal("expected error for negative MaxEntries")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := cap.New(cap.Options{MaxEntries: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStream_ZeroCap_ForwardsAll(t *testing.T) {
	c, _ := cap.New(cap.DefaultOptions())
	entries := []diff.Entry{makeEntry("a"), makeEntry("b"), makeEntry("c")}
	out := drain(c.Stream(context.Background(), feedEntries(entries)))
	if len(out) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(out))
	}
}

func TestStream_CapsAtMaxEntries(t *testing.T) {
	c, _ := cap.New(cap.Options{MaxEntries: 2})
	entries := []diff.Entry{makeEntry("a"), makeEntry("b"), makeEntry("c"), makeEntry("d")}
	out := drain(c.Stream(context.Background(), feedEntries(entries)))
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	c, _ := cap.New(cap.Options{MaxEntries: 100})
	ch := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := c.Stream(ctx, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
