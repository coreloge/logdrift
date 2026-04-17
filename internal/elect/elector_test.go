package elect

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.TTL <= 0 {
		t.Fatal("expected positive TTL")
	}
}

func TestNew_ZeroTTL_ReturnsError(t *testing.T) {
	_, err := New(Options{TTL: 0, RenewEvery: time.Second})
	if err == nil {
		t.Fatal("expected error for zero TTL")
	}
}

func TestNew_ZeroRenewEvery_ReturnsError(t *testing.T) {
	_, err := New(Options{TTL: time.Second, RenewEvery: 0})
	if err == nil {
		t.Fatal("expected error for zero RenewEvery")
	}
}

func TestAcquire_FirstCandidate_Wins(t *testing.T) {
	e, _ := New(DefaultOptions())
	if !e.Acquire("a") {
		t.Fatal("expected a to win")
	}
	leader, ok := e.Leader()
	if !ok || leader != "a" {
		t.Fatalf("expected leader=a, got %q ok=%v", leader, ok)
	}
}

func TestAcquire_SecondCandidate_Loses(t *testing.T) {
	e, _ := New(DefaultOptions())
	e.Acquire("a")
	if e.Acquire("b") {
		t.Fatal("b should not acquire while a holds lease")
	}
}

func TestAcquire_LeaseExpiry_AllowsNewLeader(t *testing.T) {
	base := time.Now()
	e, _ := New(Options{TTL: time.Second, RenewEvery: time.Millisecond})
	e.now = fixedClock(base)
	e.Acquire("a")
	e.now = fixedClock(base.Add(2 * time.Second))
	if !e.Acquire("b") {
		t.Fatal("b should win after lease expiry")
	}
}

func TestRevoke_ClearsLeader(t *testing.T) {
	e, _ := New(DefaultOptions())
	e.Acquire("a")
	e.Revoke()
	_, ok := e.Leader()
	if ok {
		t.Fatal("expected no leader after revoke")
	}
}

func TestLeader_NoAcquire_ReturnsFalse(t *testing.T) {
	e, _ := New(DefaultOptions())
	_, ok := e.Leader()
	if ok {
		t.Fatal("expected no leader on fresh elector")
	}
}
