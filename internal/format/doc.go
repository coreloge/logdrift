// Package format provides flexible log entry formatting for logdrift output.
//
// Three styles are supported:
//
//   - text  — plain single-line human-readable output (default)
//   - json  — compact JSON object per entry
//   - color — ANSI-coloured text, useful for terminal output
//
// Usage:
//
//	f := format.New(format.DefaultOptions())
//	line := f.Format(entry)
//
The Formatter is safe for concurrent use after construction.
package format
