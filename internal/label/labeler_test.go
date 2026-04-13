package label_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/label"
)

func makeEntry(service, level, msg string) diff.Entry {
	return diff.Entry{
		Service:   service,
		Level:     level,
		Message:   msg,
		Timestamp: time.Now(),
		Fields:    map[string]string{},
	}
}

func TestApply_NoLabels_PassesThrough(t *testing.T) {
	l := label.New(label.DefaultOptions())
	entry := makeEntry("svc", "info", "hello")
	out := l.Apply(entry)
	if out.Message != entry.Message {
		t.Fatalf("expected message %q, got %q", entry.Message, out.Message)
	}
	if len(out.Fields) != 0 {
		t.Fatalf("expected no fields, got %d", len(out.Fields))
	}
}

func TestApply_AddsLabel(t *testing.T) {
	opts := label.Options{
		Labels:    map[string]string{"env": "prod"},
		Overwrite: false,
	}
	l := label.New(opts)
	out := l.Apply(makeEntry("svc", "info", "msg"))
	if out.Fields["env"] != "prod" {
		t.Fatalf("expected env=prod, got %q", out.Fields["env"])
	}
}

func TestApply_Prefix(t *testing.T) {
	opts := label.Options{
		Labels:    map[string]string{"region": "us-east-1"},
		Prefix:    "meta",
		Overwrite: false,
	}
	l := label.New(opts)
	out := l.Apply(makeEntry("svc", "info", "msg"))
	if out.Fields["meta.region"] != "us-east-1" {
		t.Fatalf("expected meta.region=us-east-1, got %v", out.Fields)
	}
}

func TestApply_NoOverwrite_KeepsExisting(t *testing.T) {
	opts := label.Options{
		Labels:    map[string]string{"env": "staging"},
		Overwrite: false,
	}
	l := label.New(opts)
	entry := makeEntry("svc", "info", "msg")
	entry.Fields["env"] = "prod"
	out := l.Apply(entry)
	if out.Fields["env"] != "prod" {
		t.Fatalf("expected existing value prod, got %q", out.Fields["env"])
	}
}

func TestApply_Overwrite_ReplacesExisting(t *testing.T) {
	opts := label.Options{
		Labels:    map[string]string{"env": "staging"},
		Overwrite: true,
	}
	l := label.New(opts)
	entry := makeEntry("svc", "info", "msg")
	entry.Fields["env"] = "prod"
	out := l.Apply(entry)
	if out.Fields["env"] != "staging" {
		t.Fatalf("expected overwritten value staging, got %q", out.Fields["env"])
	}
}

func TestStream_AttachesLabels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan diff.Entry, 2)
	in <- makeEntry("a", "info", "one")
	in <- makeEntry("b", "warn", "two")
	close(in)

	opts := label.Options{
		Labels:    map[string]string{"team": "platform"},
		Overwrite: false,
	}
	out := label.Stream(ctx, in, opts)

	var results []diff.Entry
	for e := range out {
		results = append(results, e)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	for _, r := range results {
		if r.Fields["team"] != "platform" {
			t.Errorf("missing label on entry %q", r.Message)
		}
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan diff.Entry)
	out := label.Stream(ctx, in, label.DefaultOptions())
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel to be closed after context cancel")
	}
}
