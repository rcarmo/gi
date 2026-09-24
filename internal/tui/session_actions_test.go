package tui

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func selectActionSession(t *testing.T, c *chatTUI, id string) {
	t.Helper()
	c.openSessionMenu()
	for i, label := range c.modelMenuChoices {
		if c.modelMenuValues[label] == id {
			c.modelMenuSelected = i
			return
		}
	}
	t.Fatal("missing target", id)
}
func chooseSessionAction(t *testing.T, c *chatTUI, label string) {
	t.Helper()
	for i, v := range c.modelMenuChoices {
		if v == label {
			c.modelMenuSelected = i
			c.acceptModelMenuSelection()
			return
		}
	}
	t.Fatal("missing action", label, c.modelMenuChoices)
}

func TestSessionActionsCapabilitiesAndDraftOwnership(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	if _, err := c.store.CloneSession(ctx, "A", "child", "Child", "child"); err != nil {
		t.Fatal(err)
	}
	c.ensureInput()
	c.input.SetText("draft 中文\nsecond")
	c.input.cursorPos = 4
	c.input.undoText = "undo"
	c.input.hasUndo = true
	c.transcriptScroll = 3
	c.stickToBottom = false
	selectActionSession(t, c, "A")
	c.openSessionActions()
	if !reflect.DeepEqual(c.modelMenuChoices, []string{"Pin", "Rename"}) {
		t.Fatal("main capabilities", c.modelMenuChoices)
	}
	c.backFromSessionActions()
	c.closeModelMenu()
	selectActionSession(t, c, "child")
	selected := c.modelMenuSelected
	source := c.selectionScope()
	c.openSessionActions()
	if !reflect.DeepEqual(c.modelMenuChoices, []string{"Pin", "Archive", "Rename"}) {
		t.Fatal(c.modelMenuChoices)
	}
	c.modelMenuTypeRune('x')
	if c.modelMenuQuery != "" {
		t.Fatal("submenu stole filter input")
	}
	chooseSessionAction(t, c, "Pin")
	if c.modelMenuKind != "session" || c.modelMenuSelected != selected || c.selectionScope() != source {
		t.Fatal("action switched ownership")
	}
	if !strings.Contains(c.modelMenuChoices[selected], "pinned") {
		t.Fatal("label stale")
	}
	c.openSessionActions()
	chooseSessionAction(t, c, "Unpin")
	c.openSessionActions()
	chooseSessionAction(t, c, "Archive")
	if !strings.Contains(c.modelMenuChoices[selected], "archived") {
		t.Fatal("archive label stale")
	}
	c.openSessionActions()
	if !reflect.DeepEqual(c.modelMenuChoices, []string{"Restore"}) {
		t.Fatal("archived allowed edit", c.modelMenuChoices)
	}
	chooseSessionAction(t, c, "Restore")
	c.closeModelMenu()
	if c.sessionID != "A" || c.input.Text() != "draft 中文\nsecond" || c.input.cursorPos != 4 || c.input.undoText != "undo" || !c.input.hasUndo || c.transcriptScroll != 3 || c.stickToBottom {
		t.Fatal("draft/reader changed")
	}
	for _, id := range []string{"A", "child"} {
		turns, err := c.store.ListTurns(ctx, id)
		if err != nil || len(turns) != 0 {
			t.Fatal(turns, err)
		}
	}
}

func TestSessionActionsStaleScopeAndArchiveRace(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	if _, err := c.store.CloneSession(ctx, "A", "child", "Child", "child"); err != nil {
		t.Fatal(err)
	}
	selectActionSession(t, c, "child")
	c.openSessionActions()
	c.sessionGeneration++
	chooseSessionAction(t, c, "Pin")
	s, _ := c.store.GetSession(ctx, "child")
	if s.State["pinned"] == true || !strings.Contains(c.modelMenuError, "session changed") {
		t.Fatal("stale scope wrote")
	}
	c.closeModelMenu()
	selectActionSession(t, c, "child")
	c.openSessionActions()
	if _, err := c.store.CreateTurnWithStatus(ctx, "queued", "child", "queued", "later", nil); err != nil {
		t.Fatal(err)
	}
	chooseSessionAction(t, c, "Archive")
	s, _ = c.store.GetSession(ctx, "child")
	if s.State["archived_at"] != nil || c.modelMenuKind != "session" || c.modelMenuError == "" {
		t.Fatal("archive race accepted")
	}
	c.openSessionActions()
	if !reflect.DeepEqual(c.modelMenuChoices, []string{"Pin", "Rename"}) {
		t.Fatal("busy archive advertised")
	}
	c.closeModelMenu()
	// Remote archive after opening Pin must not permit editing an archived target.
	if err := c.store.DeleteTurn(ctx, "queued"); err != nil {
		t.Fatal(err)
	}
	selectActionSession(t, c, "child")
	c.openSessionActions()
	if err := c.store.MutateSession(ctx, "child", store.SessionMutation{Action: "archive"}); err != nil {
		t.Fatal(err)
	}
	chooseSessionAction(t, c, "Pin")
	if c.modelMenuError == "" {
		t.Fatal("concurrent archive ignored")
	}
}

func TestSessionActionsStorageFailureIsNotReportedAsConflict(t *testing.T) {
	c := sessionTestChat(t)
	selectActionSession(t, c, "A")
	c.openSessionActions()
	if err := c.store.Close(); err != nil {
		t.Fatal(err)
	}
	chooseSessionAction(t, c, "Pin")
	if c.modelMenuKind != "session" || !strings.Contains(c.modelMenuError, "storage error") || !strings.Contains(c.modelMenuError, "not confirmed") {
		t.Fatal(c.modelMenuError)
	}
}
