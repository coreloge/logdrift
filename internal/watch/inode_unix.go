//go:build !windows

package watch

import (
	"os"
	"syscall"
)

// inoFromInfo extracts the inode number from a FileInfo on Unix systems.
func inoFromInfo(info os.FileInfo) uint64 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return stat.Ino
	}
	return 0
}
