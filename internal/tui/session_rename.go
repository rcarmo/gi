package tui

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/store"
)

func (c *chatTUI) openSessionRename() {
	sess, err := c.store.GetSession(context.Background(), c.sessionActions.target)
	if err != nil {
		c.modelMenuError = "session unavailable; Esc back"
		if c.app != nil {
			c.app.MarkDirty()
		}
		return
	}
	input := newMultilineInput(c.currentContentWidth(), "", func(string) { c.saveSessionRename() }, func(string) { c.modelMenuError = "" })
	input.SetText(sess.Title)
	input.Focus()
	c.sessionActions.renameInput = input
	c.modelMenuKind = "session-rename"
	c.modelMenuError = ""
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) saveSessionRename() {
	if !c.ownsScope(c.modelMenuSession) {
		c.modelMenuError = "session changed; Esc cancel"
		return
	}
	input := c.sessionActions.renameInput
	if input == nil {
		return
	}
	err := c.store.MutateSession(context.Background(), c.sessionActions.target, store.SessionMutation{Action: "rename", Title: input.Text()})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrSessionMutationInvalid):
			c.modelMenuError = "title must be 1–160 characters on one line"
		case errors.Is(err, store.ErrSessionMutationConflict):
			c.modelMenuError = "session archived; Esc back, restore before renaming"
		default:
			c.modelMenuError = "storage error; not confirmed, Esc back to check"
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
		return
	}
	c.finishSessionMutation()
}

// Reuse editor movement/undo/yank semantics, but never focus or edit the
// composer. All bindings are preemptive while this transient prompt owns keys.
func (c *chatTUI) sessionRenameKeys() gotui.KeyMap {
	m := c.sessionActions.renameInput
	if m == nil {
		return nil
	}
	out := gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyEscape, func(gotui.KeyEvent) { c.backFromSessionActions() }),
		gotui.OnPreemptStop(gotui.KeyCtrlC, func(gotui.KeyEvent) { c.backFromSessionActions() }),
	}
	for _, binding := range m.KeyMap() {
		if binding.Pattern.Key == gotui.KeyEscape && binding.Pattern.Rune == 0 && !binding.Pattern.AnyRune {
			continue
		}
		// Every Enter variant submits rather than inserting a second line. Ctrl-J
		// is ignored so pasted multiline text cannot submit through that alias.
		if binding.Pattern.Key == gotui.KeyEnter {
			binding.Handler = func(gotui.KeyEvent) { c.saveSessionRename() }
		}
		if binding.Pattern.Rune == 'j' && binding.Pattern.Mod == gotui.ModCtrl {
			binding.Handler = func(gotui.KeyEvent) {}
		}
		handler := binding.Handler
		binding.Pattern.FocusRequired = false
		binding.Stop, binding.Preempt = true, true
		binding.Handler = func(event gotui.KeyEvent) {
			if event.Key == gotui.KeyRune && event.Mod&(gotui.ModCtrl|gotui.ModAlt) == 0 && (unicode.IsControl(event.Rune) || utf8.RuneCountInString(m.Text()) >= 160) {
				return
			}
			handler(event)
			if c.app != nil {
				c.app.MarkDirty()
			}
		}
		out = append(out, binding)
	}
	return out
}

// Horizontal caret window occupies one row, even for long/wide titles.
// Controls from externally set titles are visible spaces, never terminal escapes.
func sessionRenameLine(m *multilineInput, width int) string {
	width = max(1, width)
	runes := []rune(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, m.Text()))
	pos := min(max(0, m.cursorPos), len(runes))
	start := 0
	for start < pos && gotui.StringWidth(string(runes[start:pos])+"▌") > width {
		start++
	}
	line := string(runes[start:pos]) + "▌"
	for end := pos; end < len(runes); end++ {
		next := line + string(runes[end])
		if gotui.StringWidth(next) > width {
			break
		}
		line = next
	}
	return line
}

func (c *chatTUI) renderSessionRename(width int) *gotui.Element {
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithHeight(4))
	add := func(text string, style gotui.Style) {
		root.AddChild(gotui.New(gotui.WithHeight(1), gotui.WithWidthPercent(100), gotui.WithWrap(false), gotui.WithText(selectorText(text, width)), gotui.WithTextStyle(style)))
	}
	add("Rename session · Enter save · Esc cancel", gotui.NewStyle().Bold())
	add("target: "+c.sessionActions.targetLabel, gotui.NewStyle().Dim())
	if m := c.sessionActions.renameInput; m != nil {
		add(sessionRenameLine(m, width), gotui.NewStyle())
	}
	notice := "1–160 characters · Ctrl-A/E · Ctrl-U/K · Ctrl-Z undo"
	if c.modelMenuError != "" {
		notice = "error: " + c.modelMenuError
	}
	add(notice, gotui.NewStyle().Dim())
	return root
}
