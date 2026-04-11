package metrics_test

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/metrics"
)

func TestNew_InitialisesEmpty(t *testing.T) {
	c := metrics.New()
	if len(c.Entries()) != 0 {
		t.Fatalf("expected empty entries, got %v", c.Entries())
	}
	if len(c.Drifts()) != 0 {
		t.Fatalf("expected empty drifts, got %v", c.Drifts())
	}
}

func TestRecordEntry_IncrementsCount(t *testing.T) {
	c := metrics.New()
	c.RecordEntry("svc-a")
	c.RecordEntry("svc-a")
	c.RecordEntry("svc-b")

	if c.Entries()["svc-a"] != 2 {
		t.Errorf("expected 2 entries for svc-a, got %d", c.Entries()["svc-a"])
	}
	if c.Entries()["svc-b"] != 1 {
		t.Errorf("expected 1 entry for svc-b, got %d", c.Entries()["svc-b"])
	}
}

func TestRecordDrift_IncrementsCount(t *testing.T) {
	c := metrics.New()
	c.RecordDrift("svc-a")
	c.RecordDrift("svc-a")

	if c.Drifts()["svc-a"] != 2 {
		t.Errorf("expected 2 drifts for svc-a, got %d", c.Drifts()["svc-a"])
	}
	if c.Drifts()["svc-b"] != 0 {
		t.Errorf("expected 0 drifts for svc-b, got %d", c.Drifts()["svc-b"])
	}
}

func TestEntries_ReturnsCopy(t *testing.T) {
	c := metrics.New()
	c.RecordEntry("svc-a")
	snap := c.Entries()
	snap["svc-a"] = 999
	if c.Entries()["svc-a"] != 1 {
		t.Error("Entries should return an independent copy")
	}
}

func TestReset_ZeroesCounters(t *testing.T) {
	c := metrics.New()
	c.RecordEntry("svc-a")
	c.RecordDrift("svc-a")
	c.Reset()

	if len(c.Entries()) != 0 {
		t.Errorf("expected empty entries after reset, got %v", c.Entries())
	}
	if len(c.Drifts()) != 0 {
		t.Errorf("expected empty drifts after reset, got %v", c.Drifts())
	}
}

func TestUptime_IsPositive(t *testing.T) {
	c := metrics.New()
	time.Sleep(time.Millisecond)
	if c.Uptime() <= 0 {
		t.Error("expected positive uptime")
	}
}

func TestConcurrentAccess_DoesNotRace(t *testing.T) {
	c := metrics.New()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			c.RecordEntry("svc-a")
			c.RecordDrift("svc-a")
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		_ = c.Entries()
		_ = c.Drifts()
	}
	<-done
}
