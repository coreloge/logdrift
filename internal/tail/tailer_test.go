package tail_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/tail"
)

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "logdrift-*.log")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestNew(t *testing.T) {
	tr := tail.New("svc-a", "/var/log/svc-a.log")
	if tr.Service != "svc-a" {
		t.Errorf("expected service svc-a, got %s", tr.Service)
	}
	if tr.Lines == nil {
		t.Error("expected Lines channel to be initialised")
	}
}

func TestTail_MissingFile(t *testing.T) {
	tr := tail.New("svc", "/nonexistent/path/file.log")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := tr.Tail(ctx)
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestTail_ReceivesNewLines(t *testing.T) {
	path := writeTempLog(t, "") // start empty
	tr := tail.New("svc", path)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = tr.Tail(ctx)
	}()

	// Give the tailer a moment to seek to end.
	time.Sleep(150 * time.Millisecond)

	// Append a line to the file.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open file for append: %v", err)
	}
	if _, err := f.WriteString("hello logdrift\n"); err != nil {
		t.Fatalf("write line: %v", err)
	}
	f.Close()

	select {
	case line := <-tr.Lines:
		if line.Service != "svc" {
			t.Errorf("expected service svc, got %s", line.Service)
		}
		if line.Text != "hello logdrift\n" {
			t.Errorf("unexpected line text: %q", line.Text)
		}
		if line.Time.IsZero() {
			t.Error("expected non-zero timestamp")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for tailed line")
	}
}
