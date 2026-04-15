// Package cursor provides a lightweight, thread-safe store for tracking read
// positions (byte offset + line number) within log streams.
//
// Typical usage:
//
//	store := cursor.New()
//
//	// Restore a previously saved position before starting the tailer.
//	if pos, err := store.Get("auth-service"); err == nil {
//		// seek tailer to pos.Offset
//	}
//
//	// Wrap an entry channel so positions are updated automatically.
//	tracked := cursor.Stream(ctx, store, entryCh)
//
// Positions are held in memory only; persist them via the checkpoint package
// if durability across process restarts is required.
package cursor
