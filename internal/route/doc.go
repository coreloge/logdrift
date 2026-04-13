// Package route provides content-based routing for log entry streams.
//
// A Router inspects configurable fields on each [diff.Entry] and dispatches it
// to a named subscriber channel. Rules are evaluated in declaration order; the
// first matching rule wins. Entries that match no rule are sent to the
// DefaultRoute, if one is configured, or silently dropped.
//
// # Basic usage
//
//	opts := route.Options{
//		Rules: []route.Rule{
//			{Name: "errors", Field: "level", Values: []string{"error", "fatal"}},
//			{Name: "slow",   Field: "service", Values: []string{"payment-svc"}},
//		},
//		DefaultRoute: "general",
//	}
//
//	r := route.New(opts)
//	errCh   := r.Subscribe("errors")
//	slowCh  := r.Subscribe("slow")
//	general := r.Subscribe("general")
//
//	go r.Run(ctx, incomingEntries)
package route
