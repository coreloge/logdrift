package stream

import (
	"testing"
	"time"
)

func drain(t *testing.T, ch <-chan Entry, timeout time.Duration) []Entry {
	t.Helper()
	var entries []Entry
	timer := time.After(timeout)
	for {
		select {
		case e, ok := <-ch:
			if !ok {
				return entries
			}
			entries = append(entries, e)
		case <-timer:
			return entries
		}
	}
}

func feedChan(entries []Entry) chan Entry {
	ch := make(chan Entry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func TestFilter_NoOptions_PassesAll(t *testing.T) {
	input := []Entry{
		{Service: "svc-a", Line: `{"level":"info","msg":"ok"}` },
		{Service: "svc-b", Line: `{"level":"error","msg":"fail"}`},
	}
	stop := make(chan struct{})
	out := Filter(feedChan(input), FilterOptions{}, stop)
	got := drain(t, out, time.Second)
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
}

func TestFilter_LevelFilter(t *testing.T) {
	input := []Entry{
		{Service: "svc", Line: `{"level":"info","msg":"ok"}`},
		{Service: "svc", Line: `{"level":"error","msg":"fail"}`},
	}
	stop := make(chan struct{})
	opts := FilterOptions{Levels: []string{"error"}}
	out := Filter(feedChan(input), opts, stop)
	got := drain(t, out, time.Second)
	if len(got) != 1 {
		t.Errorf("expected 1 entry, got %d", len(got))
	}
	if got[0].Line != input[1].Line {
		t.Errorf("unexpected entry: %s", got[0].Line)
	}
}

func TestFilter_ServiceFilter(t *testing.T) {
	input := []Entry{
		{Service: "svc-a", Line: `{"level":"info","msg":"a"}`},
		{Service: "svc-b", Line: `{"level":"info","msg":"b"}`},
	}
	stop := make(chan struct{})
	opts := FilterOptions{Services: []string{"svc-b"}}
	out := Filter(feedChan(input), opts, stop)
	got := drain(t, out, time.Second)
	if len(got) != 1 || got[0].Service != "svc-b" {
		t.Errorf("expected only svc-b, got %v", got)
	}
}

func TestFilter_Stop(t *testing.T) {
	ch := make(chan Entry)
	stop := make(chan struct{})
	out := Filter(ch, FilterOptions{}, stop)
	close(stop)
	select {
	case <-out:
	case <-time.After(time.Second):
		t.Fatal("filter did not close output after stop")
	}
}
