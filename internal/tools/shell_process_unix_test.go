//go:build unix

package tools

import (
	"bufio"
	"os/exec"
	"testing"
	"time"
)

func TestKillShellProcessTerminatesDescendantPipe(t *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 60 & printf 'ready\\n'; wait")
	configureShellProcess(cmd)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pipe.Close()
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer killShellProcess(cmd)
	reader := bufio.NewReader(pipe)
	if _, err = reader.ReadString('\n'); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := reader.ReadByte(); done <- err }()
	killShellProcess(cmd)
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("unexpected output after kill")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("descendant retained output pipe")
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("killed shell exited successfully")
	}
}
