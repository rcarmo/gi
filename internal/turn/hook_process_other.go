//go:build !unix

package turn

import "os/exec"

func configureHookProcess(*exec.Cmd) {}
func killHookProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
