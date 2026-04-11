package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/logdrift/internal/output"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	w := output.New(nil, "")
	if w == nil {
		t.Fatal("expected non-nil writer")
	}
	if w.Format() != output.FormatText {
		t.Errorf("expected format %q, got %q", output.FormatText, w.Format())
	}
}

func TestNew_CustomFormat(t *testing.T) {
	w := output.New(nil, output.FormatJSON)
	if w.Format() != output.FormatJSON {
		t.Errorf("expected format %q, got %q", output.FormatJSON, w.Format())
	}
}

func TestWriteLine_AppendsNewline(t *testing.T) {
	var buf bytes.Buffer
	w := output.New(&buf, output.FormatText)

	if err := w.WriteLine("hello world"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if got != "hello world\n" {
		t.Errorf("expected %q, got %q", "hello world\n", got)
	}
}

func TestWriteLines_AllLinesWritten(t *testing.T) {
	var buf bytes.Buffer
	w := output.New(&buf, output.FormatText)

	lines := []string{"line one", "line two", "line three"}
	if err := w.WriteLines(lines); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	for _, l := range lines {
		if !strings.Contains(result, l) {
			t.Errorf("expected output to contain %q", l)
		}
	}

	got := strings.Count(result, "\n")
	if got != len(lines) {
		t.Errorf("expected %d newlines, got %d", len(lines), got)
	}
}

func TestWriteLines_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := output.New(&buf, output.FormatText)

	if err := w.WriteLines(nil); err != nil {
		t.Fatalf("unexpected error on empty lines: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty buffer", buf.String())
	}
}

func TestWriteLine_ConcurrentSafe(t *testing.T) {
	var buf bytes.Buffer
	w := output.New(&buf, output.FormatText)

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			_ = w.WriteLine("concurrent line")
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
