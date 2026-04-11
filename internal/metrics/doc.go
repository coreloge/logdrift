// Package metrics provides lightweight counters and exporters for tracking
// logdrift runtime statistics.
//
// # Counter
//
// Counter records per-service entry and drift counts and is safe for
// concurrent use. Call RecordEntry for every log line processed and
// RecordDrift whenever a DiffResult contains deltas.
//
// # Exporter
//
// Exporter snapshots a Counter and writes the result to any io.Writer in
// either human-readable text (tabwriter) or machine-readable JSON format.
//
//	exporter := metrics.NewExporter(counter, "json")
//	exporter.Export(os.Stdout)
package metrics
