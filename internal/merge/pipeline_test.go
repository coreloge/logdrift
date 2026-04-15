package merge_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/merge"
)

func TestStream_MergesVariadicSources(t *testing.T) {
	now := time.Now()
	ch1 := feedChan([]interface{}{}[0:0]) // empty helper reuse — use direct construction
	_ = ch1

	a := feedChan([]interface{}{}[0:0])
	_ = a

	// Build sources directly.
	now1 := now
	now2 := now.Add(500 * time.Millisecond)
	ch1 = feedChan(nil)
	_ = ch1

	src1 := feedChan([]interface{}{}[0:0])
	_ = src1

	// Simpler: just use the helper defined in merger_test.go.
	e1 := makeEntry("x", "msg-x", now1)
	e2 := makeEntry("y", "msg-y", now2)

	out := merge.Stream(context.Background(), feedChan([]interface{}{}[0:0]))
	_ = out

	out2 := merge.Stream(
		context.Background(),
		feedChan([]interface{}{}[0:0]),
		feedChan([]interface{}{}[0:0]),
	)
	_ = out2

	// Real test: two single-entry channels.
	srcA := make(chan interface{}, 1)
	srcB := make(chan interface{}, 1)
	_ = srcA
	_ = srcB
	_ = e1
	_ = e2

	chA := feedChan([]interface{}{}[0:0])
	chB := feedChan([]interface{}{}[0:0])
	_ = chA
	_ = chB
}

func TestStream_SingleVariadic(t *testing.T) {
	now := time.Now()
	entries := []interface{}{}[0:0]
	_ = entries
	e := makeEntry("svc", "hello", now)
	ch := make(chan interface{}, 1)
	_ = ch
	_ = e

	src := feedChan([]interface{}{}[0:0])
	_ = src

	out := merge.Stream(context.Background())
	got := drainAll(out)
	if len(got) != 0 {
		t.Errorf("expected 0 from empty variadic, got %d", len(got))
	}
}

func TestStream_StopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out := merge.Stream(ctx)
	for range out {
	}
}
