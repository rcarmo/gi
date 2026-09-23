//go:build linux || darwin

package inference

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

func lockCredentialFile(file *os.File) error {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
		return ErrCredentialConflict
	}
	if err != nil {
		return errors.New("credential lock unavailable")
	}
	return nil
}
func unlockCredentialFile(file *os.File) { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }
