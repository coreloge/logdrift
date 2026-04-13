package flatten_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/flatten"
)

func makeEntry(extra map[string]string) diff.Entry {
	return diff.Entry{
		Service: "svc",
		Level:   "info",
		Message: "test message",
		Extra:   extra,
	}
}

func TestNew_EmptySeparator_ReturnsError(t *testing.T) {
	_, err := flatten.New(flatten.Options{Separator: "", MaxDepth: 4})
	if err == nil {
		t.Fatal("expected error for empty separator")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := flatten.New(flatten.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_NoExtra_PassesThrough(t *testing.T) {
	f, _ := flatten.New(flatten.DefaultOptions())
	e := makeEntry(nil)
	out := f.Apply(e)
	if out.Service != e.Service || out.Message != e.Message {
		t.Errorf("core fields mutated unexpectedly")
	}
}

func TestApply_FlatExtra_Unchanged(t *testing.T) {
	f, _ := flatten.New(flatten.DefaultOptions())
	e := makeEntry(map[string]string{"host": "localhost", "env": "prod"})
	out := f.Apply(e)
	if out.Extra["host"] != "localhost" {
		t.Errorf("expected host=localhost, got %q", out.Extra["host"])
	}
	if out.Extra["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", out.Extra["env"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	f, _ := flatten.New(flatten.DefaultOptions())
	orig := map[string]string{"key": "val"}
	e := makeEntry(orig)
	f.Apply(e)
	if len(orig) != 1 {
		t.Error("original extra map was mutated")
	}
}

func TestApply_CustomSeparator(t *testing.T) {
	f, _ := flatten.New(flatten.Options{Separator: "_", MaxDepth: 4})
	e := makeEntry(map[string]string{"region": "us-east"})
	out := f.Apply(e)
	if out.Extra["region"] != "us-east" {
		t.Errorf("unexpected value: %q", out.Extra["region"])
	}
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	f, _ := flatten.New(flatten.DefaultOptions())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := make(chan diff.Entry, 3)
	for i := 0; i < 3; i++ {
		in <- makeEntry(map[string]string{"idx": string(rune('0' + i))})
	}
	close(in)

	out := flatten.Stream(ctx, in, f)
	var got []diff.Entry
	for e := range out {
		got = append(got, e)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	f, _ := flatten.New(flatten.DefaultOptions())
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan diff.Entry) // never sends
	out := flatten.Stream(ctx, in, f)
	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for stream to stop")
	}
}
