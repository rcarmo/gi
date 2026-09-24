package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func configureShellProcess(_ *exec.Cmd) {}

func killShellProcess(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	// sh may have spawned descendants that retain the output handles. taskkill
	// provides Windows tree termination without importing Unix process syscalls.
	// Resolve from SystemRoot rather than the workspace/PATH.
	if root := os.Getenv("SystemRoot"); root != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, filepath.Join(root, "System32", "taskkill.exe"), "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F").Run()
	}
	_ = cmd.Process.Kill()
}
