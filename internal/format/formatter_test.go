package format_test

import (
	"strings"
	"testing"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/format"
)

func makeEntry(service, level, message string) diff.Entry {
	return diff.Entry{
		Service: service,
		Level:   level,
		Message: message,
		Fields:  map[string]interface{}{"time": "2024-01-01T00:00:00Z"},
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := format.DefaultOptions()
	if opts.Style != format.StyleText {
		t.Errorf("expected text style, got %s", opts.Style)
	}
	if !opts.ShowService {
		t.Error("expected ShowService to be true")
	}
}

func TestFormat_TextContainsLevelAndMessage(t *testing.T) {
	f := format.New(format.DefaultOptions())
	entry := makeEntry("api", "info", "request received")
	out := f.Format(entry)
	if !strings.Contains(out, "info") {
		t.Errorf("expected level in output, got: %s", out)
	}
	if !strings.Contains(out, "request received") {
		t.Errorf("expected message in output, got: %s", out)
	}
}

func TestFormat_TextContainsService(t *testing.T) {
	f := format.New(format.DefaultOptions())
	out := f.Format(makeEntry("worker", "warn", "slow query"))
	if !strings.Contains(out, "worker") {
		t.Errorf("expected service name in output, got: %s", out)
	}
}

func TestFormat_JSONContainsKeys(t *testing.T) {
	opts := format.DefaultOptions()
	opts.Style = format.StyleJSON
	f := format.New(opts)
	out := f.Format(makeEntry("svc", "error", "boom"))
	for _, key := range []string{"level", "message", "service"} {
		if !strings.Contains(out, key) {
			t.Errorf("expected key %q in JSON output: %s", key, out)
		}
	}
}

func TestFormat_ColorContainsMessage(t *testing.T) {
	opts := format.DefaultOptions()
	opts.Style = format.StyleColor
	f := format.New(opts)
	out := f.Format(makeEntry("svc", "error", "critical failure"))
	if !strings.Contains(out, "critical failure") {
		t.Errorf("expected message in colour output, got: %s", out)
	}
	// ANSI escape code present
	if !strings.Contains(out, "\033[") {
		t.Errorf("expected ANSI codes in colour output, got: %s", out)
	}
}

func TestFormat_HideService(t *testing.T) {
	opts := format.DefaultOptions()
	opts.ShowService = false
	f := format.New(opts)
	out := f.Format(makeEntry("hidden", "info", "msg"))
	if strings.Contains(out, "hidden") {
		t.Errorf("expected service to be hidden, got: %s", out)
	}
}

func TestFormat_EmptyStyleDefaultsToText(t *testing.T) {
	opts := format.Options{Style: "", ShowService: true, TimestampField: "time"}
	f := format.New(opts)
	out := f.Format(makeEntry("svc", "debug", "hello"))
	if !strings.Contains(out, "hello") {
		t.Errorf("expected message in output, got: %s", out)
	}
}
