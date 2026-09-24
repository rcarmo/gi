//go:build !linux && !darwin && !windows

package auth

import (
	"errors"
	"os"
)

func lockStateFile(*os.File) error { return errors.New("auth writes require supported file locking") }
func unlockStateFile(*os.File)     {}
func syncStateDir(*os.Root) error  { return nil }
