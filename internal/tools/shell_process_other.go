//go:build !unix && !windows

package tools

import "os/exec"

func configureShellProcess(_ *exec.Cmd) {}

func killShellProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
