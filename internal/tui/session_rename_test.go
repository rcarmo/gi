package tui

import (
	"context"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/store"
)

func renameKey(t *testing.T, c *chatTUI, key gotui.Key, r rune, mod gotui.Modifier) {
	t.Helper()
	for _, b := range c.sessionRenameKeys() {
		p := b.Pattern
		if p.AnyKey || (key == gotui.KeyRune && (p.AnyRune || p.Rune != 0)) || (p.Rune == 0 && !p.AnyRune && p.Key == key) {
			if p.Rune != 0 && p.Rune != r {
				continue
			}
			if p.Mod != 0 && p.Mod != mod {
				continue
			}
			if p.ExcludeMods&mod != 0 {
				continue
			}
			if !b.Preempt || !b.Stop || p.FocusRequired {
				t.Fatal("rename binding not isolated")
			}
			b.Handler(gotui.KeyEvent{Key: key, Rune: r, Mod: mod})
			return
		}
	}
	t.Fatalf("unhandled key %v %q %v", key, r, mod)
}
func startRename(t *testing.T, c *chatTUI, id string) {
	t.Helper()
	selectActionSession(t, c, id)
	c.openSessionActions()
	chooseSessionAction(t, c, "Rename")
	if c.modelMenuKind != "session-rename" || c.modelMenuHeight() != 4 {
		t.Fatal("unbounded rename prompt")
	}
}

func TestSessionRenameSaveCancelValidationAndComposerIsolation(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	if _, err := c.store.CloneSession(ctx, "A", "child", "Original", "child"); err != nil {
		t.Fatal(err)
	}
	c.ensureInput()
	c.input.SetText("composer 中文\nsecond")
	c.input.cursorPos = 4
	c.input.undoText = "undo"
	c.input.hasUndo = true
	c.input.yankText = "yank"
	c.transcriptScroll = 3
	c.stickToBottom = false
	startRename(t, c, "child")
	m := c.sessionActions.renameInput
	if m.Text() != "Original" {
		t.Fatal(m.Text())
	}
	renameKey(t, c, gotui.KeyRune, 'a', gotui.ModCtrl)
	renameKey(t, c, gotui.KeyRune, 'k', gotui.ModCtrl)
	renameKey(t, c, gotui.KeyEnter, 0, 0)
	if !strings.Contains(c.modelMenuError, "1–160") || c.modelMenuKind != "session-rename" {
		t.Fatal("blank accepted")
	}
	for _, r := range "Title 界 e\u0301" {
		renameKey(t, c, gotui.KeyRune, r, 0)
	}
	if c.modelMenuError != "" {
		t.Fatal("edit retained stale validation error")
	}
	if !c.HandleMouse(gotui.MouseEvent{Action: gotui.MousePress, Button: gotui.MouseLeft, X: 1, Y: 1}) || c.input.cursorPos != 4 {
		t.Fatal("mouse escaped rename prompt")
	}
	renameKey(t, c, gotui.KeyLeft, 0, 0)
	renameKey(t, c, gotui.KeyRune, 'X', 0)
	renameKey(t, c, gotui.KeyRune, 'z', gotui.ModCtrl)
	if m.Text() != "Title 界 e\u0301" {
		t.Fatal("undo damaged unicode", m.Text())
	}
	renameKey(t, c, gotui.KeyRune, 'j', gotui.ModCtrl)
	if strings.Contains(m.Text(), "\n") {
		t.Fatal("multiline inserted")
	}
	renameKey(t, c, gotui.KeyEnter, 0, gotui.ModShift)
	s, _ := c.store.GetSession(ctx, "child")
	if s.Title != "Title 界 e\u0301" || c.modelMenuKind != "session" || c.sessionID != "A" {
		t.Fatal("save ownership", s.Title, c.modelMenuKind)
	}
	c.openSessionActions()
	chooseSessionAction(t, c, "Rename")
	c.sessionActions.renameInput.SetText("cancel me")
	renameKey(t, c, gotui.KeyEscape, 0, 0)
	if c.modelMenuKind != "session-actions" || c.sessionActions.renameInput != nil {
		t.Fatal("cancel state")
	}
	c.backFromSessionActions()
	c.closeModelMenu()
	s, _ = c.store.GetSession(ctx, "child")
	if s.Title != "Title 界 e\u0301" {
		t.Fatal("cancel wrote")
	}
	if c.input.Text() != "composer 中文\nsecond" || c.input.cursorPos != 4 || c.input.undoText != "undo" || c.input.yankText != "yank" || !c.input.hasUndo || c.transcriptScroll != 3 || c.stickToBottom {
		t.Fatal("composer or reader changed")
	}
}

func TestSessionRenameStaleArchiveAndStorageErrorsRetainInput(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	if _, err := c.store.CloneSession(ctx, "A", "child", "Original", "child"); err != nil {
		t.Fatal(err)
	}
	startRename(t, c, "child")
	m := c.sessionActions.renameInput
	m.SetText("new title")
	c.sessionGeneration++
	c.saveSessionRename()
	if !strings.Contains(c.modelMenuError, "session changed") || m.Text() != "new title" {
		t.Fatal("stale scope")
	}
	c.closeModelMenu()
	startRename(t, c, "child")
	m = c.sessionActions.renameInput
	m.SetText("new title")
	if err := c.store.MutateSession(ctx, "child", store.SessionMutation{Action: "archive"}); err != nil {
		t.Fatal(err)
	}
	c.saveSessionRename()
	s, _ := c.store.GetSession(ctx, "child")
	if !strings.Contains(c.modelMenuError, "archived") || s.Title != "Original" || m.Text() != "new title" {
		t.Fatal("archive race")
	}
	c.closeModelMenu()
	if err := c.store.MutateSession(ctx, "child", store.SessionMutation{Action: "restore"}); err != nil {
		t.Fatal(err)
	}
	startRename(t, c, "child")
	m = c.sessionActions.renameInput
	m.SetText("retain on failure")
	if err := c.store.Close(); err != nil {
		t.Fatal(err)
	}
	c.saveSessionRename()
	if !strings.Contains(c.modelMenuError, "not confirmed") || m.Text() != "retain on failure" || c.modelMenuKind != "session-rename" {
		t.Fatal("storage failure lost editor")
	}
}

func TestSessionRenameSingleRowWindowAndNativeLimit(t *testing.T) {
	c := sessionTestChat(t)
	startRename(t, c, "A")
	m := c.sessionActions.renameInput
	m.SetText(strings.Repeat("界e\u0301", 53) + "Z")
	for _, width := range []int{1, 8, 60, 100, 140} {
		for _, pos := range []int{0, 40, 100, len([]rune(m.Text()))} {
			m.cursorPos = pos
			line := sessionRenameLine(m, width)
			if gotui.StringWidth(line) > width || !strings.Contains(line, "▌") || strings.Contains(line, "\n") {
				t.Fatal(width, pos, line)
			}
		}
	}
	before := m.Text()
	renameKey(t, c, gotui.KeyRune, 'X', 0)
	if m.Text() != before {
		t.Fatal("typed length exceeded160")
	}
	m.SetText(strings.Repeat("a", 161))
	c.saveSessionRename()
	if !strings.Contains(c.modelMenuError, "1–160") {
		t.Fatal("native limit missing")
	}
	m.SetText("bad\nname")
	c.saveSessionRename()
	if !strings.Contains(c.modelMenuError, "one line") {
		t.Fatal("native single-line validation missing")
	}
}
