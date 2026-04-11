// Package watch provides file system watching capabilities for logdrift,
// detecting when log source files are rotated or recreated.
package watch

import (
	"context"
	"os"
	"time"
)

// Event represents a file system event for a watched path.
type Event struct {
	Path    string
	Rotated bool // true if the file was recreated (inode changed or truncated)
}

// Watcher polls a set of file paths and emits Events when rotation is detected.
type Watcher struct {
	paths    []string
	interval time.Duration
	inodes   map[string]uint64
	sizes    map[string]int64
	Events   chan Event
}

// New creates a Watcher for the given paths, polling at the given interval.
func New(paths []string, interval time.Duration) *Watcher {
	return &Watcher{
		paths:    paths,
		interval: interval,
		inodes:   make(map[string]uint64),
		sizes:    make(map[string]int64),
		Events:   make(chan Event, len(paths)),
	}
}

// Start begins polling in a goroutine and stops when ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				close(w.Events)
				return
			case <-ticker.C:
				w.poll()
			}
		}
	}()
}

func (w *Watcher) poll() {
	for _, p := range w.paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		inode := inoFromInfo(info)
		size := info.Size()
		prevInode, seenInode := w.inodes[p]
		prevSize := w.sizes[p]
		w.inodes[p] = inode
		w.sizes[p] = size
		if !seenInode {
			continue
		}
		rotated := inode != prevInode || size < prevSize
		if rotated {
			w.Events <- Event{Path: p, Rotated: true}
		}
	}
}
