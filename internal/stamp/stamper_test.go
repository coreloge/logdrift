package stamp_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/diff"
	"github.com/logdrift/logdrift/internal/stamp"
)

func makeEntry(svc, level, msg string) diff.Entry {
	return diff.Entry{Service: svc, Level: level, Message: msg}
}

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	opts := stamp.DefaultOptions()
	opts.Field = ""
	_, err := stamp.New(opts)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_EmptyFormat_ReturnsError(t *testing.T) {
	opts := stamp.DefaultOptions()
	opts.Format = ""
	_, err := stamp.New(opts)
	if err == nil {
		t.Fatal("expected error for empty format")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := stamp.New(stamp.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_AddsTimestampField(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	opts := stamp.DefaultOptions()
	opts.ReferenceNow = fixedNow(now)
	s, _ := stamp.New(opts)

	e := makeEntry("svc", "info", "hello")
	out := s.Apply(e)

	got, ok := out.Extra[opts.Field]
	if !ok {
		t.Fatalf("field %q not set", opts.Field)
	}
	want := now.Format(opts.Format)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	s, _ := stamp.New(stamp.DefaultOptions())
	e := makeEntry("svc", "info", "hello")
	s.Apply(e)
	if e.Extra != nil {
		t.Fatal("original entry was mutated")
	}
}

func TestStream_ForwardsStampedEntries(t *testing.T) {
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	opts := stamp.DefaultOptions()
	opts.ReferenceNow = fixedNow(now)
	s, _ := stamp.New(opts)

	in := make(chan diff.Entry, 2)
	in <- makeEntry("a", "info", "one")
	in <- makeEntry("b", "warn", "two")
	close(in)

	ctx := context.Background()
	out := stamp.Stream(ctx, s, in)

	var results []diff.Entry
	for e := range out {
		results = append(results, e)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	for _, r := range results {
		if _, ok := r.Extra[opts.Field]; !ok {
			t.Errorf("entry missing stamp field: %+v", r)
		}
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	s, _ := stamp.New(stamp.DefaultOptions())
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := stamp.Stream(ctx, s, in)
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel to be closed after cancel")
	}
}
