//go:build linux || darwin

package tools

import (
	"path/filepath"
	"syscall"
	"testing"
)

func TestWriteRejectsFIFOWithoutWaitingForReader(t *testing.T) {
	db, cfg, scopes := producerFixture(t)
	if err := syscall.Mkfifo(filepath.Join(cfg.WorkspaceRoot, "notes/pipe.md"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(t.Context(), cfg, db, "notes/pipe.md", "wrong"); err == nil {
		t.Fatal("FIFO accepted")
	}
	for _, c := range scopes {
		if st := producerStatus(t, db, c); st.RequestedRevision != 0 {
			t.Fatal(st)
		}
	}
}
