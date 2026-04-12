// Package enrich provides field enrichment for structured log entries.
//
// An Enricher holds a set of static key-value fields that are merged into
// the Extra map of every diff.Entry that passes through it. Fields can be
// added or removed at runtime without restarting the pipeline.
//
// Basic usage:
//
//	opts := enrich.DefaultOptions()
//	opts.StaticFields = map[string]string{
//		"env":    "production",
//		"region": "us-east-1",
//	}
//	e := enrich.New(opts)
//
//	// Wrap a channel-based pipeline:
//	enriched := enrich.Stream(ctx, e, rawEntries)
//
// By default existing fields on an entry are preserved. Set
// Options.OverwriteExisting = true to let static fields take precedence.
package enrich
