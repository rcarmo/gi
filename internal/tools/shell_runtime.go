package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"

	"github.com/rcarmo/gi/internal/logutil"
)

func RunShellPrompt(ctx context.Context, prompt string, onStart func(*exec.Cmd), onDelta func(string)) (string, error, bool) {
	cmd := exec.Command("sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\"")
	cmd.Env = append(cmd.Environ(), "GI_PROMPT="+prompt)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err, false
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err, false
	}
	if err := cmd.Start(); err != nil {
		return "", err, false
	}
	if onStart != nil {
		onStart(cmd)
	}
	var stdout, stderr bytes.Buffer
	var readWG sync.WaitGroup
	readWG.Add(2)
	go func() {
		defer readWG.Done()
		buf := make([]byte, 128)
		for {
			n, readErr := stdoutPipe.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				stdout.WriteString(chunk)
				if onDelta != nil {
					onDelta(chunk)
				}
			}
			if readErr != nil {
				return
			}
		}
	}()
	go func() {
		defer readWG.Done()
		if _, err := io.Copy(&stderr, stderrPipe); err != nil {
			logutil.WarnIfErr("copy shell stderr", err)
		}
	}()
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
				logutil.WarnIfErr("kill timed out shell process group", err)
			}
		}
		<-waitCh
		readWG.Wait()
		return stdout.String(), nil, true
	case err := <-waitCh:
		readWG.Wait()
		if err != nil {
			if stderr.Len() > 0 {
				return "", fmt.Errorf("%w: %s", err, stderr.String()), false
			}
			return "", err, false
		}
		if stderr.Len() > 0 {
			return stdout.String(), fmt.Errorf("stderr: %s", stderr.String()), false
		}
		return stdout.String(), nil, false
	}
}
