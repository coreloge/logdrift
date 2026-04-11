package alert

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
)

func feedResults(src chan<- ResultStream, items []ResultStream) {
	for _, r := range items {
		src <- r
	}
	close(src)
}

func drainAlerts(ch <-chan Alert, timeout time.Duration) []Alert {
	var out []Alert
	deadline := time.After(timeout)
	for {
		select {
		case a, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, a)
		case <-deadline:
			return out
		}
	}
}

func TestWatcher_NoAlerts_WhenAllEqual(t *testing.T) {
	src := make(chan ResultStream, 4)
	feedResults(src, []ResultStream{
		{Service: "svc", Result: diff.Result{Equal: true}},
	})
	w := NewWatcher(context.Background(), src, DefaultConfig())
	alerts := drainAlerts(w.Out, 200*time.Millisecond)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestWatcher_EmitsAlert_OnDrift(t *testing.T) {
	src := make(chan ResultStream, 4)
	feedResults(src, []ResultStream{
		{
			Service: "api",
			Result: diff.Result{
				Equal:  false,
				Deltas: []diff.Delta{{Field: "level", A: "info", B: "error"}},
			},
		},
	})
	w := NewWatcher(context.Background(), src, DefaultConfig())
	alerts := drainAlerts(w.Out, 500*time.Millisecond)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Service != "api" {
		t.Errorf("expected service 'api', got %s", alerts[0].Service)
	}
}

func TestWatcher_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan ResultStream)
	w := NewWatcher(ctx, src, DefaultConfig())
	cancel()
	select {
	case _, ok := <-w.Out:
		if ok {
			t.Error("expected channel to be closed after context cancel")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timed out waiting for watcher to stop")
	}
}
