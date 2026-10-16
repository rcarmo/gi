package tui

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
)

// editorKeyEvents are the key events go-tui reads for the golden's legacy
// bytes (scripts/golden-editor-keys.mjs).
var editorKeyEvents = map[string]gotui.KeyEvent{
	"left":          {Key: gotui.KeyLeft},
	"right":         {Key: gotui.KeyRight},
	"ctrl+a":        {Key: gotui.KeyRune, Rune: 'a', Mod: gotui.ModCtrl},
	"ctrl+e":        {Key: gotui.KeyRune, Rune: 'e', Mod: gotui.ModCtrl},
	"ctrl+b":        {Key: gotui.KeyRune, Rune: 'b', Mod: gotui.ModCtrl},
	"ctrl+f":        {Key: gotui.KeyRune, Rune: 'f', Mod: gotui.ModCtrl},
	"alt+b":         {Key: gotui.KeyRune, Rune: 'b', Mod: gotui.ModAlt},
	"alt+f":         {Key: gotui.KeyRune, Rune: 'f', Mod: gotui.ModAlt},
	"alt+left":      {Key: gotui.KeyLeft, Mod: gotui.ModAlt},
	"alt+right":     {Key: gotui.KeyRight, Mod: gotui.ModAlt},
	"ctrl+left":     {Key: gotui.KeyLeft, Mod: gotui.ModCtrl},
	"ctrl+right":    {Key: gotui.KeyRight, Mod: gotui.ModCtrl},
	"backspace":     {Key: gotui.KeyBackspace},
	"ctrl+w":        {Key: gotui.KeyRune, Rune: 'w', Mod: gotui.ModCtrl},
	"alt+backspace": {Key: gotui.KeyBackspace, Mod: gotui.ModAlt},
	"alt+d":         {Key: gotui.KeyRune, Rune: 'd', Mod: gotui.ModAlt},
	"ctrl+u":        {Key: gotui.KeyRune, Rune: 'u', Mod: gotui.ModCtrl},
	"ctrl+k":        {Key: gotui.KeyRune, Rune: 'k', Mod: gotui.ModCtrl},
	"ctrl+y":        {Key: gotui.KeyRune, Rune: 'y', Mod: gotui.ModCtrl},
	"alt+y":         {Key: gotui.KeyRune, Rune: 'y', Mod: gotui.ModAlt},
	"ctrl+-":        {Key: gotui.KeyRune, Rune: '-', Mod: gotui.ModCtrl},
	"ctrl+]":        {Key: gotui.KeyRune, Rune: ']', Mod: gotui.ModCtrl},
	"ctrl+alt+]":    {Key: gotui.KeyRune, Rune: ']', Mod: gotui.ModCtrl | gotui.ModAlt},
}

// pressEditorKey runs the editor binding for ev: one with ev's exact
// modifiers, else the printable-rune binding for an unmodified rune.
func pressEditorKey(t *testing.T, m *multilineInput, ev gotui.KeyEvent) {
	t.Helper()
	keys := m.KeyMap()
	for _, b := range keys {
		p := b.Pattern
		if p.AnyRune || p.Mod != ev.Mod {
			continue
		}
		if (p.Rune != 0 && ev.Key == gotui.KeyRune && p.Rune == ev.Rune) || (p.Rune == 0 && p.Key != gotui.KeyRune && p.Key == ev.Key) {
			b.Handler(ev)
			return
		}
	}
	if ev.Key == gotui.KeyRune && ev.Mod == 0 {
		for _, b := range keys {
			if b.Pattern.AnyRune {
				b.Handler(ev)
				return
			}
		}
	}
	t.Fatalf("no editor binding for %+v", ev)
}

// The editor's word moves, kills, kill ring, yank-pop, undo and character
// jump match pi-tui's Editor key by key (bun scripts/golden-editor-keys.mjs).
func TestEditorKeysMatchPi(t *testing.T) {
	defer func(keys piKeys) { defaultPiKeys = keys }(defaultPiKeys)
	defaultPiKeys = piKeys{} // the golden's Linux keys
	raw, err := os.ReadFile("testdata/pi-editor-keys.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden map[string]struct {
		Steps  []json.RawMessage `json:"steps"`
		States []struct {
			Step   string `json:"step"`
			Text   string `json:"text"`
			Cursor int    `json:"cursor"`
		} `json:"states"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	for name, scenario := range golden {
		t.Run(name, func(t *testing.T) {
			m := newMultilineInput(80, "", func(string) {}, nil)
			m.Focus()
			for i, rawStep := range scenario.Steps {
				var key string
				var pair []string
				if json.Unmarshal(rawStep, &key) == nil {
					ev, ok := editorKeyEvents[key]
					if !ok {
						t.Fatalf("unknown key %q", key)
					}
					pressEditorKey(t, m, ev)
				} else if err := json.Unmarshal(rawStep, &pair); err == nil {
					switch pair[0] {
					case "set":
						m.SetText(pair[1])
					case "paste":
						m.paste(pair[1])
					case "type":
						for _, r := range pair[1] {
							pressEditorKey(t, m, gotui.KeyEvent{Key: gotui.KeyRune, Rune: r})
						}
					}
				} else {
					t.Fatalf("bad step %s", rawStep)
				}
				want := scenario.States[i]
				if m.text != want.Text || m.clampCursor() != want.Cursor {
					t.Fatalf("after step %d (%s): text %q cursor %d, Pi %q cursor %d", i, want.Step, m.text, m.clampCursor(), want.Text, want.Cursor)
				}
			}
		})
	}
}

func TestEditorVerticalNavigationWrappedUnicode(t *testing.T) {
	m := newMultilineInput(12, "", nil, nil)
	m.width = 12
	m.SetText("first 中文🙂 e\u0301\n" + strings.Repeat("x", 40) + "\nlast")
	original := m.Text()
	steps := 0
	for m.moveVertical(-1) {
		steps++
		if steps > 100 {
			t.Fatal("Up stuck on a soft-wrap boundary")
		}
	}
	m.moveHome()
	if m.cursorPos != 0 || steps < 3 {
		t.Fatalf("home=%d steps=%d", m.cursorPos, steps)
	}
	steps = 0
	for m.moveVertical(1) {
		steps++
		if steps > 100 {
			t.Fatal("Down stuck on a soft-wrap boundary")
		}
	}
	m.moveEnd()
	if m.cursorPos != len([]rune(original)) || m.Text() != original {
		t.Fatalf("navigation mutated text or lost final line: cursor=%d", m.cursorPos)
	}
}
