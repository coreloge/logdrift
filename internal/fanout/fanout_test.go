package fanout_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/internal/diff"
	"github.com/logdrift/internal/fanout"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{Service: svc, Message: msg, Level: "info"}
}

func collect(ch <-chan diff.Entry, n int) []diff.Entry {
	var out []diff.Entry
	for i := 0; i < n; i++ {
		select {
		case e, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, e)
			case <-time.After(time.Second):
			return out
		}
	}
	return out
}

func TestFanout_SingleSubscriber(t *testing.T) {
	f := fanout.New(0)
	sub := f.Subscribe(8)

	src := make(chan diff.Entry, 2)
	src <- makeEntry("svc-a", "hello")
	src <- makeEntry("svc-b", "world")
	close(src)

	ctx := context.Background()
	go f.Run(ctx, src)

	got := collect(sub, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Service != "svc-a" {
		t.Errorf("expected svc-a, got %s", got[0].Service)
	}
}

func TestFanout_MultipleSubscribers_ReceiveAll(t *testing.T) {
	f := fanout.New(0)
	sub1 := f.Subscribe(8)
	sub2 := f.Subscribe(8)

	src := make(chan diff.Entry, 3)
	for i := 0; i < 3; i++ {
		src <- makeEntry("svc", "msg")
	}
	close(src)

	ctx := context.Background()
	go f.Run(ctx, src)

	if got := collect(sub1, 3); len(got) != 3 {
		t.Errorf("sub1: expected 3, got %d", len(got))
	}
	if got := collect(sub2, 3); len(got) != 3 {
		t.Errorf("sub2: expected 3, got %d", len(got))
	}
}

func TestFanout_StopsOnContextCancel(t *testing.T) {
	f := fanout.New(0)
	sub := f.Subscribe(8)

	src := make(chan diff.Entry) // never sends
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		f.Run(ctx, src)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after context cancellation")
	}

	// subscriber channel should be closed
	select {
	case _, ok := <-sub:
		if ok {
			t.Error("expected subscriber channel to be closed")
		}
	case <-time.After(time.Second):
		t.Error("subscriber channel was not closed")
	}
}

func TestFanout_SourceClosedWithNoSubscribers(t *testing.T) {
	// Ensure Run completes cleanly when there are no subscribers and the
	// source channel is closed immediately.
	f := fanout.New(0)

	src := make(chan diff.Entry)
	close(src)

	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		f.Run(ctx, src)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after source channel was closed")
	}
}
