package tui

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/store"
)

func historyKey(c *chatTUI, key gotui.Key) {
	for _, binding := range c.KeyMap() {
		if binding.Pattern.Key == key && binding.Pattern.Mod == 0 {
			binding.Handler(gotui.KeyEvent{Key: key})
			return
		}
	}
	panic("missing history key")
}

func TestCursorHistoryRoundTripAndSessionPersistence(t *testing.T) {
	c := sessionTestChat(t)
	for _, input := range []string{"first prompt", "second prompt", "third prompt"} {
		c.recordInputHistory(input)
	}
	c.input.SetText("draft 中文🙂")
	c.input.cursorPos = 5
	historyKey(c, gotui.KeyUp)
	if c.input.Text() != "draft 中文🙂" || c.input.cursorPos != 0 {
		t.Fatal("first Up must reach line start before history")
	}
	historyKey(c, gotui.KeyUp)
	if c.input.Text() != "third prompt" {
		t.Fatal("Up did not recall newest entry")
	}
	historyKey(c, gotui.KeyUp)
	historyKey(c, gotui.KeyUp)
	historyKey(c, gotui.KeyUp) // clamp to oldest
	if c.input.Text() != "first prompt" {
		t.Fatalf("oldest history: %q", c.input.Text())
	}
	historyKey(c, gotui.KeyDown)
	if c.input.Text() != "second prompt" {
		t.Fatal("Down did not advance")
	}
	historyKey(c, gotui.KeyDown)
	historyKey(c, gotui.KeyDown)
	if c.input.Text() != "draft 中文🙂" || c.input.cursorPos != 0 {
		t.Fatalf("draft lost: %q cursor %d", c.input.Text(), c.input.cursorPos)
	}
	historyKey(c, gotui.KeyDown)
	if c.input.Text() != "draft 中文🙂" {
		t.Fatal("Down without history changed draft")
	}
	historyKey(c, gotui.KeyUp) // Down at document end moved to its end; Up reaches start.
	historyKey(c, gotui.KeyUp)
	c.input.insertRune(gotui.KeyEvent{Rune: '!'})
	if c.histIdx != -1 {
		t.Fatal("editing recalled entry did not leave history navigation")
	}
	historyKey(c, gotui.KeyDown)
	if c.input.Text() != "third prompt!" {
		t.Fatalf("edited entry replaced: %q", c.input.Text())
	}
	c.input.SetText("A scratch")
	historyKey(c, gotui.KeyUp) // first line: move to start
	historyKey(c, gotui.KeyUp)
	if !c.switchSession("B") {
		t.Fatal("switch B")
	}
	if len(c.history) != 0 || c.input.Text() != "" {
		t.Fatal("A history leaked to B")
	}
	c.recordInputHistory("B prompt")
	historyKey(c, gotui.KeyUp)
	if c.input.Text() != "B prompt" {
		t.Fatal("B history missing")
	}
	if !c.switchSession("A") {
		t.Fatal("switch A")
	}
	if c.input.Text() != "third prompt" || c.historyDraft != "A scratch" {
		t.Fatalf("A navigation lost across switch: %q / %q", c.input.Text(), c.historyDraft)
	}
	historyKey(c, gotui.KeyDown)
	if c.input.Text() != "A scratch" {
		t.Fatal("A draft not restored after switch")
	}
	if !reflect.DeepEqual(c.loadCommandHistory(), []string{"first prompt", "second prompt", "third prompt"}) {
		t.Fatal("A history not persisted")
	}
	if !c.switchSession("B") {
		t.Fatal("switch B again")
	}
	if !reflect.DeepEqual(c.loadCommandHistory(), []string{"B prompt"}) {
		t.Fatal("B history not persisted")
	}
}

func TestHistoryReloadAndSearchCycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gi.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"first alpha", "other", "second alpha"} {
		if err := s.RecordTUIInput(ctx, "A", text, 10000); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, sessionID: "A"}
	c.ensureInput()
	c.historySearchIdx = -1
	c.history = c.loadCommandHistory()
	c.input.SetText("alpha")
	c.searchHistoryBackward()
	if c.input.Text() != "second alpha" {
		t.Fatal("search newest")
	}
	c.searchHistoryBackward()
	if c.input.Text() != "first alpha" {
		t.Fatal("Ctrl-R did not repeat search")
	}
	c.recallHistory(1)
	if c.input.Text() != "other" {
		t.Fatal("Down after search")
	}
	c.recallHistory(1)
	c.recallHistory(1)
	if c.input.Text() != "alpha" {
		t.Fatal("search draft not restored")
	}
}

func TestDurableHistoryDoesNotRecordRejectedRoute(t *testing.T) {
	c := durableTestChat(t)
	child, err := c.engine.ResolveOrCreatePeerSessionID(context.Background(), "A", "other")
	if err != nil {
		t.Fatal(err)
	}
	if !c.switchSession(child) {
		t.Fatal("switch peer")
	}
	// Plain text stays in the peer session; an @mention of another agent is
	// rejected by the durable composer and must not enter history.
	c.input.SetText("@agent unroutable peer prompt")
	c.onSubmit(c.input.Text())
	if len(c.history) != 0 {
		t.Fatalf("rejected prompt entered history: %v", c.history)
	}
	stored, err := c.store.ListTUIInputHistory(context.Background(), child, 10000)
	if err != nil || len(stored) != 0 {
		t.Fatalf("rejected prompt stored: %v %v", stored, err)
	}

}
