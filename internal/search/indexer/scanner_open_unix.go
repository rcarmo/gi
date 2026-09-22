//go:build linux || darwin

package indexer

import (
	"os"
	"syscall"
)

// Root confines path resolution. O_NOFOLLOW rejects a final-component swap to a
// symlink; O_NONBLOCK prevents a replacement FIFO from wedging the worker before
// its fstat can reject the special file. Parent identity is rechecked by ScanScope.
func openScanEntry(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
}
