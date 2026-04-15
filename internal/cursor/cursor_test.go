package cursor

import (
	"testing"
)

func TestNew_StartsEmpty(t *testing.T) {
	s := New()
	if len(s.Keys()) != 0 {
		t.Fatalf("expected empty store, got %d keys", len(s.Keys()))
	}
}

func TestSet_And_Get(t *testing.T) {
	s := New()
	s.Set("svc", Position{Offset: 42, Line: 7})
	p, err := s.Get("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Offset != 42 || p.Line != 7 {
		t.Fatalf("got %+v, want {42 7}", p)
	}
}

func TestGet_MissingKey_ReturnsErrNotFound(t *testing.T) {
	s := New()
	_, err := s.Get("missing")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	s := New()
	s.Set("svc", Position{Offset: 1})
	s.Delete("svc")
	_, err := s.Get("svc")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestAdvance_AccumulatesDeltas(t *testing.T) {
	s := New()
	s.Advance("svc", 10, 1)
	s.Advance("svc", 20, 1)
	p, err := s.Get("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Offset != 30 || p.Line != 2 {
		t.Fatalf("got %+v, want {30 2}", p)
	}
}

func TestKeys_ReturnsAllKeys(t *testing.T) {
	s := New()
	s.Set("a", Position{})
	s.Set("b", Position{})
	keys := s.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}
