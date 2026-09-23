//go:build linux || darwin

package config

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func lockIdentityFile(file *os.File) error {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
		return fmt.Errorf("%w: another process is saving identity", ErrIdentityConflict)
	}
	return err
}
func unlockIdentityFile(file *os.File) { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }
