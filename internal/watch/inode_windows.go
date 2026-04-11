//go:build windows

package watch

import "os"

// inoFromInfo returns 0 on Windows where inode is not available;
// rotation is detected by size decrease alone.
func inoFromInfo(info os.FileInfo) uint64 {
	return 0
}
