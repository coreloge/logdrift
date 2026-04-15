package clone_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/clone"
	"github.com/logdrift/logdrift/internal/diff"
)

func makeEntry(svc, level, msg string) diff.Entry {
	return diff.Entry{
		Service:   svc,
		Level:     level,
		Message:   msg,
		Timestamp: time.Now(),
		Extra:     map[string]string{"key": "value"},
	}
}

func TestEntry_CopiesScalarFields(t *testing.T) {
	orig := makeEntry("svc", "info", "hello")
	cp := clone.Entry(orig)

	if cp.Service != orig.Service || cp.Level != orig.Level || cp.Message != orig.Message {
		t.Fatal("scalar fields not copied correctly")
	}
	if !cp.Timestamp.Equal(orig.Timestamp) {
		t.Fatal("timestamp not copied correctly")
	}
}

func TestEntry_ExtraIsDeepCopied(t *testing.T) {
	orig := makeEntry("svc", "info", "hello")
	cp := clone.Entry(orig)

	cp.Extra["key"] = "mutated"
	if orig.Extra["key"] == "mutated" {
		t.Fatal("mutation of clone affected original")
	}
}

func TestEntry_NilExtra_IsNilInCopy(t *testing.T) {
	orig := diff.Entry{Service: "svc", Level: "info", Message: "hi"}
	cp := clone.Entry(orig)
	if cp.Extra != nil {
		t.Fatalf("expected nil Extra, got %v", cp.Extra)
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

func drainStream(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	entries := []diff.Entry{
		makeEntry("a", "info", "one"),
		makeEntry("b", "warn", "two"),
	}
	ctx := context.Background()
	out := clone.Stream(ctx, feedEntries(entries))
	got := drainStream(out)

	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch := make(chan diff.Entry)
	out := clone.Stream(ctx, ch)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("stream did not stop after context cancellation")
	}
}
