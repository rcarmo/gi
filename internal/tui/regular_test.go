package tui

import (
	gotui "github.com/grindlemire/go-tui"
	"strings"
	"testing"
)

func TestRegularModeValidationHasNoSideEffects(t *testing.T) {
	if err := RunMode("/does/not/exist/gi.db", "/missing", "", "unknown"); err == nil || !strings.Contains(err.Error(), "invalid tui mode") {
		t.Fatal(err)
	}
}
func TestRegularStableOutputWaitsForFinalDraftAndTool(t *testing.T) {
	c := &chatTUI{transcript: []string{"you: accepted", "partial draft"}, regularPrinted: 1, draftLineIndex: 1, draftLineCount: 1}
	if c.regularStableEnd() != 1 {
		t.Fatal("committed partial draft")
	}
	c.draftLineIndex = -1
	c.draftLineCount = 0
	if c.regularStableEnd() != 2 {
		t.Fatal("lost final draft")
	}
	c.transcript = append(c.transcript, encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "tool", Kind: "tool", Status: "running"}), "│ partial")
	if c.regularStableEnd() != 2 {
		t.Fatal("committed mutable tool")
	}
	c.transcript[2] = encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "tool", Kind: "tool", Status: "ok"})
	if c.regularStableEnd() != 4 {
		t.Fatal("lost complete tool")
	}
	c.regularPrinted = 4
	if len(c.regularPending()) != 0 {
		t.Fatal("redraw reprints history")
	}
}
func TestRegularScrollbackPruningDoesNotReprint(t *testing.T) {
	c := &chatTUI{regularPrinted: 3, transcript: []string{"a", "b", "c", "new"}, draftLineIndex: -1}
	c.cfg.ScrollbackLimit = 3
	c.applyTranscriptLimit()
	if c.regularPrinted != 2 || strings.Join(c.regularPending(), " ") != "new" {
		t.Fatal(c.regularPrinted, c.regularPending())
	}
}
func TestRegularEditorAndSelectorNavigation(t *testing.T) {
	c := &chatTUI{regularMode: true}
	c.ensureInput()
	if c.input.onTranscriptTop != nil || c.input.onTranscriptEnd != nil {
		t.Fatal("regular intercepted editor Home/End")
	}
	for _, b := range c.KeyMap() {
		switch b.Pattern.Key {
		case gotui.KeyHome, gotui.KeyEnd, gotui.KeyPageUp, gotui.KeyPageDown:
			t.Fatal("regular captured native history key")
		}
	}
	c.modelMenuOpen = true
	found := false
	for _, b := range c.KeyMap() {
		if b.Pattern.Key == gotui.KeyHome {
			found = true
		}
	}
	if !found {
		t.Fatal("selector lost navigation")
	}
	c.regularMode = false
	c.ensureInput()
	if c.input.onTranscriptTop == nil || c.input.onTranscriptEnd == nil {
		t.Fatal("fullscreen lost transcript Home/End")
	}
}
func TestRegularSessionSwitchResetsPrintedCursorAndPreservesEditor(t *testing.T) {
	c := sessionTestChat(t)
	c.regularMode = true
	c.ensureInput()
	c.regularPrinted = 99
	c.input.SetText("A unsent")
	c.input.cursorPos = 2
	c.switchSession("B")
	if c.regularPrinted != 0 || !c.regularSessionPending {
		t.Fatal("new session not marked")
	}
	c.input.SetText("B unsent")
	c.switchSession("A")
	if c.input.Text() != "A unsent" || c.input.cursorPos != 2 {
		t.Fatal("regular switch lost editor")
	}
}

func TestMultilineEditorRendersSeparateRowsAndBindsNewline(t *testing.T) {
	m := newMultilineInput(60, "", nil, nil)
	m.SetText("first line\nsecond line")
	root := m.Render(nil)
	buf := gotui.NewBuffer(60, 4)
	root.Render(buf, 60, 2)
	text := buf.StringTrimmed()
	if !strings.Contains(text, "first line\nsecond line") {
		t.Fatal("editor lines concatenated", text)
	}
	shift, ctrl := false, false
	for _, b := range m.KeyMap() {
		if b.Pattern.Key == gotui.KeyEnter && b.Pattern.Mod == gotui.ModShift {
			shift = true
		}
		if b.Pattern.Rune == 'j' && b.Pattern.Mod == gotui.ModCtrl {
			ctrl = true
		}
	}
	if !shift || !ctrl {
		t.Fatal("missing explicit newline binding")
	}
}
