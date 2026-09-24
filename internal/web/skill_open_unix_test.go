//go:build linux || darwin

package web

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestWebSkillOpenConfinesSymlinksAndDoesNotBlockOnReplacementFIFO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skill")
	if err := os.WriteFile(path, []byte("regular"), 0600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	// Reproduce the swap after the reader's initial regular-file check.
	if info, err := root.Stat("skill"); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("initial file: %v %v", info, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		file, err := openWebSkill(root, "skill")
		if err == nil {
			defer file.Close()
			info, statErr := file.Stat()
			err = statErr
			if err == nil && info.Mode().IsRegular() {
				err = syscall.EINVAL
			}
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		// Unblock a broken blocking implementation before failing the test.
		writer, _ := os.OpenFile(path, os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if writer != nil {
			writer.Close()
		}
		t.Fatal("skill open blocked on replacement FIFO")
	}
	if _, err := readWebSkill(dir, "skill"); err == nil {
		t.Fatal("FIFO accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "target"), []byte("same workspace"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", path); err != nil {
		t.Fatal(err)
	}
	// os.Root follows in-root links, even with O_NOFOLLOW. The contract is
	// workspace confinement plus content hashing, not blanket link rejection.
	if raw, err := readWebSkill(dir, "skill"); err != nil || string(raw) != "same workspace" {
		t.Fatalf("in-root link: %q %v", raw, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("external"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, path); err != nil {
		t.Fatal(err)
	}
	if _, err := readWebSkill(dir, "skill"); err == nil {
		t.Fatal("escaping symlink accepted")
	}
}
