package sequence

import (
	"context"
	"strconv"
	"testing"

	"github.com/yourorg/logdrift/internal/diff"
)

func makeEntry(svc, msg string) diff.Entry {
	return diff.Entry{Service: svc, Message: msg, Level: "info", Extra: map[string]string{}}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Field == "" {
		t.Fatal("expected non-empty default field")
	}
	if opts.PerService {
		t.Fatal("expected PerService to default to false")
	}
}

func TestApply_GlobalCounter(t *testing.T) {
	s, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for i := 1; i <= 3; i++ {
		out := s.Apply(makeEntry("svcA", "msg"))
		got := out.Extra[DefaultOptions().Field]
		if got != strconv.Itoa(i) {
			t.Fatalf("step %d: want %d got %s", i, i, got)
		}
	}
}

func TestApply_PerService(t *testing.T) {
	s, _ := New(Options{Field: "_seq", PerService: true})
	s.Apply(makeEntry("a", "x"))
	s.Apply(makeEntry("b", "x"))
	out := s.Apply(makeEntry("a", "x"))
	if out.Extra["_seq"] != "2" {
		t.Fatalf("want 2 got %s", out.Extra["_seq"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	s, _ := New(DefaultOptions())
	e := makeEntry("svc", "hello")
	s.Apply(e)
	if _, ok := e.Extra[DefaultOptions().Field]; ok {
		t.Fatal("original entry was mutated")
	}
}

func TestReset_ZeroesCounters(t *testing.T) {
	s, _ := New(DefaultOptions())
	s.Apply(makeEntry("svc", "a"))
	s.Apply(makeEntry("svc", "b"))
	s.Reset()
	out := s.Apply(makeEntry("svc", "c"))
	if out.Extra[DefaultOptions().Field] != "1" {
		t.Fatalf("expected 1 after reset, got %s", out.Extra[DefaultOptions().Field])
	}
}

func TestStream_StampsEntries(t *testing.T) {
	s, _ := New(DefaultOptions())
	in := make(chan diff.Entry, 3)
	in <- makeEntry("svc", "one")
	in <- makeEntry("svc", "two")
	in <- makeEntry("svc", "three")
	close(in)

	ctx := context.Background()
	out := Stream(ctx, s, in)

	for i := 1; i <= 3; i++ {
		e := <-out
		got := e.Extra[DefaultOptions().Field]
		if got != strconv.Itoa(i) {
			t.Fatalf("entry %d: want %d got %s", i, i, got)
		}
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	s, _ := New(DefaultOptions())
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, s, in)
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel to be closed after cancel")
	}
}
