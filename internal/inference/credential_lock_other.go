//go:build !linux && !darwin

package inference

import (
	"errors"
	"os"
)

func lockCredentialFile(file *os.File) error {
	return errors.New("credential writes require Linux or macOS file locking")
}
func unlockCredentialFile(file *os.File) {}
