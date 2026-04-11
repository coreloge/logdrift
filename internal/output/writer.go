// Package output handles writing formatted log entries and diff results
// to various output destinations (stdout, file, etc.).
package output

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Format controls how output is serialised.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Writer is a thread-safe output writer that supports multiple formats
// and configurable destinations.
type Writer struct {
	mu     sync.Mutex
	out    io.Writer
	format Format
}

// New creates a Writer that writes to dst using the given format.
// If dst is nil, os.Stdout is used.
func New(dst io.Writer, format Format) *Writer {
	if dst == nil {
		dst = os.Stdout
	}
	if format == "" {
		format = FormatText
	}
	return &Writer{out: dst, format: format}
}

// WriteLine writes a single pre-formatted line followed by a newline.
func (w *Writer) WriteLine(line string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := fmt.Fprintln(w.out, line)
	return err
}

// WriteLines writes multiple lines in order, each followed by a newline.
// Returns the first error encountered, if any.
func (w *Writer) WriteLines(lines []string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, l := range lines {
		if _, err := fmt.Fprintln(w.out, l); err != nil {
			return err
		}
	}
	return nil
}

// Format returns the configured output format.
func (w *Writer) Format() Format {
	return w.format
}

// WriteLinef formats a line according to the given format specifier and
// writes it followed by a newline. It is equivalent to calling
// WriteLine(fmt.Sprintf(format, args...)).
func (w *Writer) WriteLinef(format string, args ...any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := fmt.Fprintf(w.out, format+"\n", args...)
	return err
}
