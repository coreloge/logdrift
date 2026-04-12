package transform_test

import (
	"testing"

	"logdrift/internal/diff"
	"logdrift/internal/transform"
)

func makeEntry(service, level, message string, fields map[string]string) diff.Entry {
	if fields == nil {
		fields = map[string]string{}
	}
	return diff.Entry{Service: service, Level: level, Message: message, Fields: fields}
}

func TestApply_NoRules_PassesThrough(t *testing.T) {
	tr := transform.New(transform.DefaultOptions())
	in := makeEntry("svc", "info", "hello world", nil)
	out := tr.Apply(in)
	if out.Message != in.Message || out.Level != in.Level {
		t.Fatalf("expected unchanged entry, got %+v", out)
	}
}

func TestApply_Uppercase_Message(t *testing.T) {
	tr := transform.New(transform.Options{
		Rules: []transform.Rule{{Field: "message", Op: transform.OpUppercase}},
	})
	out := tr.Apply(makeEntry("svc", "info", "hello", nil))
	if out.Message != "HELLO" {
		t.Fatalf("expected HELLO, got %q", out.Message)
	}
}

func TestApply_Lowercase_Level(t *testing.T) {
	tr := transform.New(transform.Options{
		Rules: []transform.Rule{{Field: "level", Op: transform.OpLowercase}},
	})
	out := tr.Apply(makeEntry("svc", "ERROR", "boom", nil))
	if out.Level != "error" {
		t.Fatalf("expected error, got %q", out.Level)
	}
}

func TestApply_Prefix_CustomField(t *testing.T) {
	tr := transform.New(transform.Options{
		Rules: []transform.Rule{{Field: "trace_id", Op: transform.OpPrefix, Arg: "trace-"}},
	})
	in := makeEntry("svc", "info", "msg", map[string]string{"trace_id": "abc123"})
	out := tr.Apply(in)
	if out.Fields["trace_id"] != "trace-abc123" {
		t.Fatalf("expected trace-abc123, got %q", out.Fields["trace_id"])
	}
}

func TestApply_Truncate_Message(t *testing.T) {
	tr := transform.New(transform.Options{
		Rules: []transform.Rule{{Field: "message", Op: transform.OpTruncate, Arg: "5"}},
	})
	out := tr.Apply(makeEntry("svc", "info", "hello world", nil))
	if out.Message != "hello" {
		t.Fatalf("expected hello, got %q", out.Message)
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	tr := transform.New(transform.Options{
		Rules: []transform.Rule{{Field: "message", Op: transform.OpUppercase}},
	})
	in := makeEntry("svc", "info", "original", nil)
	_ = tr.Apply(in)
	if in.Message != "original" {
		t.Fatalf("original entry mutated: got %q", in.Message)
	}
}

func TestApply_Suffix_Field(t *testing.T) {
	tr := transform.New(transform.Options{
		Rules: []transform.Rule{{Field: "env", Op: transform.OpSuffix, Arg: "-v2"}},
	})
	in := makeEntry("svc", "info", "msg", map[string]string{"env": "production"})
	out := tr.Apply(in)
	if out.Fields["env"] != "production-v2" {
		t.Fatalf("expected production-v2, got %q", out.Fields["env"])
	}
}
