package tui

import (
	"strings"
	"time"
	"unicode/utf8"

	gotui "github.com/grindlemire/go-tui"
)

type multilineInput struct {
	app              *gotui.App
	width            int
	border           gotui.BorderStyle
	placeholder      string
	placeholderStyle gotui.Style
	textStyle        gotui.Style
	cursorRune       rune
	autoFocus        bool
	onSubmit         func(string)
	onRestoreQueued  func()
	onComplete       func(string, int) (string, int, bool)
	onChange         func(string)
	text             string
	cursorPos        int
	undoText         string
	undoCursor       int
	hasUndo          bool
	yankText         string
	focused          bool
	blink            bool
}

func newMultilineInput(width int, placeholder string, onSubmit func(string), onChange func(string)) *multilineInput {
	return &multilineInput{
		width:            width,
		border:           gotui.BorderNone,
		placeholder:      placeholder,
		placeholderStyle: gotui.NewStyle().Dim(),
		textStyle:        gotui.NewStyle(),
		cursorRune:       '▌',
		autoFocus:        true,
		onSubmit:         onSubmit,
		onChange:         onChange,
		blink:            true,
	}
}

func (m *multilineInput) BindApp(app *gotui.App) { m.app = app }
func (m *multilineInput) Text() string           { return m.text }
func (m *multilineInput) SetText(s string) {
	m.text = s
	m.cursorPos = utf8.RuneCountInString(s)
	m.notifyChanged()
}
func (m *multilineInput) Clear()            { m.SetText("") }
func (m *multilineInput) IsFocusable() bool { return true }
func (m *multilineInput) IsTabStop() bool   { return true }
func (m *multilineInput) IsFocused() bool   { return m.focused }
func (m *multilineInput) Focus() {
	m.focused = true
	m.blink = true
}
func (m *multilineInput) Blur() { m.focused = false }

func (m *multilineInput) Watchers() []gotui.Watcher {
	return []gotui.Watcher{
		gotui.OnTimer(500*time.Millisecond, func() {
			if !m.focused {
				return
			}
			m.blink = !m.blink
			if m.app != nil {
				m.app.MarkDirty()
			}
		}),
	}
}

func (m *multilineInput) KeyMap() gotui.KeyMap {
	return gotui.KeyMap{
		gotui.OnFocused(gotui.AnyRune, m.insertRune),
		gotui.OnFocused(gotui.KeyBackspace, func(ke gotui.KeyEvent) { m.backspace() }),
		gotui.OnFocused(gotui.KeyDelete, func(ke gotui.KeyEvent) { m.delete() }),
		gotui.OnFocused(gotui.KeyLeft, func(ke gotui.KeyEvent) { m.moveLeft() }),
		gotui.OnFocused(gotui.KeyLeft.Alt(), func(ke gotui.KeyEvent) { m.moveWordLeft() }),
		gotui.OnFocused(gotui.KeyRight, func(ke gotui.KeyEvent) { m.moveRight() }),
		gotui.OnFocused(gotui.KeyRight.Alt(), func(ke gotui.KeyEvent) { m.moveWordRight() }),
		gotui.OnFocused(gotui.KeyHome, func(ke gotui.KeyEvent) { m.moveHome() }),
		gotui.OnFocused(gotui.KeyEnd, func(ke gotui.KeyEvent) { m.moveEnd() }),
		gotui.OnFocused(gotui.KeyCtrlA, func(ke gotui.KeyEvent) { m.moveHome() }),
		gotui.OnFocused(gotui.KeyCtrlE, func(ke gotui.KeyEvent) { m.moveEnd() }),
		gotui.OnFocused(gotui.KeyCtrlU, func(ke gotui.KeyEvent) { m.deleteToLineStart() }),
		gotui.OnFocused(gotui.KeyCtrlK, func(ke gotui.KeyEvent) { m.deleteToLineEnd() }),
		gotui.OnFocused(gotui.KeyCtrlW, func(ke gotui.KeyEvent) { m.deleteWordBackward() }),
		gotui.OnFocused(gotui.KeyBackspace.Alt(), func(ke gotui.KeyEvent) { m.deleteWordBackward() }),
		gotui.OnFocused(gotui.KeyDelete.Alt(), func(ke gotui.KeyEvent) { m.deleteWordForward() }),
		gotui.OnFocused(gotui.KeyCtrlZ, func(ke gotui.KeyEvent) { m.undo() }),
		gotui.OnFocused(gotui.KeyCtrlY, func(ke gotui.KeyEvent) { m.yank() }),
		gotui.OnFocused(gotui.KeyTab, func(ke gotui.KeyEvent) { m.complete() }),
		gotui.OnFocused(gotui.KeyEnter, m.enter),
		gotui.OnFocused(gotui.KeyEnter.Alt(), m.enter),
		gotui.OnFocused(gotui.KeyUp.Alt(), func(ke gotui.KeyEvent) {
			if m.onRestoreQueued != nil {
				m.onRestoreQueued()
			}
		}),
		gotui.OnFocused(gotui.KeyEscape, func(ke gotui.KeyEvent) {
			if app := ke.App(); app != nil {
				app.BlurFocused()
			}
		}),
	}
}

