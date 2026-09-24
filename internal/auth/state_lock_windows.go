//go:build windows

package auth

import (
	"errors"
	"golang.org/x/sys/windows"
	"os"
)

func lockStateFile(file *os.File) error {
	err := windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return ErrStateConflict
	}
	return err
}
func unlockStateFile(file *os.File) {
	_ = windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, &windows.Overlapped{})
}

// Go's Root.Rename uses Windows replacement semantics; the temporary file is
// flushed before replacement. Directory fsync is not portable on Windows:
// sudden-power-loss durability of the replaced directory entry is NOT promised.
// Never fall back to truncation if Windows rejects the replacement.
func syncStateDir(_ *os.Root) error { return nil }
