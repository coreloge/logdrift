// Package buffer implements a fixed-capacity ring buffer for log entries.
//
// The Ring type stores the N most recent [diff.Entry] values written to it,
// automatically discarding the oldest entry once capacity is reached. All
// operations are safe for concurrent use.
//
// Typical usage:
//
//	buf := buffer.New(512)
//	buf.Push(entry)
//	entries := buf.Snapshot() // chronological copy
//
// The buffer is useful wherever a bounded, in-memory window of recent log
// activity is required — for example, feeding a diff snapshot or providing
// context when an alert fires.
package buffer
