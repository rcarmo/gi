//go:build linux

package turn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/scripting"
)

func hookProcessRunning(pid int) bool {
	if syscall.Kill(pid, 0) != nil {
		return false
	}
	// Orphan descendants are reaped by init, not by exec.Cmd.Wait. A zombie
	// has terminated and holds no hook capabilities or open descriptors.
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if os.IsNotExist(err) {
		return false
	}
	if err == nil {
		after := strings.LastIndexByte(string(data), ')')
		if after >= 0 && strings.HasPrefix(string(data)[after+1:], " Z") {
			return false
		}
	}
	return true
}
func TestProcessHookLifecycleOwnsProcessGroupAndReaders(t *testing.T) {
	for _, boundary := range []string{"close", "clear", "unregister", "timeout", "blocked_write"} {
		t.Run(boundary, func(t *testing.T) {
			root := t.TempDir()
			childFile := filepath.Join(root, "child.pid")
			scriptPath := filepath.Join(root, "hook.sh")
			script := fmt.Sprintf(`#!/bin/sh
set -eu
sleep 600 >&2 &
printf '%%s %%s\n' "$$" "$!" > %q
while IFS= read -r line; do
 case "$line" in
 *'"method":"hook.hello"'*)
  printf '%%s\n' '{"jsonrpc":"2.0","id":1,"result":{"ok":true}}'
  if [ %q = blocked_write ]; then wait; fi;;
 *'"method":"hook.before_tool"'*)
  if [ %q = timeout ]; then wait; fi
  printf '%%s\n' '{"jsonrpc":"2.0","id":2,"result":{}}';;
 esac
done
`, childFile, boundary, boundary)
			if err := os.WriteFile(scriptPath, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			s := openTestStore(t)
			e := New(s)
			defer s.Close()
			defer e.Close()
			unregister, err := e.registerProcessHook(root, scripting.EventHookSpec{Name: HookToolCall, Engine: "process", Command: scriptPath})
			if err != nil {
				t.Fatal(err)
			}
			e.hookMu.Lock()
			var mounted *mountedProcessHook
			for process := range e.hookProcesses {
				mounted = process
			}
			e.hookMu.Unlock()
			timeout := boundary == "timeout" || boundary == "blocked_write"
			if timeout {
				e.runtimeCfg.Hooks.TimeoutMS = 100
				e.runtimeCfg.Hooks.OnTimeout = "error"
			}
			req := HookRequest{Name: HookToolCall}
			if boundary == "blocked_write" {
				req.Payload = map[string]any{"large": strings.Repeat("x", 1<<20)}
			}
			_, err = e.emitHook(context.Background(), req)
			if timeout {
				if err == nil {
					t.Fatal("expected timeout")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(childFile)
			if err != nil {
				t.Fatal(err)
			}
			pids := strings.Fields(string(data))
			if len(pids) != 2 {
				t.Fatal("invalid process identities", string(data))
			}
			parentPID, err := strconv.Atoi(pids[0])
			if err != nil {
				t.Fatal(err)
			}
			defer syscall.Kill(-parentPID, syscall.SIGKILL)
			childPID, err := strconv.Atoi(pids[1])
			if err != nil {
				t.Fatal(err)
			}
			defer syscall.Kill(childPID, syscall.SIGKILL)
			switch boundary {
			case "close":
				if err := e.Close(); err != nil {
					t.Fatal(err)
				}
			case "clear":
				e.ClearHooks()
			case "unregister":
				unregister()
			case "timeout", "blocked_write":
				closed := make(chan struct{})
				go func() { e.closeHookHosts(); close(closed) }()
				select {
				case <-closed:
				case <-time.After(time.Second):
					t.Fatal("process cancellation failed to interrupt IO")
				}
			}
			waitForCondition(t, time.Second, func() bool { return !hookProcessRunning(childPID) }, "hook descendant terminated at lifecycle boundary")
			mounted.mu.Lock()
			cmd, started := mounted.cmd, mounted.started
			mounted.mu.Unlock()
			if cmd != nil || started {
				t.Fatal("mounted process retained live ownership")
			}
			// Closing the mounted host joins both stderr and JSON decode goroutines.
			joined := make(chan struct{})
			go func() { mounted.ioHosts.Wait(); close(joined) }()
			select {
			case <-joined:
			case <-time.After(time.Second):
				t.Fatal("hook IO reader escaped teardown")
			}
			if _, err := mounted.invoke(context.Background(), HookRequest{Name: HookToolCall}); err == nil {
				t.Fatal("closed mount restarted")
			}
			if boundary == "close" {
				if _, err := e.registerProcessHook(root, scripting.EventHookSpec{Name: HookToolCall, Command: scriptPath}); err == nil {
					t.Fatal("registration bypassed shutdown fence")
				}
			}
		})
	}
}
