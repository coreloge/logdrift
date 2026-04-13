// Package format provides log entry formatting for different output styles.
package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/logdrift/internal/diff"
)

// Style controls the output format of a formatted log entry.
type Style string

const (
	StyleText  Style = "text"
	StyleJSON  Style = "json"
	StyleColor Style = "color"
)

// Options configures the Formatter.
type Options struct {
	Style          Style
	TimestampField string
	ShowService    bool
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Style:          StyleText,
		TimestampField: "time",
		ShowService:    true,
	}
}

// Formatter converts log entries to human-readable strings.
type Formatter struct {
	opts Options
}

// New creates a Formatter with the given options.
func New(opts Options) *Formatter {
	if opts.Style == "" {
		opts.Style = StyleText
	}
	return &Formatter{opts: opts}
}

// Format renders a single log entry as a string.
func (f *Formatter) Format(entry diff.Entry) string {
	switch f.opts.Style {
	case StyleJSON:
		return formatJSON(entry, f.opts)
	case StyleColor:
		return formatColor(entry, f.opts)
	default:
		return formatText(entry, f.opts)
	}
}

func formatText(entry diff.Entry, opts Options) string {
	var sb strings.Builder
	if ts, ok := entry.Fields[opts.TimestampField]; ok {
		sb.WriteString(fmt.Sprintf("[%v] ", ts))
	} else {
		sb.WriteString(fmt.Sprintf("[%s] ", time.Now().Format(time.RFC3339)))
	}
	if opts.ShowService && entry.Service != "" {
		sb.WriteString(fmt.Sprintf("%-12s ", entry.Service))
	}
	sb.WriteString(fmt.Sprintf("%-7s %s", entry.Level, entry.Message))
	return sb.String()
}

func formatJSON(entry diff.Entry, opts Options) string {
	parts := []string{
		fmt.Sprintf(`"level":%q`, entry.Level),
		fmt.Sprintf(`"message":%q`, entry.Message),
	}
	if opts.ShowService && entry.Service != "" {
		parts = append([]string{fmt.Sprintf(`"service":%q`, entry.Service)}, parts...)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func formatColor(entry diff.Entry, opts Options) string {
	const (
		red    = "\033[31m"
		yellow = "\033[33m"
		green  = "\033[32m"
		reset  = "\033[0m"
	)
	color := reset
	switch strings.ToLower(entry.Level) {
	case "error", "fatal":
		color = red
	case "warn", "warning":
		color = yellow
	case "info":
		color = green
	}
	return color + formatText(entry, opts) + reset
}
