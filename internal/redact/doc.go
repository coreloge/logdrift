// Package redact implements field-level redaction for structured log entries.
//
// A Redactor is configured with a set of Rules, each targeting a field name
// and an optional regexp pattern. When Apply is called, any field whose name
// matches a rule (optionally case-folded) and whose value matches the pattern
// (if provided) is replaced with the rule's Replacement string.
//
// The Stream helper wraps a channel of diff.LogEntry values, applying
// redaction transparently as entries flow through the pipeline so that
// sensitive data is never forwarded to renderers or writers.
//
// Usage:
//
//	opts := redact.DefaultOptions()
//	r := redact.New(opts)
//	redacted := redact.Stream(ctx, r, entryChan)
package redact
