package checkpoint

import (
	"context"

	"github.com/yourorg/logdrift/internal/diff"
)

// Stream wraps an entry channel and updates the checkpoint store whenever
// a new entry is forwarded, recording the service and approximate offset.
// The offset is incremented by the byte length of the formatted entry.
func Stream(
	ctx context.Context,
	in <-chan diff.Entry,
	store *Store,
) <-chan diff.Entry {
	out := make(chan diff.Entry)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-in:
				if !ok {
					return
				}
				updateOffset(store, e)
				select {
				case out <- e:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

func updateOffset(store *Store, e diff.Entry) {
	current, _ := store.Get(e.Service)
	line := diff.FormatEntry(e)
	_ = store.Set(e.Service, current+int64(len(line))+1)
}
