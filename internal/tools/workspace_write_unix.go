//go:build linux || darwin

package tools

import (
	"os"
	"syscall"
)

func openWorkspaceWrite(root *os.Root, name string) (*os.File, error) {
	// Avoid truncating a symlink replacement and avoid blocking on a FIFO before
	// the caller can check regular-file identity. Truncation follows fstat.
	return root.OpenFile(name, os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0644)
}
