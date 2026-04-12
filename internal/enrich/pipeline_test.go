package enrich_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/enrich"
)

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

func TestStream_EnrichesEntries(t *testing.T) {
	opts := enrich.DefaultOptions()
	opts.StaticFields = map[string]string{"env": "test"}
	e := enrich.New(opts)

	in := feedEntries([]diff.Entry{
		{Service: "a", Message: "one", Level: "info"},
		{Service: "b", Message: "two", Level: "warn"},
	})

	out := drainStream(enrich.Stream(context.Background(), e, in))
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	for _, entry := range out {
		if entry.Extra["env"] != "test" {
			t.Errorf("entry %q missing env field", entry.Message)
		}
	}
}

func TestStream_PassesThroughUnaffectedEntries(t *testing.T) {
	e := enrich.New(enrich.DefaultOptions())
	in := feedEntries([]diff.Entry{
		{Service: "svc", Message: "hello", Level: "info"},
	})
	out := drainStream(enrich.Stream(context.Background(), e, in))
	if len(out) != 1 || out[0].Message != "hello" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	e := enrich.New(enrich.DefaultOptions())
	blocking := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := enrich.Stream(ctx, e, blocking)
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
