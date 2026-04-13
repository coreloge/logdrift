package join_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
	"github.com/logdrift/logdrift/internal/join"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{Service: svc, Message: msg, Level: "info"}
}

func feedSource(entries []diff.Entry) <-chan diff.Entry {
	ch := make(chan diff.Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func drainAll(ch <-chan diff.Entry, timeout time.Duration) []diff.Entry {
	var out []diff.Entry
	timer := time.After(timeout)
	for {
		select {
		case e, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, e)
		case <-timer:
			return out
		}
	}
}

func TestJoiner_MergesAllSources(t *testing.T) {
	sources := []join.Source{
		{Label: "svcA", Ch: feedSource([]diff.Entry{makeEntry("", "hello"), makeEntry("", "world")})},
		{Label: "svcB", Ch: feedSource([]diff.Entry{makeEntry("", "foo")})},
	}

	j := join.New(sources, join.DefaultOptions())
	ctx := context.Background()
	out := j.Stream(ctx)

	results := drainAll(out, 2*time.Second)
	if len(results) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(results))
	}
}

func TestJoiner_TagsServiceLabel(t *testing.T) {
	sources := []join.Source{
		{Label: "alpha", Ch: feedSource([]diff.Entry{makeEntry("", "msg1")})},
		{Label: "beta", Ch: feedSource([]diff.Entry{makeEntry("", "msg2")})},
	}

	j := join.New(sources, join.DefaultOptions())
	out := j.Stream(context.Background())
	results := drainAll(out, 2*time.Second)

	services := make([]string, len(results))
	for i, e := range results {
		services[i] = e.Service
	}
	sort.Strings(services)

	if services[0] != "alpha" || services[1] != "beta" {
		t.Fatalf("unexpected service labels: %v", services)
	}
}

func TestJoiner_PreservesExistingServiceLabel(t *testing.T) {
	entry := makeEntry("original", "keep me")
	sources := []join.Source{
		{Label: "override", Ch: feedSource([]diff.Entry{entry})},
	}

	j := join.New(sources, join.DefaultOptions())
	out := j.Stream(context.Background())
	results := drainAll(out, time.Second)

	if len(results) != 1 || results[0].Service != "original" {
		t.Fatalf("expected service 'original', got %q", results[0].Service)
	}
}

func TestJoiner_StopsOnContextCancel(t *testing.T) {
	blocking := make(chan diff.Entry)
	sources := []join.Source{
		{Label: "slow", Ch: blocking},
	}

	j := join.New(sources, join.DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	out := j.Stream(ctx)
	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to close after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
