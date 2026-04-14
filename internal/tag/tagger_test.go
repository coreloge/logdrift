package tag

import (
	"context"
	"regexp"
	"testing"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(svc, level, msg string) diff.LogEntry {
	return diff.LogEntry{
		Service: svc,
		Level:   level,
		Message: msg,
		Extra:   map[string]interface{}{},
	}
}

func TestNew_EmptyOutputField_ReturnsError(t *testing.T) {
	_, err := New(Options{OutputField: ""})
	if err == nil {
		t.Fatal("expected error for empty OutputField")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_NoRules_PassesThrough(t *testing.T) {
	tgr, _ := New(DefaultOptions())
	entry := makeEntry("svc", "info", "hello")
	out := tgr.Apply(entry)
	if out.Message != "hello" {
		t.Errorf("unexpected message: %s", out.Message)
	}
	if _, ok := out.Extra["tag"]; ok {
		t.Error("expected no tag field when no rules match")
	}
}

func TestApply_TagsOnMessageMatch(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "message", Pattern: regexp.MustCompile(`error`), Tag: "needs-attention"},
	}
	tgr, _ := New(opts)

	out := tgr.Apply(makeEntry("svc", "error", "disk error occurred"))
	if out.Extra["tag"] != "needs-attention" {
		t.Errorf("expected tag 'needs-attention', got %v", out.Extra["tag"])
	}
}

func TestApply_FirstRuleWins(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "level", Pattern: regexp.MustCompile(`error`), Tag: "first"},
		{Field: "level", Pattern: regexp.MustCompile(`error`), Tag: "second"},
	}
	tgr, _ := New(opts)

	out := tgr.Apply(makeEntry("svc", "error", "msg"))
	if out.Extra["tag"] != "first" {
		t.Errorf("expected 'first', got %v", out.Extra["tag"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "level", Pattern: regexp.MustCompile(`warn`), Tag: "warning"},
	}
	tgr, _ := New(opts)
	entry := makeEntry("svc", "warn", "low disk")
	_ = tgr.Apply(entry)
	if _, ok := entry.Extra["tag"]; ok {
		t.Error("original entry must not be mutated")
	}
}

func TestStream_TagsAndForwards(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "service", Pattern: regexp.MustCompile(`auth`), Tag: "auth-svc"},
	}
	tgr, _ := New(opts)

	in := make(chan diff.LogEntry, 2)
	in <- makeEntry("auth", "info", "login")
	in <- makeEntry("db", "info", "query")
	close(in)

	ctx := context.Background()
	out := Stream(ctx, tgr, in)

	first := <-out
	if first.Extra["tag"] != "auth-svc" {
		t.Errorf("expected 'auth-svc', got %v", first.Extra["tag"])
	}
	second := <-out
	if _, ok := second.Extra["tag"]; ok {
		t.Error("db entry should not be tagged")
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	tgr, _ := New(DefaultOptions())
	in := make(chan diff.LogEntry)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, tgr, in)
	cancel()
	_, ok := <-out
	if ok {
		t.Error("expected channel to be closed after context cancel")
	}
}
