//go:build unix

package codemode

import (
 "os/exec"
 "syscall"
 "time"
)
func prepareProcess(cmd *exec.Cmd) {
 cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
 cmd.Cancel = func() error { killProcess(cmd); return nil }
 cmd.WaitDelay = time.Second
}
func killProcess(cmd *exec.Cmd) {
 if cmd.Process != nil { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); _ = cmd.Process.Kill() }
}
