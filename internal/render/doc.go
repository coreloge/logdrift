// Package render handles terminal presentation for logdrift.
//
// It provides colourised output for log entries streamed from multiple
// services and for diff results produced by the diff package.  Colour
// output can be disabled via the noColor flag (or automatically when the
// output is not a TTY, thanks to the fatih/color library).
//
// Typical usage:
//
//	r := render.New(os.Stdout, false)
//	for entry := range entries {
//		r.Entry(entry.Service, entry.Entry)
//	}
package render
