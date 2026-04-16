package tee

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(svc, level, msg string) diff.Entry {
	return diff.Entry{
		Service: svc,
		Level:   level,
		Message: msg,
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

func drainStream(ch <-chan diff.Entry, timeout time.Duration) []diff.Entry {
	var out []diff.Entry
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case e, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, e)
		case <-timer.C:
			return out
		}
	}
}

func TestNew_NilWriter_DefaultsToDiscard(t *testing.T) {
	tt, err := New(Options{Writer: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tt.opts.Writer != io.Discard {
		t.Error("expected io.Discard for nil writer")
	}
}

func TestApply_WritesToSideChannel(t *testing.T) {
	var buf bytes.Buffer
	tt, _ := New(Options{Writer: &buf, Format: "text"})
	e := makeEntry("api", "error", "something broke")
	out := tt.Apply(e)
	if out.Service != e.Service || out.Message != e.Message {
		t.Error("Apply must return entry unchanged")
	}
	if !strings.Contains(buf.String(), "something broke") {
		t.Errorf("side-channel output missing message, got: %q", buf.String())
	}
}

func TestApply_DoesNotMutateEntry(t *testing.T) {
	var buf bytes.Buffer
	tt, _ := New(Options{Writer: &buf})
	e := makeEntry("svc", "info", "hello")
	out := tt.Apply(e)
	if out.Level != e.Level {
		t.Error("Apply must not change Level")
	}
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	tt, _ := New(DefaultOptions())
	entries := []diff.Entry{
		makeEntry("a", "info", "one"),
		makeEntry("b", "warn", "two"),
	}
	in := feedEntries(entries)
	out := Stream(context.Background(), in, tt)
	got := drainStream(out, time.Second)
	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	tt, _ := New(DefaultOptions())
	ch := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, ch, tt)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for stream to stop")
	}
}
