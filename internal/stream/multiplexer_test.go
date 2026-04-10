package stream

import (
	"os"
	"testing"
	"time"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "logdrift-*.log")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestNew_ValidSources(t *testing.T) {
	p := writeTempFile(t, "")
	sources := map[string]string{"svc-a": p}
	m, err := New(sources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil multiplexer")
	}
}

func TestNew_MissingFile(t *testing.T) {
	sources := map[string]string{"svc-a": "/nonexistent/path.log"}
	_, err := New(sources)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestMultiplexer_ReceivesEntries(t *testing.T) {
	pA := writeTempFile(t, "{\"level\":\"info\",\"msg\":\"hello\"}\n")
	pB := writeTempFile(t, "{\"level\":\"warn\",\"msg\":\"world\"}\n")

	sources := map[string]string{"svc-a": pA, "svc-b": pB}
	m, err := New(sources)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	m.Start()

	seen := map[string]bool{}
	timeout := time.After(2 * time.Second)
	for len(seen) < 2 {
		select {
		case entry := <-m.Out():
			seen[entry.Service] = true
		case <-timeout:
			t.Fatalf("timed out waiting for entries; got services: %v", seen)
		}
	}

	m.Stop()

	if !seen["svc-a"] || !seen["svc-b"] {
		t.Errorf("missing entries; seen: %v", seen)
	}
}

func TestMultiplexer_Stop(t *testing.T) {
	p := writeTempFile(t, "")
	m, err := New(map[string]string{"svc": p})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.Start()
	done := make(chan struct{})
	go func() {
		m.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not return in time")
	}
}
