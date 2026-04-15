// Package head implements a pipeline stage that limits a log entry stream to
// the first N entries — mirroring the behaviour of the Unix `head` utility.
//
// # Usage
//
//	h, err := head.New(head.Options{Max: 20})
//	if err != nil {
//		log.Fatal(err)
//	}
//	out := h.Stream(ctx, in)
//
// The output channel is closed as soon as Max entries have been forwarded or
// the context is cancelled, whichever comes first.  The upstream producer is
// not explicitly cancelled; callers should pair this stage with a context that
// they can cancel independently if early termination of upstream work is
// required.
package head
