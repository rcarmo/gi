//go:build !linux && !darwin

package web

import (
	"errors"
	"os"
	"os/exec"
)

func terminalSupported() bool { return false }
func startTerminalProcess(*exec.Cmd, int, int) (*os.File, error) {
	return nil, errors.New("web terminal PTY unavailable on this platform")
}
func resizeTerminalProcess(*os.File, int, int) error { return errors.New("PTY resize unavailable") }
func stopTerminalProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
