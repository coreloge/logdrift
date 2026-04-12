// Package rotate detects log file rotation events (truncation or inode
// replacement) and notifies callers so they can re-open the file and
// resume tailing from the beginning.
//
// # Usage
//
//	opts := rotate.DefaultOptions()
//	opts.PollInterval = 250 * time.Millisecond
//
//	r, err := rotate.New(ctx, "/var/log/app.log", opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for ev := range r.Events {
//	    fmt.Printf("rotation detected for %s (reopen #%d)\n", ev.Path, ev.Reopens)
//	    // re-open and tail the file from offset 0
//	}
//
// The Rotator relies on [watch.Watcher] for low-level inode and size
// change detection, adding reopen counting and a clean shutdown path
// via context cancellation.
package rotate
