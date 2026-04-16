// Package drop provides a pipeline stage that discards log entries matching
// one or more field/pattern rules.
//
// Rules are evaluated in order; the first match causes the entry to be dropped.
// An empty field name targets the entry message. The special field names
// "level" and "service" target those scalar fields; all other names are looked
// up in Entry.Extra.
//
// Usage:
//
//	d, err := drop.New(drop.Options{
//		Rules: []drop.Rule{
//			{Field: "level",   Pattern: "^debug$"},
//			{Field: "message", Pattern: "healthcheck"},
//		},
//	})
//	out := drop.Stream(ctx, in, d)
package drop
