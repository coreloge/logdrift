package classify

import (
	"regexp"
	"testing"

	"github.com/user/logdrift/internal/diff"
)

func makeEntry(level, message string, fields map[string]string) diff.Entry {
	if fields == nil {
		fields = map[string]string{}
	}
	return diff.Entry{Service: "svc", Level: level, Message: message, Fields: fields}
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

func TestApply_NoRules_NoDefaultCategory_Passthrough(t *testing.T) {
	c, _ := New(DefaultOptions())
	e := makeEntry("info", "hello", nil)
	out := c.Apply(e)
	if _, ok := out.Fields["category"]; ok {
		t.Error("expected no category field when no rules and no default")
	}
}

func TestApply_NoRules_DefaultCategory_Applied(t *testing.T) {
	opts := DefaultOptions()
	opts.DefaultCategory = "general"
	c, _ := New(opts)
	out := c.Apply(makeEntry("info", "hello", nil))
	if got := out.Fields["category"]; got != "general" {
		t.Errorf("expected 'general', got %q", got)
	}
}

func TestApply_MatchesLevelRule(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "level", Pattern: regexp.MustCompile(`^error$`), Category: "errors"},
	}
	c, _ := New(opts)
	out := c.Apply(makeEntry("error", "boom", nil))
	if got := out.Fields["category"]; got != "errors" {
		t.Errorf("expected 'errors', got %q", got)
	}
}

func TestApply_MatchesMessageRule(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "message", Pattern: regexp.MustCompile(`timeout`), Category: "latency"},
	}
	c, _ := New(opts)
	out := c.Apply(makeEntry("warn", "request timeout occurred", nil))
	if got := out.Fields["category"]; got != "latency" {
		t.Errorf("expected 'latency', got %q", got)
	}
}

func TestApply_MatchesCustomFieldRule(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "env", Pattern: regexp.MustCompile(`prod`), Category: "production"},
	}
	c, _ := New(opts)
	out := c.Apply(makeEntry("info", "deploy", map[string]string{"env": "production"}))
	if got := out.Fields["category"]; got != "production" {
		t.Errorf("expected 'production', got %q", got)
	}
}

func TestApply_FirstRuleWins(t *testing.T) {
	opts := DefaultOptions()
	opts.Rules = []Rule{
		{Field: "level", Pattern: regexp.MustCompile(`error`), Category: "errors"},
		{Field: "level", Pattern: regexp.MustCompile(`error`), Category: "should-not-win"},
	}
	c, _ := New(opts)
	out := c.Apply(makeEntry("error", "msg", nil))
	if got := out.Fields["category"]; got != "errors" {
		t.Errorf("expected 'errors', got %q", got)
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	opts := DefaultOptions()
	opts.DefaultCategory = "general"
	c, _ := New(opts)
	original := makeEntry("info", "hello", map[string]string{"x": "1"})
	_ = c.Apply(original)
	if _, ok := original.Fields["category"]; ok {
		t.Error("original entry was mutated")
	}
}
