package quota

import (
	"testing"
	"time"
)

func TestDefaultOptions_NoLimit(t *testing.T) {
	opts := DefaultOptions()
	if opts.Max != 0 {
		t.Fatalf("expected Max=0, got %d", opts.Max)
	}
}

func TestNew_NegativeMax_ReturnsError(t *testing.T) {
	_, err := New(Options{Max: -1, Window: time.Minute})
	if err == nil {
		t.Fatal("expected error for negative Max")
	}
}

func TestNew_ZeroWindow_ReturnsError(t *testing.T) {
	_, err := New(Options{Max: 10, Window: 0})
	if err == nil {
		t.Fatal("expected error for zero Window")
	}
}

func TestAllow_ZeroMax_AlwaysTrue(t *testing.T) {
	q, _ := New(Options{Max: 0, Window: time.Minute})
	for i := 0; i < 1000; i++ {
		if !q.Allow("svc") {
			t.Fatal("expected Allow=true for unlimited quota")
		}
	}
}

func TestAllow_LimitEnforced(t *testing.T) {
	q, _ := New(Options{Max: 3, Window: time.Minute})
	for i := 0; i < 3; i++ {
		if !q.Allow("svc") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
	if q.Allow("svc") {
		t.Fatal("expected Allow=false after limit exceeded")
	}
}

func TestAllow_PerServiceIsolation(t *testing.T) {
	q, _ := New(Options{Max: 1, Window: time.Minute})
	if !q.Allow("a") {
		t.Fatal("expected Allow=true for service a")
	}
	if !q.Allow("b") {
		t.Fatal("expected Allow=true for service b (separate bucket)")
	}
	if q.Allow("a") {
		t.Fatal("expected Allow=false for service a after limit")
	}
}

func TestAllow_WindowResets(t *testing.T) {
	q, _ := New(Options{Max: 1, Window: 50 * time.Millisecond})
	if !q.Allow("svc") {
		t.Fatal("expected Allow=true on first call")
	}
	if q.Allow("svc") {
		t.Fatal("expected Allow=false within window")
	}
	time.Sleep(60 * time.Millisecond)
	if !q.Allow("svc") {
		t.Fatal("expected Allow=true after window reset")
	}
}

func TestReset_ClearsCounters(t *testing.T) {
	q, _ := New(Options{Max: 1, Window: time.Minute})
	q.Allow("svc")
	if q.Allow("svc") {
		t.Fatal("expected Allow=false before reset")
	}
	q.Reset()
	if !q.Allow("svc") {
		t.Fatal("expected Allow=true after Reset")
	}
}
