//go:build linux || darwin

package web

import (
	"os"
	"syscall"
)

// Root confines symlink resolution to the workspace (in-root links are allowed).
// NONBLOCK lets the caller reject a swapped FIFO using fstat before reading.
func openWebSkill(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
