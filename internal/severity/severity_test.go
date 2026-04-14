package severity_test

import (
	"testing"

	"github.com/user/logdrift/internal/severity"
)

func TestParse_KnownLevels(t *testing.T) {
	cases := []struct {
		input string
		want  severity.Level
	}{
		{"debug", severity.Debug},
		{"TRACE", severity.Debug},
		{"info", severity.Info},
		{"NOTICE", severity.Info},
		{"warn", severity.Warn},
		{"WARNING", severity.Warn},
		{"error", severity.Error},
		{"ERR", severity.Error},
		{"fatal", severity.Fatal},
		{"PANIC", severity.Fatal},
		{"CRITICAL", severity.Fatal},
	}
	for _, tc := range cases {
		got := severity.Parse(tc.input)
		if got != tc.want {
			t.Errorf("Parse(%q) = %v; want %v", tc.input, got, tc.want)
		}
	}
}

func TestParse_Unknown(t *testing.T) {
	if got := severity.Parse("bogus"); got != severity.Unknown {
		t.Fatalf("expected Unknown, got %v", got)
	}
}

func TestLevel_String(t *testing.T) {
	if s := severity.Info.String(); s != "info" {
		t.Fatalf("expected \"info\", got %q", s)
	}
	if s := severity.Unknown.String(); s != "unknown" {
		t.Fatalf("expected \"unknown\", got %q", s)
	}
}

func TestLevel_AtLeast(t *testing.T) {
	if !severity.Error.AtLeast(severity.Warn) {
		t.Fatal("Error should be at least Warn")
	}
	if severity.Debug.AtLeast(severity.Info) {
		t.Fatal("Debug should not be at least Info")
	}
	if !severity.Warn.AtLeast(severity.Warn) {
		t.Fatal("Warn should be at least Warn")
	}
}

func TestCompare(t *testing.T) {
	if severity.Compare(severity.Debug, severity.Error) != -1 {
		t.Fatal("expected -1")
	}
	if severity.Compare(severity.Fatal, severity.Info) != 1 {
		t.Fatal("expected 1")
	}
	if severity.Compare(severity.Warn, severity.Warn) != 0 {
		t.Fatal("expected 0")
	}
}
