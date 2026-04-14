package annotate

import (
	"context"
	"testing"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(svc, level, msg string, extra map[string]string) diff.Entry {
	if extra == nil {
		extra = map[string]string{}
	}
	return diff.Entry{Service: svc, Level: level, Message: msg, Extra: extra}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{{Field: "message", Pattern: "[", AnnotationKey: "k", AnnotationValue: "v"}}
	_, err := New(opts)
	if err == nil {
		t.Fatal("expected error for invalid pattern, got nil")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{{Field: "message", Pattern: "error", AnnotationKey: "severity", AnnotationValue: "high"}}
	_, err := New(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_NoRules_PassesThrough(t *testing.T) {
	a, _ := New(DefaultOptions())
	entry := makeEntry("svc", "info", "hello", nil)
	out := a.Apply(entry)
	if out.Message != "hello" || out.Level != "info" {
		t.Errorf("unexpected mutation: %+v", out)
	}
}

func TestApply_AnnotatesOnMessageMatch(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "message", Pattern: "timeout", AnnotationKey: "alert", AnnotationValue: "true"},
	}
	a, _ := New(opts)
	entry := makeEntry("svc", "error", "connection timeout occurred", nil)
	out := a.Apply(entry)
	if out.Extra["annotation.alert"] != "true" {
		t.Errorf("expected annotation.alert=true, got %q", out.Extra["annotation.alert"])
	}
}

func TestApply_AnnotatesOnLevelMatch(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "level", Pattern: "^error$", AnnotationKey: "tier", AnnotationValue: "critical"},
	}
	a, _ := New(opts)
	entry := makeEntry("svc", "error", "boom", nil)
	out := a.Apply(entry)
	if out.Extra["annotation.tier"] != "critical" {
		t.Errorf("expected annotation.tier=critical, got %q", out.Extra["annotation.tier"])
	}
}

func TestApply_NoMatchLeavesExtraUnchanged(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "message", Pattern: "panic", AnnotationKey: "flag", AnnotationValue: "1"},
	}
	a, _ := New(opts)
	entry := makeEntry("svc", "info", "all good", map[string]string{"existing": "yes"})
	out := a.Apply(entry)
	if _, ok := out.Extra["annotation.flag"]; ok {
		t.Error("annotation should not be set when pattern does not match")
	}
	if out.Extra["existing"] != "yes" {
		t.Error("pre-existing extra field should be preserved")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "message", Pattern: ".", AnnotationKey: "touched", AnnotationValue: "yes"},
	}
	a, _ := New(opts)
	origExtra := map[string]string{"k": "v"}
	entry := makeEntry("svc", "info", "msg", origExtra)
	a.Apply(entry)
	if _, ok := origExtra["annotation.touched"]; ok {
		t.Error("Apply must not mutate the original Extra map")
	}
}

func TestStream_AnnotatesAndForwards(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "level", Pattern: "warn", AnnotationKey: "watch", AnnotationValue: "yes"},
	}
	a, _ := New(opts)

	in := make(chan diff.Entry, 2)
	in <- makeEntry("svc", "warn", "low disk", nil)
	in <- makeEntry("svc", "info", "ok", nil)
	close(in)

	ctx := context.Background()
	out := Stream(ctx, a, in)

	first := <-out
	if first.Extra["annotation.watch"] != "yes" {
		t.Errorf("expected annotation on warn entry, got %q", first.Extra["annotation.watch"])
	}
	second := <-out
	if _, ok := second.Extra["annotation.watch"]; ok {
		t.Error("info entry should not have annotation")
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	a, _ := New(DefaultOptions())
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, a, in)
	cancel()
	_, ok := <-out
	if ok {
		t.Error("expected output channel to be closed after context cancel")
	}
}
