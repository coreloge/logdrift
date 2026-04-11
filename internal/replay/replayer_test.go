package replay_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/replay"
)

func writeTempLog(t *testing.T, lines []string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "replay-*.log")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
	f.Close()
	return f.Name()
}

func TestNew_MissingFile(t *testing.T) {
	r := replay.New("/no/such/file.log", replay.DefaultOptions())
	_, err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestReplayer_EmitsEntries(t *testing.T) {
	lines := []string{
		`{"level":"info","msg":"started","service":"svc"}`,
		`{"level":"warn","msg":"slow query","service":"svc"}`,
	}
	path := writeTempLog(t, lines)

	opts := replay.Options{DelayPerLine: 0, Service: "test-svc"}
	r := replay.New(path, opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []string
	for e := range ch {
		got = append(got, e.Message)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestReplayer_ServiceLabel(t *testing.T) {
	lines := []string{`{"level":"info","msg":"hello"}`}
	path := writeTempLog(t, lines)

	opts := replay.Options{DelayPerLine: 0, Service: "my-service"}
	r := replay.New(path, opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, ok := <-ch
	if !ok {
		t.Fatal("channel closed before receiving entry")
	}
	if entry.Service != "my-service" {
		t.Errorf("expected service %q, got %q", "my-service", entry.Service)
	}
}

func TestReplayer_CancelStopsEarly(t *testing.T) {
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, `{"level":"info","msg":"line"}`)
	}
	path := writeTempLog(t, lines)

	opts := replay.Options{DelayPerLine: 5 * time.Millisecond, Service: "svc"}
	r := replay.New(path, opts)

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// read a couple then cancel
	<-ch
	<-ch
	cancel()

	// drain; channel must close within a reasonable time
	done := make(chan struct{})
	go func() { for range ch {}; close(done) }()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("replayer did not stop after context cancel")
	}
}
