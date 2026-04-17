package unwrap

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(extra map[string]any) diff.Entry {
	return diff.Entry{
		Service: "svc",
		Level:   "info",
		Message: "hello",
		Extra:   extra,
	}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := New(Options{})
	if err == nil {
		t.Fatal("expected error for empty Field")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(Options{Field: "meta"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_NoNestedField_PassesThrough(t *testing.T) {
	u, _ := New(Options{Field: "meta"})
	e := makeEntry(map[string]any{"foo": "bar"})
	out := u.Apply(e)
	if out.Extra["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", out.Extra["foo"])
	}
}

func TestApply_UnwrapsNestedMap(t *testing.T) {
	u, _ := New(Options{Field: "meta", Remove: true})
	e := makeEntry(map[string]any{
		"meta": map[string]any{"region": "us-east", "env": "prod"},
	})
	out := u.Apply(e)
	if out.Extra["region"] != "us-east" {
		t.Errorf("expected region=us-east, got %v", out.Extra["region"])
	}
	if _, ok := out.Extra["meta"]; ok {
		t.Error("expected meta to be removed")
	}
}

func TestApply_Prefix(t *testing.T) {
	u, _ := New(Options{Field: "meta", Prefix: "m_", Remove: false})
	e := makeEntry(map[string]any{"meta": map[string]any{"k": "v"}})
	out := u.Apply(e)
	if out.Extra["m_k"] != "v" {
		t.Errorf("expected m_k=v, got %v", out.Extra["m_k"])
	}
	if _, ok := out.Extra["meta"]; !ok {
		t.Error("expected meta to be retained when Remove=false")
	}
}

func TestApply_NoOverwrite_KeepsExisting(t *testing.T) {
	u, _ := New(Options{Field: "meta", Overwrite: false})
	e := makeEntry(map[string]any{
		"region": "eu-west",
		"meta":   map[string]any{"region": "us-east"},
	})
	out := u.Apply(e)
	if out.Extra["region"] != "eu-west" {
		t.Errorf("expected region to remain eu-west, got %v", out.Extra["region"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	u, _ := New(Options{Field: "meta", Remove: true})
	extra := map[string]any{"meta": map[string]any{"k": "v"}}
	e := makeEntry(extra)
	u.Apply(e)
	if _, ok := e.Extra["meta"]; !ok {
		t.Error("original entry was mutated")
	}
}

func TestStream_UnwrapsAndForwards(t *testing.T) {
	u, _ := New(Options{Field: "meta"})
	in := make(chan diff.Entry, 2)
	in <- makeEntry(map[string]any{"meta": map[string]any{"x": 1}})
	in <- makeEntry(map[string]any{"other": "y"})
	close(in)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out := Stream(ctx, u, in)

	var results []diff.Entry
	for e := range out {
		results = append(results, e)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	if results[0].Extra["x"] != 1 {
		t.Errorf("expected x=1, got %v", results[0].Extra["x"])
	}
}
