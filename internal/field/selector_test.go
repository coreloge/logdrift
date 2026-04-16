package field_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
	"github.com/logdrift/logdrift/internal/field"
)

func makeEntry(extra map[string]string) diff.Entry {
	return diff.Entry{
		Service:   "svc",
		Level:     "info",
		Message:   "hello",
		Timestamp: time.Now(),
		Extra:     extra,
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

func drain(ch <-chan diff.Entry) []diff.Entry {
	var out []diff.Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestNew_EmptyFieldName_ReturnsError(t *testing.T) {
	_, err := field.New(field.Options{Fields: []string{""}})
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := field.New(field.Options{Fields: []string{"request_id"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_NoFields_PassesThrough(t *testing.T) {
	s, _ := field.New(field.DefaultOptions())
	e := makeEntry(map[string]string{"a": "1", "b": "2"})
	got := s.Apply(e)
	if len(got.Extra) != 2 {
		t.Fatalf("expected 2 extra fields, got %d", len(got.Extra))
	}
}

func TestApply_RetainsSelectedFields(t *testing.T) {
	s, _ := field.New(field.Options{Fields: []string{"request_id"}})
	e := makeEntry(map[string]string{"request_id": "abc", "user_id": "42"})
	got := s.Apply(e)
	if _, ok := got.Extra["request_id"]; !ok {
		t.Error("expected request_id to be retained")
	}
	if _, ok := got.Extra["user_id"]; ok {
		t.Error("expected user_id to be dropped")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	s, _ := field.New(field.Options{Fields: []string{"x"}})
	e := makeEntry(map[string]string{"x": "1", "y": "2"})
	s.Apply(e)
	if len(e.Extra) != 2 {
		t.Error("original entry was mutated")
	}
}

func TestStream_ForwardsProjectedEntries(t *testing.T) {
	s, _ := field.New(field.Options{Fields: []string{"keep"}})
	entries := []diff.Entry{
		makeEntry(map[string]string{"keep": "yes", "drop": "no"}),
		makeEntry(map[string]string{"keep": "also", "drop": "no"}),
	}
	ctx := context.Background()
	out := field.Stream(ctx, feedEntries(entries), s)
	got := drain(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	for _, e := range got {
		if _, ok := e.Extra["drop"]; ok {
			t.Error("drop field should have been removed")
		}
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	s, _ := field.New(field.DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch := make(chan diff.Entry)
	close(ch)
	out := field.Stream(ctx, ch, s)
	drain(out) // must not block
}
