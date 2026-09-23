//go:build !linux && !darwin

package config

import (
	"errors"
	"os"
)

func lockIdentityFile(file *os.File) error {
	return errors.New("identity writes require supported file locking on Linux or macOS")
}
func unlockIdentityFile(file *os.File) {}
