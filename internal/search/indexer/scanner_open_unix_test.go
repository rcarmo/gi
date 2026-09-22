//go:build linux || darwin

package indexer

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/rcarmo/gi/internal/search/chunking"
	searchstore "github.com/rcarmo/gi/internal/search/store"
)

func TestScanNeverOpensSymlinkOrBlocksOnFIFOReplacement(t *testing.T) {
	root, write := scanFixture(t)
	write("notes/a.md", "original")
	dir, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()
	old, err := dir.Lstat("notes/a.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "notes/a.md")); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(root, "notes/a.md"), 0600); err != nil {
		t.Fatal(err)
	}
	// O_NONBLOCK is required: this synchronous call has no writer for the FIFO.
	if _, err := readScanFile(context.Background(), dir, "notes/a.md", old); err == nil {
		t.Fatal("FIFO accepted")
	}
	c := scanConfig(t, root, "notes", []string{"notes"})
	snap, err := ScanScope(t.Context(), c)
	if err != nil || !snap.Complete || len(snap.Documents) != 0 {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "notes/a.md")); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "notes/a.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := readScanFile(t.Context(), dir, "notes/a.md", old); err == nil {
		t.Fatal("symlink replacement accepted")
	}
}

func TestScanUnreadableDirectoryAndBoundedBinaryCandidates(t *testing.T) {
	root, write := scanFixture(t)
	write("notes/secret/a.md", "unreadable")
	c := scanConfig(t, root, "notes", []string{"notes"})
	if os.Geteuid() != 0 {
		p := filepath.Join(root, "notes/secret")
		if err := os.Chmod(p, 0); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(p, 0700)
		if snap, err := ScanScope(t.Context(), c); err == nil || snap.Complete {
			t.Fatal("unreadable treated as empty")
		}
		if err := os.Chmod(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	write("notes/binary.md", string([]byte{0, 255, 1, 2, 3, 4}))
	limits := defaultScanLimits()
	limits.totalText = 4
	if snap, err := scanScope(t.Context(), c, limits, nil); err == nil || snap.Complete {
		t.Fatal("binary bytes bypassed bound")
	}
	// Stock all-scope roots are required until optional-root semantics are wired.
	missing, err := searchstore.DefaultScopeConfig(root, "all", nil, nil, chunking.LineVersion)
	if err != nil {
		t.Fatal(err)
	}
	if snap, err := ScanScope(t.Context(), missing); err == nil || snap.Complete {
		t.Fatal("missing skills root treated as empty")
	}
}
