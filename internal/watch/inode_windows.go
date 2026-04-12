//go:build windows

package watch

import "os"

// inoFromInfo returns 0 on Windows where inode numbers are not available
// via the standard os.FileInfo interface. On Windows, file identity and
// rotation detection is handled by monitoring file size decreases and
// file name changes rather than inode comparisons.
func inoFromInfo(info os.FileInfo) uint64 {
	return 0
}

// sameFile reports whether two FileInfo values refer to the same file.
// On Windows, inodes are unavailable, so we fall back to os.SameFile
// which uses the underlying file index information from the OS.
func sameFile(a, b os.FileInfo) bool {
	return os.SameFile(a, b)
}
