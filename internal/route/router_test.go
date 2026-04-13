package route

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(service, level, msg string) diff.Entry {
	return diff.Entry{Service: service, Level: level, Message: msg, Fields: map[string]string{}}
}

func collect(ch <-chan diff.Entry, n int) []diff.Entry {
	var out []diff.Entry
	for i := 0; i < n; i++ {
		select {
		case e := <-ch:
			out = append(out, e)
		case <-time.After(200 * time.Millisecond):
			return out
		}
	}
	return out
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.DefaultRoute != "" || len(opts.Rules) != 0 {
		t.Fatal("expected zero-value options")
	}
}

func TestRouter_MatchesByLevel(t *testing.T) {
	opts := Options{
		Rules: []Rule{{Name: "errors", Field: "level", Values: []string{"error"}}},
	}
	r := New(opts)
	errCh := r.Subscribe("errors")

	in := make(chan diff.Entry, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx, in)

	in <- makeEntry("svc", "error", "boom")
	in <- makeEntry("svc", "info", "ok")
	close(in)

	got := collect(errCh, 2)
	if len(got) != 1 || got[0].Level != "error" {
		t.Fatalf("expected 1 error entry, got %v", got)
	}
}

func TestRouter_DefaultRoute(t *testing.T) {
	opts := Options{
		Rules:        []Rule{{Name: "errors", Field: "level", Values: []string{"error"}}},
		DefaultRoute: "rest",
	}
	r := New(opts)
	r.Subscribe("errors")
	restCh := r.Subscribe("rest")

	in := make(chan diff.Entry, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx, in)

	in <- makeEntry("svc", "info", "hello")
	in <- makeEntry("svc", "warn", "careful")
	close(in)

	got := collect(restCh, 3)
	if len(got) != 2 {
		t.Fatalf("expected 2 rest entries, got %d", len(got))
	}
}

func TestRouter_MatchesByService(t *testing.T) {
	opts := Options{
		Rules: []Rule{{Name: "api", Field: "service", Values: []string{"api-gateway"}}},
	}
	r := New(opts)
	apiCh := r.Subscribe("api")

	in := make(chan diff.Entry, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx, in)

	in <- makeEntry("api-gateway", "info", "req")
	in <- makeEntry("worker", "info", "job")
	close(in)

	got := collect(apiCh, 2)
	if len(got) != 1 || got[0].Service != "api-gateway" {
		t.Fatalf("expected 1 api entry, got %v", got)
	}
}

func TestRouter_StopsOnContextCancel(t *testing.T) {
	r := New(DefaultOptions())
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { r.Run(ctx, in); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not stop after context cancel")
	}
}
