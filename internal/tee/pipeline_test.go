package tee

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func TestPipe_WritesAndForwards(t *testing.T) {
	var buf bytes.Buffer
	entries := []diff.Entry{
		{Service: "x", Level: "info", Message: "alpha"},
		{Service: "y", Level: "error", Message: "beta"},
	}
	in := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		in <- e
	}
	close(in)

	out := Pipe(context.Background(), in, &buf, "text")
	var got []diff.Entry
	for e := range out {
		got = append(got, e)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 forwarded entries, got %d", len(got))
	}
	if !strings.Contains(buf.String(), "alpha") {
		t.Error("side-channel missing first message")
	}
	if !strings.Contains(buf.String(), "beta") {
		t.Error("side-channel missing second message")
	}
}

func TestPipe_CancelStopsEarly(t *testing.T) {
	var buf bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan diff.Entry)
	out := Pipe(ctx, in, &buf, "text")
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected closed channel")
		}
	case <-time.After(time.Second):
		t.Error("timed out")
	}
}
