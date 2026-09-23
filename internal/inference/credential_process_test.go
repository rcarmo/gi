//go:build linux || darwin

package inference

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

func TestCredentialProcessHelper(t *testing.T) {
	if os.Getenv("GI_CREDENTIAL_TEST_MODE") == "" {
		return
	}
	switch os.Getenv("GI_CREDENTIAL_TEST_MODE") {
	case "hold":
		file, err := os.OpenFile(filepath.Join(filepath.Dir(AuthFilePath()), ".gi-auth.lock"), os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if err = lockCredentialFile(file); err != nil {
			t.Fatal(err)
		}
		defer unlockCredentialFile(file)
		fmt.Println("held")
		io.Copy(io.Discard, os.Stdin)
	case "consume":
		key, _, err := loadAuth("openai")
		if err != nil || key != "fixture-process-key" {
			t.Fatal("reopened credential was not consumed")
		}
	}
}
func TestCredentialCrossProcessLockAndReopen(t *testing.T) {
	path := credentialFixture(t, `{}`)
	snapshot, err := ReadProviderSettings()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCredentialProcessHelper$", "-test.timeout=15s")
	cmd.Env = append(os.Environ(), "GI_CREDENTIAL_TEST_MODE=hold")
	input, _ := cmd.StdinPipe()
	output, _ := cmd.StdoutPipe()
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
		t.Fatal("lock handshake failed")
	}
	if err = SaveProviderKey("openai", snapshot.Revision, "fixture-process-key"); !errors.Is(err, ErrCredentialConflict) {
		t.Fatalf("lock not enforced: %v", err)
	}
	if _, err = RemoveAuthEntry("openai"); !errors.Is(err, ErrCredentialConflict) {
		t.Fatal("logout bypassed lock")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "{}" {
		t.Fatal("failed write changed credentials")
	}
	input.Close()
	if err = cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if err = SaveProviderKey("openai", snapshot.Revision, "fixture-process-key"); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestCredentialProcessHelper$", "-test.timeout=15s")
	child.Env = append(os.Environ(), "GI_CREDENTIAL_TEST_MODE=consume")
	if _, err = child.CombinedOutput(); err != nil {
		t.Fatal("fresh-process credential load failed")
	}
}
func TestCredentialRejectsFIFOAndDirectoryLinks(t *testing.T) {
	path := credentialFixture(t, `{}`)
	os.Remove(path)
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadProviderSettings(); err == nil {
		t.Fatal("FIFO accepted")
	}
	if _, err := RemoveAuthEntry("x"); err == nil {
		t.Fatal("FIFO replaced")
	}
	os.Remove(path)
	if err := os.RemoveAll(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadProviderSettings(); err == nil {
		t.Fatal("symlink directory read")
	}
}
