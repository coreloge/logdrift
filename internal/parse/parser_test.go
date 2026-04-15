package parse_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/parse"
)

func makeEntry(msg, level, svc string) diff.Entry {
	return diff.Entry{Message: msg, Level: level, Service: svc}
}

func TestNew_EmptySourceField_ReturnsError(t *testing.T) {
	_, err := parse.New(parse.Options{
		Rules: []parse.Rule{{SourceField: "", Pattern: `(?P<x>\d+)`}},
	})
	if err == nil {
		t.Fatal("expected error for empty SourceField")
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := parse.New(parse.Options{
		Rules: []parse.Rule{{SourceField: "message", Pattern: `[invalid`}},
	})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := parse.New(parse.Options{
		Rules: []parse.Rule{{SourceField: "message", Pattern: `(?P<code>\d+)`}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_NoRules_PassesThrough(t *testing.T) {
	p, _ := parse.New(parse.DefaultOptions())
	e := makeEntry("hello world", "info", "svc")
	got := p.Apply(e)
	if got.Message != e.Message {
		t.Errorf("message mutated: got %q", got.Message)
	}
}

func TestApply_ExtractsNamedGroup(t *testing.T) {
	p, _ := parse.New(parse.Options{
		Rules: []parse.Rule{
			{SourceField: "message", Pattern: `status=(?P<status>\d+)`},
		},
	})
	e := makeEntry("request done status=200", "info", "api")
	got := p.Apply(e)
	if got.Extra["status"] != "200" {
		t.Errorf("expected status=200, got %q", got.Extra["status"])
	}
}

func TestApply_DestFieldOverridesGroupName(t *testing.T) {
	p, _ := parse.New(parse.Options{
		Rules: []parse.Rule{
			{SourceField: "message", Pattern: `status=(?P<code>\d+)`, DestField: "http_status"},
		},
	})
	e := makeEntry("done status=404", "warn", "api")
	got := p.Apply(e)
	if got.Extra["http_status"] != "404" {
		t.Errorf("expected http_status=404, got %q", got.Extra["http_status"])
	}
	if _, ok := got.Extra["code"]; ok {
		t.Error("group name should not appear when DestField is set")
	}
}

func TestApply_NoMatch_DoesNotAddField(t *testing.T) {
	p, _ := parse.New(parse.Options{
		Rules: []parse.Rule{
			{SourceField: "message", Pattern: `status=(?P<status>\d+)`},
		},
	})
	e := makeEntry("nothing here", "info", "svc")
	got := p.Apply(e)
	if _, ok := got.Extra["status"]; ok {
		t.Error("status field should not be set when pattern does not match")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	p, _ := parse.New(parse.Options{
		Rules: []parse.Rule{
			{SourceField: "message", Pattern: `id=(?P<id>\w+)`},
		},
	})
	e := makeEntry("req id=abc", "debug", "svc")
	_ = p.Apply(e)
	if e.Extra != nil {
		t.Error("original entry Extra should remain nil")
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

func TestStream_ParsesAndForwards(t *testing.T) {
	p, _ := parse.New(parse.Options{
		Rules: []parse.Rule{
			{SourceField: "message", Pattern: `code=(?P<code>\d+)`},
		},
	})
	entries := []diff.Entry{
		makeEntry("done code=201", "info", "svc"),
		makeEntry("no match here", "info", "svc"),
	}
	ctx := context.Background()
	out := parse.Stream(ctx, p, feedEntries(entries))
	var got []diff.Entry
	for e := range out {
		got = append(got, e)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Extra["code"] != "201" {
		t.Errorf("expected code=201, got %q", got[0].Extra["code"])
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	p, _ := parse.New(parse.DefaultOptions())
	ch := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := parse.Stream(ctx, p, ch)
	cancel()
	select {
	case <-out:
	case <-time.After(time.Second):
		t.Fatal("stream did not stop after context cancel")
	}
}
