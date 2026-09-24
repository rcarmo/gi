package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func selectSessionLabel(t *testing.T, c *chatTUI, id string) {
	t.Helper()
	for index, label := range c.modelMenuChoices {
		if c.modelMenuValues[label] == id {
			c.modelMenuSelected = index
			return
		}
	}
	t.Fatalf("missing session %s", id)
}

func TestSessionPickerCapturedGenerationRejectsRetargeting(t *testing.T) {
	c := sessionTestChat(t)
	c.input.SetText("origin 中文🙂")
	c.input.cursorPos = 3
	c.openSessionMenu()
	selectSessionLabel(t, c, "B")
	c.sessionGeneration++ // A -> B -> A has the same ID but a different owner.
	c.acceptModelMenuSelection()
	if c.sessionID != "A" || !c.modelMenuOpen || c.modelMenuError != "session changed; reopen picker" {
		t.Fatal("stale selector accepted", c.sessionID, c.modelMenuError)
	}
	if c.input.Text() != "origin 中文🙂" || c.input.cursorPos != 3 {
		t.Fatal("stale selector changed editor")
	}
	c.closeModelMenu()
	c.openSessionMenu()
	if c.modelMenuError != "" {
		t.Fatal("old error leaked on reopen")
	}
	selectSessionLabel(t, c, "B")
	c.acceptModelMenuSelection()
	if c.sessionID != "B" || c.modelMenuOpen || !c.inputActive {
		t.Fatal("fresh selection failed")
	}
	c.openSessionMenu()
	selectSessionLabel(t, c, "A")
	c.acceptModelMenuSelection()
	if c.input.Text() != "origin 中文🙂" || c.input.cursorPos != 3 {
		t.Fatal("A editor lost")
	}
	if c.modelMenuSession != (sessionScope{}) || c.modelMenuHeight() != 0 {
		t.Fatal("retained owner or idle rows")
	}
	for _, id := range []string{"A", "B"} {
		turns, err := c.store.ListTurns(context.Background(), id)
		if err != nil || len(turns) != 0 {
			t.Fatal("selection submitted", turns, err)
		}
	}
}

func TestSessionPickerNativeFailureKeepsDraftAndAllowsRetry(t *testing.T) {
	c, path, _ := modelTestChat(t)
	c.input.SetText("unsent")
	c.input.cursorPos = 2
	c.openSessionMenu()
	selectSessionLabel(t, c, "B")
	if err := c.store.Close(); err != nil {
		t.Fatal(err)
	}
	c.acceptModelMenuSelection()
	if !c.modelMenuOpen || c.sessionID != "A" || !strings.Contains(c.modelMenuError, "session unavailable") {
		t.Fatal("failed read switched/closed", c.modelMenuError)
	}
	if c.input.Text() != "unsent" || c.input.cursorPos != 2 {
		t.Fatal("error lost editor")
	}
	reopened, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	c.store = reopened
	c.acceptModelMenuSelection()
	if c.modelMenuOpen || c.sessionID != "B" {
		t.Fatal("retry failed", c.modelMenuError)
	}
	c.switchSession("A")
	if c.input.Text() != "unsent" || c.input.cursorPos != 2 {
		t.Fatal("retry lost origin draft")
	}
}
