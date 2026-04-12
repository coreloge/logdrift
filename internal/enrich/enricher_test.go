package enrich_test

import (
	"testing"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/enrich"
)

func makeEntry(service, msg string, extra map[string]string) diff.Entry {
	return diff.Entry{
		Service: service,
		Message: msg,
		Level:   "info",
		Extra:   extra,
	}
}

func TestApply_NoFields_PassesThrough(t *testing.T) {
	e := enrich.New(enrich.DefaultOptions())
	in := makeEntry("svc", "hello", nil)
	out := e.Apply(in)
	if out.Message != in.Message {
		t.Fatalf("expected message %q, got %q", in.Message, out.Message)
	}
}

func TestApply_AddsStaticField(t *testing.T) {
	opts := enrich.DefaultOptions()
	opts.StaticFields = map[string]string{"env": "prod"}
	e := enrich.New(opts)
	out := e.Apply(makeEntry("svc", "msg", nil))
	if out.Extra["env"] != "prod" {
		t.Fatalf("expected env=prod, got %q", out.Extra["env"])
	}
}

func TestApply_DoesNotOverwriteByDefault(t *testing.T) {
	opts := enrich.DefaultOptions()
	opts.StaticFields = map[string]string{"env": "prod"}
	e := enrich.New(opts)
	in := makeEntry("svc", "msg", map[string]string{"env": "staging"})
	out := e.Apply(in)
	if out.Extra["env"] != "staging" {
		t.Fatalf("expected env=staging, got %q", out.Extra["env"])
	}
}

func TestApply_OverwriteExisting(t *testing.T) {
	opts := enrich.Options{StaticFields: map[string]string{"env": "prod"}, OverwriteExisting: true}
	e := enrich.New(opts)
	in := makeEntry("svc", "msg", map[string]string{"env": "staging"})
	out := e.Apply(in)
	if out.Extra["env"] != "prod" {
		t.Fatalf("expected env=prod, got %q", out.Extra["env"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	opts := enrich.DefaultOptions()
	opts.StaticFields = map[string]string{"region": "us-east"}
	e := enrich.New(opts)
	in := makeEntry("svc", "msg", map[string]string{"k": "v"})
	e.Apply(in)
	if _, ok := in.Extra["region"]; ok {
		t.Fatal("Apply mutated the original entry")
	}
}

func TestSetField_AddsAtRuntime(t *testing.T) {
	e := enrich.New(enrich.DefaultOptions())
	e.SetField("dc", "eu-west")
	out := e.Apply(makeEntry("svc", "msg", nil))
	if out.Extra["dc"] != "eu-west" {
		t.Fatalf("expected dc=eu-west, got %q", out.Extra["dc"])
	}
}

func TestRemoveField_RemovesAtRuntime(t *testing.T) {
	opts := enrich.DefaultOptions()
	opts.StaticFields = map[string]string{"dc": "eu-west"}
	e := enrich.New(opts)
	e.RemoveField("dc")
	out := e.Apply(makeEntry("svc", "msg", nil))
	if _, ok := out.Extra["dc"]; ok {
		t.Fatal("expected dc to be removed")
	}
}
