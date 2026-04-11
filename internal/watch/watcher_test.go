package watch

import (
	"context"
	"os"
	"testing"
	"time"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "logdrift-watch-*.log")
	if err != nil {
		t.Fatalf("createTemp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestNew_InitialisesFields(t *testing.T) {
	w := New([]string{"/tmp/a.log"}, 50*time.Millisecond)
	if w.Events == nil {
		t.Fatal("expected Events channel to be non-nil")
	}
	if w.interval != 50*time.Millisecond {
		t.Fatalf("unexpected interval: %v", w.interval)
	}
}

func TestWatcher_DetectsTruncation(t *testing.T) {
	path := writeTempFile(t, "line1\nline2\n")
	w := New([]string{path}, 20*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	w.Start(ctx)
	// seed inode/size state with one poll
	time.Sleep(30 * time.Millisecond)
	// truncate the file to simulate rotation
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	select {
	case ev := <-w.Events:
		if ev.Path != path {
			t.Fatalf("expected path %q, got %q", path, ev.Path)
		}
		if !ev.Rotated {
			t.Fatal("expected Rotated=true")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for rotation event")
	}
}

func TestWatcher_NoEventForStableFile(t *testing.T) {
	path := writeTempFile(t, "stable content\n")
	w := New([]string{path}, 20*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	w.Start(ctx)
	select {
	case ev := <-w.Events:
		t.Fatalf("unexpected event for stable file: %+v", ev)
	case <-ctx.Done():
		// expected: no events fired
	}
}
