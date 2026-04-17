package index_test

import (
	"testing"
	"time"

	"github.com/humanlogio/logdrift/internal/diff"
	"github.com/humanlogio/logdrift/internal/index"
)

func makeEntry(field, value, msg string) diff.Entry {
	return diff.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   msg,
		Service:   "svc",
		Extra:     map[string]any{field: value},
	}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := index.New(index.Options{Field: ""})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := index.New(index.Options{Field: "request_id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdd_And_Lookup(t *testing.T) {
	idx, _ := index.New(index.Options{Field: "request_id"})
	e := makeEntry("request_id", "abc-123", "hello")
	idx.Add(e)

	results := idx.Lookup("abc-123")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Message != "hello" {
		t.Errorf("unexpected message: %s", results[0].Message)
	}
}

func TestLookup_MissingKey_ReturnsNil(t *testing.T) {
	idx, _ := index.New(index.Options{Field: "request_id"})
	if idx.Lookup("nope") != nil {
		t.Fatal("expected nil for missing key")
	}
}

func TestAdd_IgnoresEntryWithoutField(t *testing.T) {
	idx, _ := index.New(index.Options{Field: "request_id"})
	e := diff.Entry{Message: "no field", Extra: map[string]any{}}
	idx.Add(e)
	if idx.Lookup("") != nil {
		t.Fatal("expected nil")
	}
}

func TestAdd_RespectsMaxEntries(t *testing.T) {
	idx, _ := index.New(index.Options{Field: "id", MaxEntries: 2})
	for i, v := range []string{"a", "b", "c"} {
		_ = i
		idx.Add(makeEntry("id", v, "msg"))
	}
	// only first two should be indexed
	if idx.Lookup("c") != nil {
		t.Fatal("expected third entry to be dropped")
	}
	if idx.Lookup("a") == nil || idx.Lookup("b") == nil {
		t.Fatal("expected first two entries to be indexed")
	}
}

func TestReset_ClearsIndex(t *testing.T) {
	idx, _ := index.New(index.Options{Field: "id"})
	idx.Add(makeEntry("id", "x", "msg"))
	idx.Reset()
	if idx.Lookup("x") != nil {
		t.Fatal("expected nil after reset")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := index.DefaultOptions()
	if opts.Field == "" {
		t.Error("expected non-empty default field")
	}
	if opts.MaxEntries <= 0 {
		t.Error("expected positive default MaxEntries")
	}
}
