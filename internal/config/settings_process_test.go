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

func TestSettingsProcessHelper(t *testing.T) {
	root := os.Getenv("GI_SETTINGS_TEST_ROOT")
	if root == "" {
		return
	}
	switch os.Getenv("GI_SETTINGS_TEST_MODE") {
	case "hold":
		file, err := os.OpenFile(filepath.Join(root, ".pi", ".gi-settings.lock"), os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if err = lockIdentityFile(file); err != nil {
			t.Fatal(err)
		}
		defer unlockIdentityFile(file)
		fmt.Println("held")
		_, _ = io.Copy(io.Discard, os.Stdin)
	case "restart":
		cfg := Load(root)
		if cfg.Compaction.ThresholdTokens != 50000 || !cfg.Compaction.Enabled || cfg.Compaction.ContextWindow != 64000 || cfg.Compaction.ReserveTokens != 4000 || cfg.Compaction.KeepRecentTokens != 8000 {
			t.Fatalf("fresh process policy %+v", cfg.Compaction)
		}
	}
}

func TestSettingsProcessLockAndPolicyRestart(t *testing.T) {
	root, path := settingsFixture(t, `{"compaction":{"enabled":false}}`)
	snap, err := ReadCompactionPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	cmd := exec.Command(os.Args[0], "-test.run=^TestSettingsProcessHelper$", "-test.timeout=15s")
	cmd.Env = append(os.Environ(), "GI_SETTINGS_TEST_ROOT="+root, "GI_SETTINGS_TEST_MODE=hold")
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		input.Close()
		if cmd.ProcessState == nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()
	line, err := bufio.NewReader(output).ReadString('\n')
	if err != nil || line != "held\n" {
		t.Fatalf("handshake %q %v", line, err)
	}
	if _, err = SaveCompactionPolicy(root, snap.Revision, validPolicy()); !errors.Is(err, ErrSettingsConflict) {
		t.Fatalf("policy contention %v", err)
	}
	for _, save := range []func() error{
		func() error { return PersistModelSelection(root, "test", "test-model", "low", nil) }, func() error { return PersistClipboardMode(root, "native") },
		func() error { return PersistScrollbackLimit(root, 123) }, func() error { return PersistTUIHistoryLimit(root, 456) }, func() error { return PersistTUIScrollbar(root, true) },
	} {
		if err = save(); !errors.Is(err, ErrSettingsConflict) {
			t.Fatalf("legacy writer ignored shared lock: %v", err)
		}
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("blocked save modified file")
	}
	input.Close()
	if err = cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	saved, err := SaveCompactionPolicy(root, snap.Revision, validPolicy())
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestSettingsProcessHelper$", "-test.timeout=15s")
	child.Env = append(os.Environ(), "GI_SETTINGS_TEST_ROOT="+root, "GI_SETTINGS_TEST_MODE=restart")
	if out, err := child.CombinedOutput(); err != nil {
		t.Fatalf("restart %s %v", out, err)
	}
	if snap.Policy == saved.Policy {
		t.Fatal("original snapshot mutated")
	}
	if matches, _ := filepath.Glob(filepath.Join(root, ".pi", "*.tmp")); len(matches) != 0 {
		t.Fatalf("tmp leak %v", matches)
	}
}

func TestSettingsRejectsFIFOAndSymlinkDirectory(t *testing.T) {
	root, path := settingsFixture(t, `{}`)
	os.Remove(path)
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadCompactionPolicy(root); err == nil {
		t.Fatal("FIFO read")
	}
	if err := PersistTUIScrollbar(root, true); err == nil {
		t.Fatal("FIFO replaced")
	}
	other := t.TempDir()
	os.Symlink(filepath.Join(root, ".pi"), filepath.Join(other, ".pi"))
	if _, err := ReadCompactionPolicy(other); err == nil {
		t.Fatal("symlink directory read")
	}
	if err := PersistTUIHistoryLimit(other, 3); err == nil {
		t.Fatal("symlink directory write")
	}
}
