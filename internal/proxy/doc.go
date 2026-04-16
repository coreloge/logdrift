// Package proxy implements a transparent pass-through pipeline stage.
//
// A Proxy sits in a pipeline and calls a user-supplied HookFunc for every
// entry that flows through it. The entry is forwarded downstream unchanged,
// making Proxy suitable for metrics collection, debug logging, or any
// side-effect that should not alter the stream.
//
// Basic usage:
//
//	p, err := proxy.New(proxy.Options{
//		Hook: func(e diff.Entry) {
//			fmt.Println("saw:", e.Message)
//		},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	out := proxy.Stream(ctx, p, in)
package proxy
