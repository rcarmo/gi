//go:build linux || darwin

package auth

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

func lockStateFile(file *os.File) error {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
		return ErrStateConflict
	}
	return err
}
func unlockStateFile(file *os.File) { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }
func syncStateDir(root *os.Root) error {
	dir, err := root.Open(".")
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
