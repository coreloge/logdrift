package proxy

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
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

func drainStream(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestNew_NilHook_ReturnsError(t *testing.T) {
	_, err := New(Options{Hook: nil})
	if err == nil {
		t.Fatal("expected error for nil hook")
	}
}

func TestNew_ValidHook_NoError(t *testing.T) {
	_, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_CallsHook(t *testing.T) {
	var got []diff.Entry
	p, _ := New(Options{Hook: func(e diff.Entry) { got = append(got, e) }})

	e := makeEntry("hello")
	out := p.Apply(e)

	if len(got) != 1 || got[0].Message != "hello" {
		t.Fatalf("hook not called correctly: %v", got)
	}
	if out.Message != e.Message {
		t.Fatalf("entry mutated: got %q", out.Message)
	}
}

func TestApply_IncrementsSeen(t *testing.T) {
	p, _ := New(DefaultOptions())
	p.Apply(makeEntry("a"))
	p.Apply(makeEntry("b"))
	if p.Seen() != 2 {
		t.Fatalf("expected 2 seen, got %d", p.Seen())
	}
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	entries := []diff.Entry{makeEntry("x"), makeEntry("y"), makeEntry("z")}
	p, _ := New(DefaultOptions())
	ctx := context.Background()

	out := Stream(ctx, p, feedEntries(entries))
	result := drainStream(out)

	if len(result) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(result))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := make(chan diff.Entry)
	p, _ := New(DefaultOptions())
	out := Stream(ctx, p, in)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected closed channel")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for stream to stop")
	}
}
