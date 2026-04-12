package rotate_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/user/logdrift/internal/rotate"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "rotate-*.log")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

func TestNew_MissingFile(t *testing.T) {
	ctx := context.Background()
	_, err := rotate.New(ctx, "/nonexistent/path/file.log", rotate.DefaultOptions())
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestNew_ValidFile(t *testing.T) {
	path := writeTempFile(t, "initial content\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := rotate.New(ctx, path, rotate.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil rotator")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := rotate.DefaultOptions()
	if opts.PollInterval <= 0 {
		t.Errorf("expected positive PollInterval, got %v", opts.PollInterval)
	}
	if opts.MaxReopens <= 0 {
		t.Errorf("expected positive MaxReopens, got %v", opts.MaxReopens)
	}
}

func TestRotator_EmitsEventOnTruncation(t *testing.T) {
	path := writeTempFile(t, "line one\n")

	opts := rotate.DefaultOptions()
	opts.PollInterval = 30 * time.Millisecond
	opts.MaxReopens = 5

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	r, err := rotate.New(ctx, path, opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Truncate the file to simulate rotation.
	time.Sleep(60 * time.Millisecond)
	if err := os.Truncate(path, 0); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	select {
	case ev, ok := <-r.Events:
		if !ok {
			t.Fatal("events channel closed before receiving event")
		}
		if ev.Path != path {
			t.Errorf("path: got %q, want %q", ev.Path, path)
		}
		if ev.Reopens < 1 {
			t.Errorf("reopens: got %d, want >= 1", ev.Reopens)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for rotation event")
	}
}

func TestRotator_StopsOnContextCancel(t *testing.T) {
	path := writeTempFile(t, "data\n")

	opts := rotate.DefaultOptions()
	opts.PollInterval = 20 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	r, err := rotate.New(ctx, path, opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cancel()

	// Events channel should close shortly after cancel.
	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case _, ok := <-r.Events:
			if !ok {
				return // channel closed as expected
			}
		case <-deadline:
			t.Fatal("events channel not closed after context cancel")
		}
	}
}
