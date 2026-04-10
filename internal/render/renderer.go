// Package render provides terminal output formatting for logdrift.
package render

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/user/logdrift/internal/diff"
)

// Palette holds color functions keyed by log level.
var Palette = map[string]func(...interface{}) string{
	"error": color.New(color.FgRed, color.Bold).SprintFunc(),
	"warn":  color.New(color.FgYellow).SprintFunc(),
	"info":  color.New(color.FgCyan).SprintFunc(),
	"debug": color.New(color.FgWhite).SprintFunc(),
}

// Renderer writes formatted log entries and diff results to an output writer.
type Renderer struct {
	out     io.Writer
	noColor bool
}

// New returns a Renderer writing to out. Pass os.Stdout for normal use.
func New(out io.Writer, noColor bool) *Renderer {
	if noColor {
		color.NoColor = true
	}
	if out == nil {
		out = os.Stdout
	}
	return &Renderer{out: out, noColor: noColor}
}

// Entry prints a single parsed log entry with service prefix and level color.
func (r *Renderer) Entry(service string, e diff.Entry) {
	ts := e.Timestamp.Format(time.RFC3339)
	level := strings.ToLower(e.Level)
	fn, ok := Palette[level]
	if !ok {
		fn = color.New(color.FgWhite).SprintFunc()
	}
	svcTag := color.New(color.FgMagenta, color.Bold).Sprintf("[%s]", service)
	fmt.Fprintf(r.out, "%s %s %s %s\n", svcTag, ts, fn(strings.ToUpper(level)), e.Message)
}

// DiffResult prints a formatted diff result between two services.
func (r *Renderer) DiffResult(result diff.Result) {
	if result.Equal {
		fmt.Fprintf(r.out, "%s\n", color.GreenString("✓ entries match"))
		return
	}
	for _, delta := range result.Deltas {
		fmt.Fprintf(r.out, "%s\n", diff.FormatDelta(delta))
	}
}
