package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	"github.com/rcarmo/gi/internal/store"
)

func indexPanelFixture(t *testing.T) *chatTUI {
	t.Helper()
	root := t.TempDir()
	for _, p := range []string{"notes", ".pi/skills"} {
		if err := os.MkdirAll(filepath.Join(root, p), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "notes/a.md"), []byte("native orchid"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	c := &chatTUI{store: db, cfg: config.RuntimeConfig{WorkspaceRoot: root}, sessionID: "test", sessionGeneration: 1, inputActive: true, transcriptScroll: 12, stickToBottom: false, transcript: []string{"kept"}}
	c.ensureInput()
	c.input.SetText("draft 世界 tail")
	c.input.cursorPos = 8
	c.initWorkspaceIndex()
	t.Cleanup(c.stopWorkspaceIndex)
	return c
}
func indexPanelResult(t *testing.T, c *chatTUI) {
	t.Helper()
	select {
	case r := <-c.workspaceIndex.results:
		c.handleWorkspaceIndexResult(r)
	case <-time.After(5 * time.Second):
		t.Fatal("index operation stuck")
	}
}
func TestIndexPanelNativeStatusRefreshFailureAndEditorIsolation(t *testing.T) {
	c := indexPanelFixture(t)
	before := []any{c.input.Text(), c.input.cursorPos, c.input.undoText, c.input.undoCursor, c.input.yankText}
	transcript := append([]string(nil), c.transcript...)
	c.openWorkspaceIndex()
	indexPanelResult(t, c)
	if c.workspaceIndex.status.State != "never_indexed" {
		t.Fatal(c.workspaceIndex.status)
	}
	c.onSubmit("must not submit")
	c.requestWorkspaceIndex(true)
	indexPanelResult(t, c)
	if c.workspaceIndex.err != nil || c.workspaceIndex.status.State != "ready" || c.workspaceIndex.status.Generation != 1 {
		t.Fatal(c.workspaceIndex.status, c.workspaceIndex.err)
	}
	if err := os.Rename(filepath.Join(c.cfg.WorkspaceRoot, "notes"), filepath.Join(c.cfg.WorkspaceRoot, "held")); err != nil {
		t.Fatal(err)
	}
	c.requestWorkspaceIndex(true)
	indexPanelResult(t, c)
	if c.workspaceIndex.err == nil {
		t.Fatal("failed scan reported success")
	}
	for _, width := range []int{60, 100, 140} {
		lines := c.workspaceIndexLines(width)
		if len(lines) != 5 {
			t.Fatal(lines)
		}
		for _, line := range lines {
			if gotui.StringWidth(line) > width || strings.ContainsAny(line, "\n\r\x1b") {
				t.Fatal(line)
			}
		}
	}
	c.closeWorkspaceIndex()
	if !reflect.DeepEqual(before, []any{c.input.Text(), c.input.cursorPos, c.input.undoText, c.input.undoCursor, c.input.yankText}) || c.transcriptScroll != 12 || c.stickToBottom || !reflect.DeepEqual(transcript, c.transcript) || c.workspaceIndexHeight() != 0 {
		t.Fatal("panel changed editor/reader/transcript")
	}
	var n int
	if err := c.store.DB().QueryRow("SELECT count(*) FROM turns").Scan(&n); err != nil || n != 0 {
		t.Fatal(n, err)
	}
}
func TestIndexPanelLateResultCannotEnterReopenedOrDifferentSession(t *testing.T) {
	c := indexPanelFixture(t)
	c.openWorkspaceIndex()
	old := <-c.workspaceIndex.results
	c.closeWorkspaceIndex()
	c.openWorkspaceIndex()
	c.handleWorkspaceIndexResult(old) // rejected and replaced by current read
	if c.workspaceIndex.status.State != "" {
		t.Fatal("old status accepted")
	}
	indexPanelResult(t, c)
	c.requestWorkspaceIndex(false)
	old = <-c.workspaceIndex.results
	c.closeWorkspaceIndex()
	c.sessionGeneration++
	c.handleWorkspaceIndexResult(old)
	if c.workspaceIndex.active {
		t.Fatal("late result reopened panel")
	}
}
func TestIndexPanelEscapeWhileBusyAndStopJoinsWork(t *testing.T) {
	c := indexPanelFixture(t)
	c.openWorkspaceIndex()
	indexPanelResult(t, c)
	cfg := c.workspaceIndex.configs["all"]
	peer, err := searchstore.NewRefreshStore(c.store.DB()).Begin(t.Context(), cfg, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Fail(t.Context(), errors.New("done"))
	c.requestWorkspaceIndex(true)
	c.closeWorkspaceIndex()
	c.stopWorkspaceIndex()
	if err := peer.Renew(context.Background(), time.Minute); err != nil {
		t.Fatal("closed another owner's lease", err)
	}
}
func TestIndexPanelErrorsAreOneLineAndKeysAreModal(t *testing.T) {
	c := indexPanelFixture(t)
	c.openWorkspaceIndex()
	indexPanelResult(t, c)
	c.workspaceIndex.err = errors.New("bad\x1b[31m\nerror 世界" + strings.Repeat("x", 200))
	if strings.ContainsAny(strings.Join(c.workspaceIndexLines(60), ""), "\x1b\n\r") {
		t.Fatal("control leaked")
	}
	keys := c.KeyMap()
	if len(keys) != 8 {
		t.Fatal("unexpected modal bindings", len(keys))
	}
	if !c.HandleMouse(gotui.MouseEvent{}) {
		t.Fatal("mouse escaped modal")
	}
}
