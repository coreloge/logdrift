package mask_test

import (
	"context"
	"testing"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/mask"
)

func makeEntry(svc, level, msg string, fields map[string]string) diff.Entry {
	if fields == nil {
		fields = map[string]string{}
	}
	return diff.Entry{Service: svc, Level: level, Message: msg, Fields: fields}
}

func TestApply_NoFields_PassesThrough(t *testing.T) {
	m := mask.New(mask.DefaultOptions())
	e := makeEntry("svc", "info", "hello", map[string]string{"token": "secret"})
	out := m.Apply(e)
	if out.Message != "hello" || out.Fields["token"] != "secret" {
		t.Fatal("expected entry unchanged")
	}
}

func TestApply_MasksNamedField(t *testing.T) {
	opts := mask.Options{Fields: []string{"token"}, Placeholder: "***"}
	m := mask.New(opts)
	e := makeEntry("svc", "info", "msg", map[string]string{"token": "abc123", "user": "alice"})
	out := m.Apply(e)
	if out.Fields["token"] != "***" {
		t.Fatalf("expected token masked, got %q", out.Fields["token"])
	}
	if out.Fields["user"] != "alice" {
		t.Fatal("expected user unchanged")
	}
}

func TestApply_MasksMessage(t *testing.T) {
	opts := mask.Options{Fields: []string{"message"}, Placeholder: "[hidden]"}
	m := mask.New(opts)
	e := makeEntry("svc", "warn", "sensitive text", nil)
	out := m.Apply(e)
	if out.Message != "[hidden]" {
		t.Fatalf("expected message masked, got %q", out.Message)
	}
	if out.Level != "warn" {
		t.Fatal("expected level unchanged")
	}
}

func TestApply_CaseInsensitiveField(t *testing.T) {
	opts := mask.Options{Fields: []string{"Authorization"}, Placeholder: "***"}
	m := mask.New(opts)
	e := makeEntry("svc", "debug", "req", map[string]string{"authorization": "Bearer xyz"})
	out := m.Apply(e)
	if out.Fields["authorization"] != "***" {
		t.Fatalf("expected authorization masked, got %q", out.Fields["authorization"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	opts := mask.Options{Fields: []string{"secret"}, Placeholder: "***"}
	m := mask.New(opts)
	e := makeEntry("svc", "info", "msg", map[string]string{"secret": "val"})
	_ = m.Apply(e)
	if e.Fields["secret"] != "val" {
		t.Fatal("original entry was mutated")
	}
}

func TestStream_MasksEntries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan diff.Entry, 2)
	in <- makeEntry("a", "info", "hello", map[string]string{"pw": "pass"})
	in <- makeEntry("b", "error", "fail", map[string]string{"pw": "word"})
	close(in)

	opts := mask.Options{Fields: []string{"pw"}, Placeholder: "***"}
	out := mask.Stream(ctx, in, opts)

	var results []diff.Entry
	for e := range out {
		results = append(results, e)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	for _, r := range results {
		if r.Fields["pw"] != "***" {
			t.Fatalf("expected pw masked, got %q", r.Fields["pw"])
		}
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan diff.Entry)
	out := mask.Stream(ctx, in, mask.DefaultOptions())
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel closed after context cancel")
	}
}