func (m *multilineInput) Render(app *gotui.App) *gotui.Element {
	lines := m.renderLines()
	help := m.helpLine()
	if help != "" {
		lines = append(lines, renderedLine{text: help, placeholder: true})
	}
	totalHeight := len(lines)
	if totalHeight < 1 {
		totalHeight = 1
	}
	if m.border != gotui.BorderNone {
		totalHeight += 2
	}
	root := gotui.New(
		gotui.WithDirection(gotui.Row),
		gotui.WithWidth(m.width),
		gotui.WithHeight(totalHeight),
		gotui.WithFocusable(true),
		gotui.WithAutoFocus(m.autoFocus),
		gotui.WithBorder(m.border),
	)
	root.SetOnFocus(func(e *gotui.Element) { m.Focus() })
	root.SetOnBlur(func(e *gotui.Element) { m.Blur() })
	for _, line := range lines {
		style := m.textStyle
		if line.placeholder {
			style = m.placeholderStyle
		}
		root.AddChild(gotui.New(gotui.WithText(line.text), gotui.WithTextStyle(style)))
	}
	return root
}

type renderedLine struct {
	text        string
	placeholder bool
}

func (m *multilineInput) renderLines() []renderedLine {
	if m.text == "" && !m.focused && m.placeholder != "" {
		return []renderedLine{{text: "○ " + m.placeholder, placeholder: true}}
	}
	if m.text == "" && m.focused && m.placeholder != "" {
		prefix := "● "
		if m.blink {
			prefix += string(m.cursorRune)
		}
		return []renderedLine{{text: prefix + m.placeholder, placeholder: true}}
	}
	visibleWidth := m.width
	if m.border != gotui.BorderNone {
		visibleWidth -= 2
	}
	if visibleWidth < 1 {
		visibleWidth = 1
	}
	runes := []rune(m.text)
	cursorPos := m.cursorPos
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}
	if m.focused && m.blink {
		runes = append(runes[:cursorPos], append([]rune{m.cursorRune}, runes[cursorPos:]...)...)
	}
	segments := strings.Split(string(runes), "\n")
	lines := make([]renderedLine, 0, len(segments))
	for _, segment := range segments {
		segmentRunes := []rune(segment)
		if len(segmentRunes) == 0 {
			lines = append(lines, renderedLine{text: ""})
			continue
		}
		for len(segmentRunes) > visibleWidth {
			lines = append(lines, renderedLine{text: string(segmentRunes[:visibleWidth])})
			segmentRunes = segmentRunes[visibleWidth:]
		}
		lines = append(lines, renderedLine{text: string(segmentRunes)})
	}
	if len(lines) == 0 {
		lines = []renderedLine{{text: ""}}
	}
	return lines
}

func (m *multilineInput) helpLine() string {
	if !m.focused {
		return ""
	}
	if strings.TrimSpace(m.text) == "" {
		return ""
	}
	return "enter send · shift-enter newline"
}

func (m *multilineInput) insertRune(ke gotui.KeyEvent) {
	m.snapshotUndo()
	runes := []rune(m.text)
	pos := m.clampCursor()
	runes = append(runes[:pos], append([]rune{ke.Rune}, runes[pos:]...)...)
	m.text = string(runes)
	m.cursorPos = pos + 1
	m.notifyChanged()
}

func (m *multilineInput) backspace() {
	runes := []rune(m.text)
	pos := m.clampCursor()
	if pos == 0 || len(runes) == 0 {
		return
	}
	m.snapshotUndo()
	m.yankText = string(runes[pos-1 : pos])
	runes = append(runes[:pos-1], runes[pos:]...)
	m.text = string(runes)
	m.cursorPos = pos - 1
	m.notifyChanged()
}

func (m *multilineInput) delete() {
	runes := []rune(m.text)
	pos := m.clampCursor()
	if pos >= len(runes) {
		return
	}
	m.snapshotUndo()
	m.yankText = string(runes[pos : pos+1])
	runes = append(runes[:pos], runes[pos+1:]...)
	m.text = string(runes)
	m.notifyChanged()
}

func (m *multilineInput) moveLeft() {
	if m.cursorPos > 0 {
		m.cursorPos--
		m.markDirty()
	}
}
func (m *multilineInput) moveRight() {
	if m.cursorPos < utf8.RuneCountInString(m.text) {
		m.cursorPos++
		m.markDirty()
	}
}
func (m *multilineInput) moveHome() { m.cursorPos = 0; m.markDirty() }
func (m *multilineInput) moveEnd()  { m.cursorPos = utf8.RuneCountInString(m.text); m.markDirty() }

