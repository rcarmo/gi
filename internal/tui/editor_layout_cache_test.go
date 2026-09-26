package tui

import (
	"reflect"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
)

func TestEditorLayoutCacheReuseAndCompleteInvalidation(t *testing.T) {
	m := newMultilineInput(12, "", nil, nil)
	m.maxLines = 3
	m.Focus()
	m.SetText(strings.Repeat("draft 中文🙂e\u0301\n", 30))
	m.renderLines()
	initial := m.layoutCache
	if initial == nil {
		t.Fatal("cache absent")
	}
	m.blink = !m.blink
	m.textStyle = gotui.NewStyle().Bold()
	m.maxLines = 2
	m.scrollRow = 0
	m.renderLines()
	if m.layoutCache != initial {
		t.Fatal("presentation-only state rebuilt layout")
	}
	fresh := *m
	fresh.layoutCache = nil
	if !reflect.DeepEqual(m.renderLines(), fresh.renderLines()) {
		t.Fatal("cache differs from fresh viewport")
	}
	for _, change := range []func(){
		func() { m.cursorPos-- }, func() { m.width-- }, func() { m.border = gotui.BorderRounded }, func() { m.Blur() }, func() { m.Focus() }, func() { m.cursorRune = '|' }, func() { m.text += "different" },
	} {
		previous := m.layoutCache
		change()
		actual := m.renderLines()
		if m.layoutCache == previous {
			t.Fatal("layout input did not invalidate")
		}
		fresh := *m
		fresh.layoutCache = nil
		if !reflect.DeepEqual(actual, fresh.renderLines()) {
			t.Fatal("stale cached layout")
		}
	}
	m.SetText("")
	m.renderLines()
	if m.layoutCache != nil || m.scrollRow != 0 {
		t.Fatal("clear retains large cache")
	}
}

func TestEditorLayoutCacheSnapshotsCannotOverwriteEachOther(t *testing.T) {
	m := newMultilineInput(20, "", nil, nil)
	m.maxLines = 2
	m.Focus()
	m.SetText(strings.Repeat("immutable snapshot\n", 8))
	m.renderLines()
	original := m.layoutCache
	lines := append([]renderedLine(nil), original.lines...)
	clone := *m
	clone.width = 5
	clone.cursorPos = 2
	clone.renderLines()
	if m.layoutCache != original || clone.layoutCache == original || !reflect.DeepEqual(original.lines, lines) {
		t.Fatal("layout snapshot mutated")
	}
	// Callers may append a help row, but cannot overwrite another cached line.
	visible := m.renderLines()
	if len(visible) != cap(visible) {
		t.Fatal("mutable tail capacity")
	}
	_ = append(visible, renderedLine{text: "extra"})
	if !reflect.DeepEqual(original.lines, lines) {
		t.Fatal("append corrupted cached rows")
	}
	other := newMultilineInput(20, "", nil, nil)
	other.maxLines = 2
	other.Focus()
	other.SetText("search")
	other.renderLines()
	if other.layoutCache == original || m.Text() == other.Text() {
		t.Fatal("shared editor/search state")
	}
}

func TestEditorCachedLayoutRedrawDoesNotAllocate(t *testing.T) {
	m := newMultilineInput(80, "", nil, nil)
	m.maxLines = 6
	m.Focus()
	m.SetText(strings.Repeat("large 中文🙂 draft\n", 6000))
	m.renderLines()
	allocations := testing.AllocsPerRun(100, func() { m.renderLines() })
	if allocations != 0 {
		t.Fatalf("unchanged layout allocations=%v", allocations)
	}
}
