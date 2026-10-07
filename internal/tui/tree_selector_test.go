package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	gotui "github.com/grindlemire/go-tui"
)

var treeKeyEvents = map[string]gotui.KeyEvent{
	"up": {Key: gotui.KeyUp}, "down": {Key: gotui.KeyDown}, "left": {Key: gotui.KeyLeft}, "right": {Key: gotui.KeyRight},
	"pgup": {Key: gotui.KeyPageUp}, "pgdn": {Key: gotui.KeyPageDown}, "enter": {Key: gotui.KeyEnter}, "esc": {Key: gotui.KeyEscape},
	"backspace":  {Key: gotui.KeyBackspace},
	"ctrl+left":  {Key: gotui.KeyLeft, Mod: gotui.ModCtrl},
	"ctrl+right": {Key: gotui.KeyRight, Mod: gotui.ModCtrl},
	"alt+left":   {Key: gotui.KeyLeft, Mod: gotui.ModAlt},
	"ctrl+d":     {Key: gotui.KeyRune, Rune: 'd', Mod: gotui.ModCtrl},
	"ctrl+t":     {Key: gotui.KeyRune, Rune: 't', Mod: gotui.ModCtrl},
	"ctrl+u":     {Key: gotui.KeyRune, Rune: 'u', Mod: gotui.ModCtrl},
	"ctrl+l":     {Key: gotui.KeyRune, Rune: 'l', Mod: gotui.ModCtrl},
	"ctrl+a":     {Key: gotui.KeyRune, Rune: 'a', Mod: gotui.ModCtrl},
	"ctrl+o":     {Key: gotui.KeyRune, Rune: 'o', Mod: gotui.ModCtrl},
	"ctrl+x":     {Key: gotui.KeyRune, Rune: 'x', Mod: gotui.ModCtrl},
	"L":          {Key: gotui.KeyRune, Rune: 'L'},
	"T":          {Key: gotui.KeyRune, Rune: 'T'},
}

// treeFromPi builds the tree as the golden script's buildTree does.
func treeFromPi(entries []map[string]any, labels map[string]string) []*treeNode {
	nodes := map[string]*treeNode{}
	var order []*treeNode
	for _, e := range entries {
		n := &treeNode{entry: treeEntryFromPi(e), label: labels[e["id"].(string)]}
		nodes[n.entry.id] = n
		order = append(order, n)
	}
	var roots []*treeNode
	for _, n := range order {
		if p := nodes[n.entry.parentID]; p != nil {
			p.children = append(p.children, n)
		} else {
			roots = append(roots, n)
		}
	}
	return roots
}

// /tree renders, navigates, filters, searches, labels and copies as Pi's
// TreeSelectorComponent does, key by key (bun scripts/golden-tree.mjs).
func TestTreeSelectorMatchesPi(t *testing.T) {
	raw, err := os.ReadFile("testdata/pi-tree.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Entries   []map[string]any  `json:"entries"`
		Labels    map[string]string `json:"labels"`
		Scenarios map[string]struct {
			Leaf  string            `json:"leaf"`
			Steps []json.RawMessage `json:"steps"`
			Sizes map[string][]struct {
				Step      string    `json:"step"`
				Rows      []string  `json:"rows"`
				Selected  *string   `json:"selected"`
				Cancelled bool      `json:"cancelled"`
				Labelled  []*string `json:"labelled"`
				Copied    *string   `json:"copied"`
			} `json:"sizes"`
		} `json:"scenarios"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", "/home/u") // the golden's home, for ~ paths
	now, _ := time.Parse(time.RFC3339, "2026-10-03T21:12:00Z")
	str := func(p *string) string {
		if p == nil {
			return "<nil>"
		}
		return *p
	}
	for name, scenario := range golden.Scenarios {
		for size, states := range scenario.Sizes {
			t.Run(name+"@"+size, func(t *testing.T) {
				var width, height int
				fmt.Sscanf(size, "%dx%d", &width, &height)
				s := newTreeSelector(treeFromPi(golden.Entries, golden.Labels), scenario.Leaf, height, "", "")
				s.list.now = func() time.Time { return now }
				var selected, copied *string
				var labelled []*string
				cancelled := false
				s.onSelect = func(id string) { selected = &id }
				s.onCancel = func() { cancelled = true }
				s.onLabel = func(id, label string) {
					labelled = []*string{&id, nil}
					if label != "" {
						labelled[1] = &label
					}
				}
				s.onCopy = func(text string) {
					copied = nil
					if text != "" {
						copied = &text
					}
				}
				keys := treeSelectorKeys(s, func() {})
				for i, state := range states {
					selected, copied, labelled = nil, nil, nil
					if i > 0 {
						var key string
						var pair []string
						if json.Unmarshal(scenario.Steps[i-1], &key) == nil {
							ev, ok := treeKeyEvents[key]
							if !ok {
								t.Fatalf("no key event for %q", key)
							}
							pressMenuKey(t, keys, ev)
						} else if json.Unmarshal(scenario.Steps[i-1], &pair) == nil {
							for _, r := range pair[1] {
								pressMenuKey(t, keys, gotui.KeyEvent{Key: gotui.KeyRune, Rune: r})
							}
						}
					}
					if got := spanRowsText(s.list.selectorRows(width)); strings.Join(got, "\n") != strings.Join(state.Rows, "\n") {
						t.Fatalf("after %s:\n%s\nPi:\n%s", state.Step, strings.Join(got, "\n"), strings.Join(state.Rows, "\n"))
					}
					if str(selected) != str(state.Selected) || cancelled != state.Cancelled {
						t.Fatalf("after %s: selected %s cancelled %v, Pi %s %v", state.Step, str(selected), cancelled, str(state.Selected), state.Cancelled)
					}
					if fmt.Sprint(len(labelled)) != fmt.Sprint(len(state.Labelled)) || len(labelled) == 2 && (str(labelled[0]) != str(state.Labelled[0]) || str(labelled[1]) != str(state.Labelled[1])) {
						t.Fatalf("after %s: labelled %v, Pi %v", state.Step, labelled, state.Labelled)
					}
					if state.Step == "ctrl+x" && str(copied) != str(state.Copied) {
						t.Fatalf("after %s: copied %q, Pi %q", state.Step, str(copied), str(state.Copied))
					}
				}
			})
		}
	}
}
