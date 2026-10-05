//go:build linux || darwin

package web

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

func terminalSupported() bool { return true }
func startTerminalProcess(cmd *exec.Cmd, cols, rows int) (*os.File, error) {
	file, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	if err != nil {
		return nil, err
	}
	// NewFile must see O_NONBLOCK to register the PTY with Go's poller.
	// A blocking master can otherwise strand Read/Write during shutdown.
	fd, err := syscall.Dup(int(file.Fd()))
	if err == nil {
		err = syscall.SetNonblock(fd, true)
	}
	if err != nil {
		if fd >= 0 {
			_ = syscall.Close(fd)
		}
		_ = file.Close()
		stopTerminalProcess(cmd)
		_ = cmd.Wait()
		return nil, err
	}
	_ = file.Close()
	return os.NewFile(uintptr(fd), "terminal-pty"), nil
}
func resizeTerminalProcess(f *os.File, cols, rows int) error {
	// File.Fd can switch a pollable descriptor back to blocking mode. Keep
	// deadlines and Close interruption intact while applying the ioctl.
	raw, err := f.SyscallConn()
	if err != nil {
		return err
	}
	var ioctlErr error
	err = raw.Control(func(fd uintptr) {
		ioctlErr = unix.IoctlSetWinsize(int(fd), unix.TIOCSWINSZ, &unix.Winsize{Col: uint16(cols), Row: uint16(rows)})
	})
	if err != nil {
		return err
	}
	return ioctlErr
}
func stopTerminalProcess(cmd *exec.Cmd) {
	// Interactive job control gives foreground jobs their own process groups.
	// Walk descendants before killing the session leader, then target only
	// those descendants/groups. No broad process-name matching.
	if cmd.Process == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if output, err := exec.CommandContext(ctx, "ps", "-axo", "pid=,ppid=").Output(); err == nil {
		children := map[int][]int{}
		for _, line := range strings.Split(string(output), "\n") {
			fields := strings.Fields(line)
			if len(fields) != 2 {
				continue
			}
			pid, _ := strconv.Atoi(fields[0])
			parent, _ := strconv.Atoi(fields[1])
			if pid > 0 {
				children[parent] = append(children[parent], pid)
			}
		}
		var killChildren func(int)
		killChildren = func(parent int) {
			for _, pid := range children[parent] {
				killChildren(pid)
				_ = syscall.Kill(-pid, syscall.SIGKILL)
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
		killChildren(cmd.Process.Pid)
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	_ = cmd.Process.Kill()
}
