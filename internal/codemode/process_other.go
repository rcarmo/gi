//go:build !unix

package codemode

import (
 "os/exec"
 "time"
)
func prepareProcess(cmd *exec.Cmd) { cmd.WaitDelay = time.Second }
func killProcess(cmd *exec.Cmd) { if cmd.Process != nil { _ = cmd.Process.Kill() } }
