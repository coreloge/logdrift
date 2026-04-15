package fingerprint_test

import (
	"testing"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/fingerprint"
)

func makeEntry(svc, level, msg string, extra map[string]string) diff.Entry {
	return diff.Entry{
		Service: svc,
		Level:   level,
		Message: msg,
		Extra:   extra,
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := fingerprint.DefaultOptions()
	if !opts.IncludeService {
		t.Error("expected IncludeService to be true")
	}
	if !opts.IncludeLevel {
		t.Error("expected IncludeLevel to be true")
	}
	if opts.IncludeExtra {
		t.Error("expected IncludeExtra to be false")
	}
}

func TestCompute_SameEntry_SameFingerprint(t *testing.T) {
	fp := fingerprint.New(fingerprint.DefaultOptions())
	e := makeEntry("api", "error", "connection refused", nil)
	if got1, got2 := fp.Compute(e), fp.Compute(e); got1 != got2 {
		t.Errorf("expected identical fingerprints, got %q and %q", got1, got2)
	}
}

func TestCompute_DifferentMessage_DifferentFingerprint(t *testing.T) {
	fp := fingerprint.New(fingerprint.DefaultOptions())
	a := makeEntry("api", "error", "timeout", nil)
	b := makeEntry("api", "error", "connection refused", nil)
	if fp.Compute(a) == fp.Compute(b) {
		t.Error("expected different fingerprints for different messages")
	}
}

func TestCompute_ExcludeService_SameAcrossServices(t *testing.T) {
	opts := fingerprint.Options{IncludeService: false, IncludeLevel: true, IncludeExtra: false}
	fp := fingerprint.New(opts)
	a := makeEntry("api", "warn", "slow query", nil)
	b := makeEntry("db", "warn", "slow query", nil)
	if fp.Compute(a) != fp.Compute(b) {
		t.Error("expected same fingerprint when service excluded")
	}
}

func TestCompute_IncludeExtra_OrderIndependent(t *testing.T) {
	opts := fingerprint.Options{IncludeService: false, IncludeLevel: false, IncludeExtra: true}
	fp := fingerprint.New(opts)
	a := makeEntry("svc", "info", "hello", map[string]string{"x": "1", "y": "2"})
	b := makeEntry("svc", "info", "hello", map[string]string{"y": "2", "x": "1"})
	if fp.Compute(a) != fp.Compute(b) {
		t.Error("expected same fingerprint regardless of extra field insertion order")
	}
}

func TestCompute_IncludeExtra_DifferentValues_DifferentFingerprint(t *testing.T) {
	opts := fingerprint.Options{IncludeService: false, IncludeLevel: false, IncludeExtra: true}
	fp := fingerprint.New(opts)
	a := makeEntry("svc", "info", "hello", map[string]string{"x": "1"})
	b := makeEntry("svc", "info", "hello", map[string]string{"x": "2"})
	if fp.Compute(a) == fp.Compute(b) {
		t.Error("expected different fingerprints for different extra values")
	}
}

func TestCompute_ReturnsHexString(t *testing.T) {
	fp := fingerprint.New(fingerprint.DefaultOptions())
	e := makeEntry("svc", "info", "msg", nil)
	got := fp.Compute(e)
	if len(got) != 64 {
		t.Errorf("expected 64-char hex string, got len=%d: %q", len(got), got)
	}
}
