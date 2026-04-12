package redact_test

import (
	"regexp"
	"testing"

	"github.com/yourorg/logdrift/internal/redact"
)

func TestApply_NoRules_PassesThrough(t *testing.T) {
	r := redact.New(redact.Options{})
	in := map[string]string{"msg": "hello", "level": "info"}
	out := r.Apply(in)
	if out["msg"] != "hello" || out["level"] != "info" {
		t.Fatalf("expected unchanged fields, got %v", out)
	}
}

func TestApply_RedactsMatchingField(t *testing.T) {
	opts := redact.Options{
		Rules: []redact.Rule{
			{Field: "password", Replacement: "[REDACTED]"},
		},
	}
	r := redact.New(opts)
	out := r.Apply(map[string]string{"password": "s3cr3t", "user": "alice"})
	if out["password"] != "[REDACTED]" {
		t.Fatalf("expected redacted password, got %q", out["password"])
	}
	if out["user"] != "alice" {
		t.Fatalf("expected user unchanged, got %q", out["user"])
	}
}

func TestApply_CaseFold(t *testing.T) {
	opts := redact.Options{
		CaseFold: true,
		Rules: []redact.Rule{
			{Field: "token", Replacement: "***"},
		},
	}
	r := redact.New(opts)
	out := r.Apply(map[string]string{"Token": "abc123"})
	if out["Token"] != "***" {
		t.Fatalf("expected case-folded redaction, got %q", out["Token"])
	}
}

func TestApply_PatternFilter_OnlyMatchingValue(t *testing.T) {
	pat := regexp.MustCompile(`^Bearer `)
	opts := redact.Options{
		Rules: []redact.Rule{
			{Field: "authorization", Pattern: pat, Replacement: "[REDACTED]"},
		},
	}
	r := redact.New(opts)

	out1 := r.Apply(map[string]string{"authorization": "Bearer tok123"})
	if out1["authorization"] != "[REDACTED]" {
		t.Fatalf("expected redaction for Bearer token, got %q", out1["authorization"])
	}

	out2 := r.Apply(map[string]string{"authorization": "Basic dXNlcjpwYXNz"})
	if out2["authorization"] == "[REDACTED]" {
		t.Fatalf("expected Basic auth to pass through")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	opts := redact.DefaultOptions()
	r := redact.New(opts)
	orig := map[string]string{"password": "hunter2"}
	r.Apply(orig)
	if orig["password"] != "hunter2" {
		t.Fatal("original map was mutated")
	}
}

func TestDefaultOptions_ContainsCommonFields(t *testing.T) {
	opts := redact.DefaultOptions()
	fields := map[string]bool{}
	for _, rule := range opts.Rules {
		fields[rule.Field] = true
	}
	for _, expected := range []string{"password", "token", "secret", "api_key", "authorization"} {
		if !fields[expected] {
			t.Errorf("expected default rule for %q", expected)
		}
	}
}
