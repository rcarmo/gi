package tui

import (
	"context"
	"reflect"
	"testing"
)

func TestTUIUnselectedModelRetainsDraftBeforeAdmission(t *testing.T) {
	for _, busy := range []bool{false, true} {
		for _, text := range []string{"  unsent 中文🙂\nsecond line  ", "!echo guarded", "@peer directed"} {
			c := pendingMediaChat(t)
			ctx := context.Background()
			c.cfg.DefaultModel = ""
			c.running = busy
			c.input.SetText(text)
			c.input.cursorPos = 4
			c.input.undoText = "undo value"
			c.input.undoCursor = 2
			c.input.hasUndo = true
			c.input.yankText = "yank"
			c.history = []string{"prior"}
			c.histIdx = 0
			c.queuedDrafts = []string{"previous queued shortcut"}
			before := *c.input
			for i := 0; i < 2; i++ {
				c.onSubmit(text)
			}
			if c.input.text != before.text || c.input.cursorPos != before.cursorPos || c.input.undoText != before.undoText || c.input.undoCursor != before.undoCursor || c.input.hasUndo != before.hasUndo || c.input.yankText != before.yankText {
				t.Fatal("editor changed on rejection")
			}
			if !reflect.DeepEqual(c.history, []string{"prior"}) || c.histIdx != 0 || !reflect.DeepEqual(c.queuedDrafts, []string{"previous queued shortcut"}) || c.running != busy {
				t.Fatal("history/queue/activity mutated")
			}
			turns, err := c.store.ListTurns(ctx, "A")
			if err != nil || len(turns) != 0 {
				t.Fatal(turns, err)
			}
			var journal int
			if err = c.store.DB().QueryRow(`select count(*) from kv_store where namespace='tui_pending_media_v1' and key='A'`).Scan(&journal); err != nil || journal != 0 {
				t.Fatal("rejection wrote media state", journal, err)
			}
		}
	}
}

func TestTUIUnselectedModelLeavesLocalCommandsAvailable(t *testing.T) {
	c := pendingMediaChat(t)
	c.cfg.DefaultModel = ""
	c.input.SetText("/where")
	c.onSubmit("/where")
	if c.input.text != "" || len(c.history) != 1 || c.history[0] != "/where" {
		t.Fatal("slash command blocked")
	}
}
