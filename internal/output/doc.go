// Package output provides a thread-safe Writer for emitting formatted
// log entries and diff results to an io.Writer destination.
//
// Supported formats:
//
//	- FormatText (default): plain human-readable lines, one per write.
//	- FormatJSON: reserved for structured JSON output (used by downstream
//	  renderers that serialise entries before handing them to the Writer).
//
// Typical usage:
//
//	w := output.New(os.Stdout, output.FormatText)
//	w.WriteLine(render.FormatEntry(entry))
//
// All methods are safe for concurrent use.
package output