func (m *multilineInput) moveWordLeft() {
	runes := []rune(m.text)
	pos := m.clampCursor()
	for pos > 0 && isWordSpace(runes[pos-1]) {
		pos--
	}
	for pos > 0 && !isWordSpace(runes[pos-1]) {
		pos--
	}
	m.cursorPos = pos
	m.markDirty()
}

func (m *multilineInput) moveWordRight() {
	runes := []rune(m.text)
	pos := m.clampCursor()
	for pos < len(runes) && !isWordSpace(runes[pos]) {
		pos++
	}
	for pos < len(runes) && isWordSpace(runes[pos]) {
		pos++
	}
	m.cursorPos = pos
	m.markDirty()
}

func (m *multilineInput) deleteWordBackward() {
	runes := []rune(m.text)
	end := m.clampCursor()
	start := end
	for start > 0 && isWordSpace(runes[start-1]) {
		start--
	}
	for start > 0 && !isWordSpace(runes[start-1]) {
		start--
	}
	if start == end {
		return
	}
	m.snapshotUndo()
	m.yankText = string(runes[start:end])
	m.text = string(append(runes[:start], runes[end:]...))
	m.cursorPos = start
	m.notifyChanged()
}

func (m *multilineInput) deleteWordForward() {
	runes := []rune(m.text)
	start := m.clampCursor()
	end := start
	for end < len(runes) && isWordSpace(runes[end]) {
		end++
	}
	for end < len(runes) && !isWordSpace(runes[end]) {
		end++
	}
	if start == end {
		return
	}
	m.snapshotUndo()
	m.yankText = string(runes[start:end])
	m.text = string(append(runes[:start], runes[end:]...))
	m.notifyChanged()
}

func (m *multilineInput) deleteToLineStart() {
	runes := []rune(m.text)
	pos := m.clampCursor()
	if pos == 0 {
		return
	}
	m.snapshotUndo()
	m.yankText = string(runes[:pos])
	m.text = string(runes[pos:])
	m.cursorPos = 0
	m.notifyChanged()
}

func (m *multilineInput) deleteToLineEnd() {
	runes := []rune(m.text)
	pos := m.clampCursor()
	if pos >= len(runes) {
		return
	}
	m.snapshotUndo()
	m.yankText = string(runes[pos:])
	m.text = string(runes[:pos])
	m.notifyChanged()
}

func isWordSpace(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }

func (m *multilineInput) complete() {
	if m.onComplete == nil {
		return
	}
	text, cursor, ok := m.onComplete(m.text, m.clampCursor())
	if !ok {
		return
	}
	m.snapshotUndo()
	m.text = text
	m.cursorPos = cursor
	m.notifyChanged()
}

func (m *multilineInput) enter(ke gotui.KeyEvent) {
	if ke.Mod&gotui.ModShift != 0 {
		m.insertLiteral('\n')
		return
	}
	if m.onSubmit != nil {
		m.onSubmit(m.text)
	}
}

func (m *multilineInput) insertLiteral(r rune) {
	m.snapshotUndo()
	runes := []rune(m.text)
	pos := m.clampCursor()
	runes = append(runes[:pos], append([]rune{r}, runes[pos:]...)...)
	m.text = string(runes)
	m.cursorPos = pos + 1
	m.notifyChanged()
}

func (m *multilineInput) snapshotUndo() {
	m.undoText = m.text
	m.undoCursor = m.clampCursor()
	m.hasUndo = true
}

func (m *multilineInput) undo() {
	if !m.hasUndo {
		return
	}
	m.text, m.undoText = m.undoText, m.text
	m.cursorPos, m.undoCursor = m.undoCursor, m.clampCursor()
	m.notifyChanged()
}

func (m *multilineInput) yank() {
	if m.yankText == "" {
		return
	}
	m.snapshotUndo()
	runes := []rune(m.text)
	pos := m.clampCursor()
	yankRunes := []rune(m.yankText)
	runes = append(runes[:pos], append(yankRunes, runes[pos:]...)...)
	m.text = string(runes)
	m.cursorPos = pos + len(yankRunes)
	m.notifyChanged()
}

func (m *multilineInput) clampCursor() int {
	count := utf8.RuneCountInString(m.text)
	if m.cursorPos < 0 {
		return 0
	}
	if m.cursorPos > count {
		return count
	}
	return m.cursorPos
}

func (m *multilineInput) notifyChanged() {
	if m.onChange != nil {
		m.onChange(m.text)
	}
	m.markDirty()
}

func (m *multilineInput) markDirty() {
	if m.app != nil {
		m.app.MarkDirty()
	}
}
