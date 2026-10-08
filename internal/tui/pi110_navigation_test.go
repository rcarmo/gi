package tui

import (
	"encoding/json"
	"os"
	"testing"

	gotui "github.com/grindlemire/go-tui"
)

// Fixture values come from the published pi-tui 1.1.0 keybinding definitions.
func TestPi110NavigationDefinitions(t *testing.T) {
	raw, err := os.ReadFile("testdata/pi-1.1.0-navigation.json")
	if err != nil {
		t.Fatal(err)
	}
	var bindings map[string][]string
	if err := json.Unmarshal(raw, &bindings); err != nil {
		t.Fatal(err)
	}
	for action, want := range map[string][]string{
		"tui.editor.cursorLineStart": {"home", "ctrl+a"},
		"tui.editor.cursorLineEnd":   {"end", "ctrl+e"},
		"tui.altScreen.top":          {"ctrl+home"},
		"tui.altScreen.bottom":       {"ctrl+end"},
	} {
		got := bindings[action]
		if len(got) != len(want) {
			t.Fatalf("%s: %v", action, got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: %v", action, got)
			}
		}
	}
}

func TestPi110EditorHomeEndAndViewportKeys(t *testing.T) {
	for _, fullscreen := range []bool{true, false} {
		m := newMultilineInput(40, "", nil, nil)
		m.SetText("first\n日本語 middle\nlast")
		m.cursorPos = 10
		calls := 0
		if fullscreen {
			m.onTranscriptTop = func() { calls++ }
			m.onTranscriptEnd = func() { calls++ }
		}
		pressEditorKey(t, m, gotui.KeyEvent{Key: gotui.KeyHome})
		if m.cursorPos != 6 || calls != 0 {
			t.Fatal("Home did not edit current line", m.cursorPos, calls)
		}
		pressEditorKey(t, m, gotui.KeyEvent{Key: gotui.KeyEnd})
		if m.cursorPos != 16 || calls != 0 {
			t.Fatal("End did not edit current line", m.cursorPos, calls)
		}
		for _, key := range []gotui.Key{gotui.KeyHome, gotui.KeyEnd} {
			pressEditorKey(t, m, gotui.KeyEvent{Key: key, Mod: gotui.ModCtrl})
			if m.cursorPos != 16 {
				t.Fatal("viewport key moved editor", fullscreen, m.cursorPos)
			}
		}
		want := 0
		if fullscreen {
			want = 2
		}
		if calls != want {
			t.Fatal("viewport callback count", calls, want)
		}
	}
}

func TestPi110FullscreenNavigationPrecedence(t *testing.T) {
	c, _ := scrolledTranscript(2)
	c.ensureInput()
	c.input.SetText("kept\ndraft")
	c.input.cursorPos = 3
	for _, b := range c.KeyMap() {
		if (b.Pattern.Key == gotui.KeyHome || b.Pattern.Key == gotui.KeyEnd) && b.Pattern.Mod == 0 {
			t.Fatal("chat preempts editor Home/End")
		}
	}
	pressAppKey(t, c, gotui.KeyEvent{Key: gotui.KeyHome, Mod: gotui.ModCtrl})
	if c.transcriptScroll != 0 || c.stickToBottom || c.input.cursorPos != 3 {
		t.Fatal("Ctrl+Home", c.transcriptScroll, c.input.cursorPos)
	}
	pressAppKey(t, c, gotui.KeyEvent{Key: gotui.KeyEnd, Mod: gotui.ModCtrl})
	if !c.stickToBottom || c.input.cursorPos != 3 || c.input.Text() != "kept\ndraft" {
		t.Fatal("Ctrl+End changed draft")
	}
}
