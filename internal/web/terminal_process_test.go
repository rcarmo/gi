//go:build linux || darwin

package web

import (
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestWebTerminalShutdownKillsForegroundJobGroup(t *testing.T) {
	srv, server := terminalFixture(t)
	c := terminalDial(t, server, "child-cleanup-123", "")
	terminalRead(t, c)
	srv.terminals.mu.Lock()
	ts := srv.terminals.sessions["anonymous:child-cleanup-123"]
	srv.terminals.mu.Unlock()
	if err := resizeTerminalProcess(ts.pty, 90, 26); err != nil {
		t.Fatal(err)
	}
	raw, err := ts.pty.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	var flags int
	if err := raw.Control(func(fd uintptr) { flags, err = unix.FcntlInt(fd, unix.F_GETFL, 0) }); err != nil {
		t.Fatal(err)
	}
	if err != nil || flags&unix.O_NONBLOCK == 0 {
		t.Fatal("resize disabled PTY polling", flags, err)
	}
	terminalInput(t, c, "stty -echo; sleep 60 & printf '\\nCHILD_PID=%s\\n' $!; wait\r")
	output := terminalUntil(t, c, "CHILD_PID=")
	// Await a complete numeric line; the echo may contain the command but not digits.
	var pid int
	for pid == 0 {
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "CHILD_PID=") {
				pid, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "CHILD_PID=")))
			}
		}
		if pid == 0 {
			p := terminalRead(t, c)
			if p["type"] == "output" {
				output += p["data"].(string)
			}
		}
	}
	srv.CloseTerminals()
	deadline := time.Now().Add(2 * time.Second)
	for syscall.Kill(pid, 0) == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if syscall.Kill(pid, 0) == nil {
		t.Fatal("child process survived terminal shutdown", pid)
	}
}
