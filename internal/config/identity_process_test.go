//go:build linux || darwin

package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestIdentityProcessHelper(t *testing.T) {
	root := os.Getenv("GI_IDENTITY_TEST_ROOT")
	if root == "" {
		return
	}
	switch os.Getenv("GI_IDENTITY_TEST_MODE") {
	case "hold":
		f, err := os.OpenFile(filepath.Join(root, ".piclaw", ".gi-identity.lock"), os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err = lockIdentityFile(f); err != nil {
			t.Fatal(err)
		}
		defer unlockIdentityFile(f)
		fmt.Println("held")
		_, _ = io.Copy(io.Discard, os.Stdin)
	case "load":
		cfg := Load(root)
		if cfg.AssistantName != "After restart" || cfg.UserName != "Reader Two" {
			t.Fatalf("fresh process got %q %q", cfg.AssistantName, cfg.UserName)
		}
	}
}

func TestIdentityProcessLockAndRestart(t *testing.T) {
	root, path := identityFixture(t, `{"assistant":{"assistantName":"Before restart"},"user":{"userName":"Reader"}}`)
	snap, err := ReadIdentity(root)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	cmd := exec.Command(os.Args[0], "-test.run=^TestIdentityProcessHelper$", "-test.timeout=15s")
	cmd.Env = append(os.Environ(), "GI_IDENTITY_TEST_ROOT="+root, "GI_IDENTITY_TEST_MODE=hold")
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = in.Close()
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	line, err := bufio.NewReader(out).ReadString('\n')
	if err != nil || line != "held\n" {
		t.Fatalf("lock handshake %q %v", line, err)
	}
	if _, err = SaveIdentity(root, snap.Revision, IdentityNames{"Blocked", "Reader Two"}); !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("cross-process contention must be retryable conflict: %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Fatal("locked save changed file")
	}
	in.Close()
	if err = cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if _, err = SaveIdentity(root, snap.Revision, IdentityNames{"After restart", "Reader Two"}); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestIdentityProcessHelper$", "-test.timeout=15s")
	child.Env = append(os.Environ(), "GI_IDENTITY_TEST_ROOT="+root, "GI_IDENTITY_TEST_MODE=load")
	if output, err := child.CombinedOutput(); err != nil {
		t.Fatalf("fresh process: %s %v", output, err)
	}
}

func TestIdentityRejectsFIFO(t *testing.T) {
	root, path := identityFixture(t, `{}`)
	os.Remove(path)
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadIdentity(root); err == nil {
		t.Fatal("FIFO read admitted")
	}
	if _, err := SaveIdentity(root, "missing", IdentityNames{"New", "User"}); err == nil {
		t.Fatal("FIFO replaced")
	}
}
