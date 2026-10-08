package tui

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/graphemes"
	gotui "github.com/grindlemire/go-tui"
)

type multilineInput struct {
	app              *gotui.App
	width            int
	maxLines         int // zero leaves standalone inputs unbounded
	scrollRow        int
	totalRows        int
	layoutCache      *editorLayoutCache // immutable, one entry per input snapshot
	border           gotui.BorderStyle
	placeholder      string
	placeholderStyle gotui.Style
	textStyle        gotui.Style
	autoFocus        bool
	suspended        bool
	onSubmit         func(string)
	onFollowUp       func(string)
	onShiftEnter     func()
	onNewline        func()
	onRestoreQueued  func()
	onEscape         func() bool
	onTranscriptTop  func()
	onTranscriptEnd  func()
	onComplete       func(string, int) (string, int, bool)
	// interceptKey lets an open autocomplete list take Tab/Enter/Escape
	// before the editor's own bindings (Pi's Editor autocomplete mode).
	interceptKey func(gotui.Key) bool
	onChange     func(string)
	onEdit       func()
	text         string
	cursorPos    int
	// Pi's undo stack: undoText/undoCursor/undoPastes/undoCounter are the
	// newest snapshot (hasUndo), undoOlder the older ones, oldest first.
	undoText    string
	undoCursor  int
	undoPastes  map[int]string
	undoCounter int
	hasUndo     bool
	undoOlder   []inputSnapshot
	// pastes holds the content behind paste markers (Pi's large-paste
	// markers); pasteCounter numbers them.
	pastes       map[int]string
	pasteCounter int
	// Pi's kill ring: yankText is the newest entry ("" when empty),
	// killOlder the older ones, oldest first.
	yankText  string
	killOlder []string
	// lastAction is Pi's: "kill" (kills accumulate), "yank" (yank-pop
	// allowed), "type-word" (typing coalesces undo) or "".
	lastAction string
	// jumpMode is Pi's character jump awaiting its target: "forward",
	// "backward" or "".
	jumpMode string
	focused  bool
}

type inputSnapshot struct {
	text    string
	cursor  int
	pastes  map[int]string
	counter int
}

func newMultilineInput(width int, placeholder string, onSubmit func(string), onChange func(string)) *multilineInput {
	return &multilineInput{
		width:            width,
		border:           gotui.BorderNone,
		placeholder:      placeholder,
		placeholderStyle: gotui.NewStyle().Dim(),
		textStyle:        gotui.NewStyle(),
		autoFocus:        true,
		onSubmit:         onSubmit,
		onChange:         onChange,
	}
}

func (m *multilineInput) BindApp(app *gotui.App) { m.app = app }
func (m *multilineInput) Text() string           { return m.text }

// SetText is Pi's setText: undoable when the text changes; the cursor goes
// to the end.
func (m *multilineInput) SetText(s string) {
	m.jumpMode, m.lastAction = "", ""
	if s != m.text {
		m.snapshotUndo()
	}
	m.clearPastes()
	m.text = s
	m.cursorPos = utf8.RuneCountInString(s)
	m.notifyChanged()
}
func (m *multilineInput) Clear() { m.SetText("") }

// Reset replaces the text with no undo history, as Pi's submit leaves the
// editor (gi also uses it when switching sessions' drafts).
func (m *multilineInput) Reset(s string) {
	m.SetText(s)
	m.undoText, m.undoCursor, m.undoPastes, m.undoCounter, m.hasUndo, m.undoOlder = "", 0, nil, 0, false, nil
}
func (m *multilineInput) IsFocusable() bool { return !m.suspended }
func (m *multilineInput) IsTabStop() bool   { return !m.suspended }
func (m *multilineInput) IsFocused() bool   { return m.focused }
func (m *multilineInput) Focus() {
	m.focused = true
}
func (m *multilineInput) Blur() { m.focused = false }

// Watchers: none. The cursor is steady like Pi's (reverse video, no
// blink); a blink timer re-rendered the whole UI twice a second when idle.
func (m *multilineInput) Watchers() []gotui.Watcher { return nil }

func (m *multilineInput) KeyMap() gotui.KeyMap {
	if m.suspended {
		return nil
	}
	// Pi's tui.editor bindings. reset runs an action that ends kill/yank/
	// typing sequences; keep one that manages lastAction itself. Any key
	// but a printable one cancels a pending jump.
	reset := func(fn func()) func(gotui.KeyEvent) {
		return func(gotui.KeyEvent) { m.jumpMode, m.lastAction = "", ""; fn() }
	}
	keep := func(fn func()) func(gotui.KeyEvent) {
		return func(gotui.KeyEvent) { m.jumpMode = ""; fn() }
	}
	undoKey, _ := defaultPiKeys.undo()
	followUpKey, _ := defaultPiKeys.followUp()
	dequeueKey, _ := defaultPiKeys.dequeue()
	return gotui.KeyMap{
		gotui.OnFocused(gotui.AnyRune, m.insertRune),
		gotui.OnFocused(gotui.KeyBackspace, reset(m.backspace)),
		gotui.OnFocused(gotui.KeyDelete, reset(m.delete)),
		gotui.OnFocused(gotui.KeyLeft, reset(m.moveLeft)),
		gotui.OnFocused(gotui.KeyCtrlB, reset(m.moveLeft)),
		gotui.OnFocused(gotui.KeyRight, reset(m.moveRight)),
		gotui.OnFocused(gotui.KeyCtrlF, reset(m.moveRight)),
		gotui.OnFocused(gotui.KeyLeft.Alt(), reset(m.moveWordLeft)),
		gotui.OnFocused(gotui.KeyLeft.Ctrl(), reset(m.moveWordLeft)),
		gotui.OnFocused(gotui.Rune('b').Alt(), reset(m.moveWordLeft)),
		gotui.OnFocused(gotui.KeyRight.Alt(), reset(m.moveWordRight)),
		gotui.OnFocused(gotui.KeyRight.Ctrl(), reset(m.moveWordRight)),
		gotui.OnFocused(gotui.Rune('f').Alt(), reset(m.moveWordRight)),
		gotui.OnFocused(gotui.KeyHome, reset(m.moveHome)),
		gotui.OnFocused(gotui.KeyEnd, reset(m.moveEnd)),
		// Pi 1.1.0 reserves Ctrl+Home/End for the fullscreen viewport;
		// they never move the editor cursor, including in regular mode.
		gotui.OnFocused(gotui.KeyHome.Ctrl(), reset(func() {
			if m.onTranscriptTop != nil {
				m.onTranscriptTop()
			}
		})),
		gotui.OnFocused(gotui.KeyEnd.Ctrl(), reset(func() {
			if m.onTranscriptEnd != nil {
				m.onTranscriptEnd()
			}
		})),
		gotui.OnFocused(gotui.KeyCtrlA, reset(m.moveHome)),
		gotui.OnFocused(gotui.KeyCtrlE, reset(m.moveEnd)),
		gotui.OnFocused(gotui.KeyCtrlU, keep(m.deleteToLineStart)),
		gotui.OnFocused(gotui.KeyCtrlK, keep(m.deleteToLineEnd)),
		gotui.OnFocused(gotui.KeyCtrlW, keep(m.deleteWordBackward)),
		gotui.OnFocused(gotui.KeyBackspace.Alt(), keep(m.deleteWordBackward)),
		gotui.OnFocused(gotui.Rune('d').Alt(), keep(m.deleteWordForward)),
		gotui.OnFocused(gotui.KeyDelete.Alt(), keep(m.deleteWordForward)),
		gotui.OnFocused(undoKey, reset(m.undo)),
		gotui.OnFocused(gotui.KeyCtrlY, keep(m.yank)),
		gotui.OnFocused(gotui.Rune('y').Alt(), keep(m.yankPop)),
		gotui.OnFocused(gotui.Rune(']').Ctrl(), func(gotui.KeyEvent) { m.toggleJump("forward") }),
		gotui.OnFocused(gotui.Rune(']').Ctrl().Alt(), func(gotui.KeyEvent) { m.toggleJump("backward") }),
		gotui.OnFocused(gotui.KeyTab, func(ke gotui.KeyEvent) {
			if m.interceptKey != nil && m.interceptKey(gotui.KeyTab) {
				return
			}
			m.complete()
		}),
		gotui.OnFocused(gotui.KeyEnter, func(ke gotui.KeyEvent) {
			if m.interceptKey != nil && m.interceptKey(gotui.KeyEnter) {
				return
			}
			m.enter(ke)
		}),
		gotui.OnFocused(gotui.KeyEnter.Shift(), m.enter),
		gotui.OnFocused(gotui.KeyCtrlJ, func(ke gotui.KeyEvent) {
			if m.onNewline != nil {
				m.onNewline()
			} else {
				m.insertLiteral('\n')
			}
		}),
		gotui.OnFocused(followUpKey, func(gotui.KeyEvent) { m.enter(gotui.KeyEvent{Key: gotui.KeyEnter, Mod: gotui.ModAlt}) }),
		gotui.OnFocused(dequeueKey, func(ke gotui.KeyEvent) {
			m.jumpMode, m.lastAction = "", ""
			if m.onRestoreQueued != nil {
				m.onRestoreQueued()
			}
		}),
		gotui.OnFocused(gotui.KeyEscape, func(_ gotui.KeyEvent) {
			m.jumpMode, m.lastAction = "", ""
			if m.interceptKey != nil && m.interceptKey(gotui.KeyEscape) {
				return
			}
			if m.onEscape != nil {
				m.onEscape()
			}
			// Pi's editor retains focus after an idle Escape. Selection and
			// compaction handlers may consume the key, but no fallback blurs
			// the composer or discards its draft.
		}),
	}
}

func (m *multilineInput) Render(app *gotui.App) *gotui.Element {
	lines := m.renderLines()
	help := m.helpLine()
	if help != "" {
		lines = append(lines, renderedLine{text: help, placeholder: true, cursor: -1})
	}
	totalHeight := len(lines)
	if totalHeight < 1 {
		totalHeight = 1
	}
	if m.border != gotui.BorderNone {
		totalHeight += 2
	}
	root := gotui.New(
		gotui.WithDirection(gotui.Column),
		gotui.WithWidth(m.width),
		gotui.WithHeight(totalHeight),
		gotui.WithFocusable(!m.suspended),
		gotui.WithAutoFocus(m.autoFocus && !m.suspended),
		gotui.WithBorder(m.border),
	)
	root.SetOnFocus(func(e *gotui.Element) { m.Focus() })
	root.SetOnBlur(func(e *gotui.Element) { m.Blur() })
	for _, line := range lines {
		style := m.textStyle
		// The empty editor is just Pi's reverse-video cursor cell, not dim text.
		if line.placeholder && line.cursor < 0 {
			style = m.placeholderStyle
		}
		root.AddChild(renderEditorLine(line, style))
	}
	return root
}

// renderEditorLine draws Pi's fake cursor: the grapheme under the cursor in
// reverse video, or a reverse-video space at the end of the line. The draft
// text itself is never shifted by a cursor glyph.
func renderEditorLine(line renderedLine, style gotui.Style) *gotui.Element {
	options := []gotui.Option{gotui.WithHeight(1), gotui.WithWrap(false)}
	if line.cursor < 0 {
		return gotui.New(append(options, gotui.WithText(line.text), gotui.WithTextStyle(style))...)
	}
	before, under, after := line.text[:line.cursor], line.text[line.cursor:line.cursorEnd], line.text[line.cursorEnd:]
	if under == "" {
		under = " "
	}
	spans := make([]gotui.TextSpan, 0, 3)
	if before != "" {
		spans = append(spans, gotui.TextSpan{Text: before, Style: style})
	}
	spans = append(spans, gotui.TextSpan{Text: under, Style: style.Reverse()})
	if after != "" {
		spans = append(spans, gotui.TextSpan{Text: after, Style: style})
	}
	return gotui.New(append(options, gotui.WithRichText(spans...))...)
}

type renderedLine struct {
	text        string
	placeholder bool
	// cursor/cursorEnd are byte offsets of the reverse-video cursor grapheme
	// in text; cursor < 0 means no cursor, cursor == cursorEnd a trailing cell.
	cursor, cursorEnd int
}

// displayText is the row as painted, including a trailing cursor cell.
func (l renderedLine) displayText() string {
	if l.cursor >= 0 && l.cursor == l.cursorEnd {
		return l.text + " "
	}
	return l.text
}

type editorLayoutKey struct {
	text          string
	width, cursor int
	focused       bool
}
type editorLayoutCache struct {
	key       editorLayoutKey
	lines     []renderedLine
	cursorRow int
}

// editorLayoutWidth mirrors Pi's Editor with paddingX 0: one column is
// reserved so a cursor at the end of a full row still fits.
func editorLayoutWidth(width int) int {
	return max(1, width-1)
}

func (m *multilineInput) renderLines() []renderedLine {
	visibleWidth := m.width
	if m.border != gotui.BorderNone {
		visibleWidth -= 2
	}
	visibleWidth = editorLayoutWidth(visibleWidth)
	key := editorLayoutKey{text: m.text, width: visibleWidth, cursor: m.cursorPos, focused: m.focused}
	if m.text == "" {
		m.layoutCache = nil
		m.scrollRow = 0
		m.totalRows = 1
		if m.focused {
			return []renderedLine{{placeholder: true, cursor: 0, cursorEnd: 0}}
		}
		return []renderedLine{{placeholder: true, cursor: -1}}
	}
	cached := m.layoutCache
	if cached == nil || cached.key != key {
		lines, cursorRow := m.layoutLines(visibleWidth)
		cached = &editorLayoutCache{key: key, lines: lines, cursorRow: cursorRow}
		m.layoutCache = cached
	}
	lines, cursorRow := cached.lines, cached.cursorRow
	m.totalRows = len(lines)
	limit := m.maxLines
	if limit <= 0 || len(lines) <= limit {
		m.scrollRow = 0
		return lines[:len(lines):len(lines)]
	}
	if cursorRow < m.scrollRow {
		m.scrollRow = cursorRow
	}
	if cursorRow >= m.scrollRow+limit {
		m.scrollRow = cursorRow - limit + 1
	}
	m.scrollRow = max(0, min(m.scrollRow, len(lines)-limit))
	// Limit capacity so Render's optional append cannot overwrite cached rows.
	return lines[m.scrollRow : m.scrollRow+limit : m.scrollRow+limit]
}

// hiddenRows reports rows above and below the last rendered viewport, for
// Pi's "↑ N more" / "↓ N more" editor borders.
func (m *multilineInput) hiddenRows() (above, below int) {
	visible := m.totalRows - m.scrollRow
	if m.maxLines > 0 {
		visible = min(visible, m.maxLines)
	}
	return m.scrollRow, max(0, m.totalRows-m.scrollRow-visible)
}

type editorGrapheme struct {
	text      string
	width     int
	runeStart int
	runeCount int
}

// layoutLines ports Pi's Editor.layoutText/wordWrapLine: logical lines wrap at
// word boundaries (or between CJK characters), falling back to grapheme
// breaks for words longer than the row. The cursor is a rune offset.
func (m *multilineInput) layoutLines(visibleWidth int) ([]renderedLine, int) {
	cursorPos := max(0, min(m.cursorPos, utf8.RuneCountInString(m.text)))
	lines := []renderedLine{}
	cursorRow := -1
	var logical []editorGrapheme
	lineStart := 0
	runeOffset := 0
	flushLogical := func(lineEnd int) {
		hasCursor := cursorPos >= lineStart && cursorPos <= lineEnd && cursorRow < 0
		// Grapheme index holding the cursor; len(logical) means end of line.
		cursorIndex := len(logical)
		if hasCursor {
			for i, g := range logical {
				if cursorPos < g.runeStart+g.runeCount {
					cursorIndex = i
					break
				}
			}
		}
		chunks := editorWordWrap(logical, visibleWidth)
		for ci, chunk := range chunks {
			inChunk := false
			if hasCursor {
				if ci == len(chunks)-1 {
					inChunk = cursorIndex >= chunk[0]
				} else {
					inChunk = cursorIndex >= chunk[0] && cursorIndex < chunk[1]
				}
			}
			var b strings.Builder
			line := renderedLine{cursor: -1}
			for i := chunk[0]; i < chunk[1]; i++ {
				if inChunk && m.focused && i == cursorIndex {
					line.cursor = b.Len()
				}
				b.WriteString(logical[i].text)
				if inChunk && m.focused && i == cursorIndex {
					line.cursorEnd = b.Len()
				}
			}
			line.text = b.String()
			if inChunk {
				cursorRow = len(lines)
				if m.focused && cursorIndex >= chunk[1] {
					line.cursor, line.cursorEnd = len(line.text), len(line.text)
				}
			}
			lines = append(lines, line)
		}
		logical = logical[:0]
	}
	it := graphemes.FromString(m.text)
	for it.Next() {
		cluster := it.Value()
		count := utf8.RuneCountInString(cluster)
		if cluster == "\n" || cluster == "\r\n" {
			flushLogical(runeOffset)
			runeOffset += count
			lineStart = runeOffset
			continue
		}
		width := gotui.StringWidth(cluster)
		// A one-column terminal cannot paint a two-cell glyph. Substitute only
		// its display cell; the draft and logical cursor remain byte-identical.
		if width > visibleWidth {
			cluster, width = "�", 1
		}
		logical = append(logical, editorGrapheme{text: cluster, width: width, runeStart: runeOffset, runeCount: count})
		runeOffset += count
	}
	flushLogical(runeOffset)
	if cursorRow < 0 {
		cursorRow = len(lines) - 1
	}
	return lines, cursorRow
}

// editorWordWrap returns [start,end) grapheme ranges, as Pi's wordWrapLine.
func editorWordWrap(segments []editorGrapheme, maxWidth int) [][2]int {
	total := 0
	for _, g := range segments {
		total += g.width
	}
	if len(segments) == 0 || total <= maxWidth {
		return [][2]int{{0, len(segments)}}
	}
	var chunks [][2]int
	currentWidth, chunkStart := 0, 0
	wrapOppIndex, wrapOppWidth := -1, 0
	for i, g := range segments {
		isWs := isEditorWhitespace(g.text)
		if currentWidth+g.width > maxWidth {
			if wrapOppIndex >= 0 && currentWidth-wrapOppWidth+g.width <= maxWidth {
				chunks = append(chunks, [2]int{chunkStart, wrapOppIndex})
				chunkStart = wrapOppIndex
				currentWidth -= wrapOppWidth
			} else if chunkStart < i {
				chunks = append(chunks, [2]int{chunkStart, i})
				chunkStart = i
				currentWidth = 0
			}
			wrapOppIndex = -1
		}
		currentWidth += g.width
		if i+1 < len(segments) {
			next := segments[i+1]
			nextWs := isEditorWhitespace(next.text)
			if isWs && !nextWs {
				wrapOppIndex, wrapOppWidth = i+1, currentWidth
			} else if !isWs && !nextWs && (isEditorCJK(g.text) || isEditorCJK(next.text)) {
				wrapOppIndex, wrapOppWidth = i+1, currentWidth
			}
		}
	}
	return append(chunks, [2]int{chunkStart, len(segments)})
}

func isEditorWhitespace(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsSpace(r)
}

func isEditorCJK(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) ||
		(r >= 0x3000 && r <= 0x303f) || (r >= 0xff00 && r <= 0xffef)
}

func (m *multilineInput) helpLine() string {
	return ""
}

func (m *multilineInput) insertRune(ke gotui.KeyEvent) {
	if m.jumpMode != "" { // Pi: a printable key is the jump target
		direction := m.jumpMode
		m.jumpMode, m.lastAction = "", ""
		m.jumpToChar(ke.Rune, direction)
		return
	}
	// Pi's fish-style undo coalescing: a word typed is one undo unit, each
	// whitespace its own.
	if unicode.IsSpace(ke.Rune) || m.lastAction != "type-word" {
		m.snapshotUndo()
	}
	m.lastAction = "type-word"
	runes := []rune(m.text)
	pos := m.clampCursor()
	runes = append(runes[:pos], append([]rune{ke.Rune}, runes[pos:]...)...)
	m.text = string(runes)
	m.cursorPos = pos + 1
	m.notifyEdited()
}

func (m *multilineInput) backspace() {
	runes := []rune(m.text)
	pos := m.clampCursor()
	if pos == 0 || len(runes) == 0 {
		return
	}
	m.snapshotUndo()
	if start, ok := m.markerEndingAt(pos); ok { // a paste marker is one unit
		marker := string(runes[start:pos])
		m.text = string(append(runes[:start:start], runes[pos:]...))
		m.removePaste(marker)
		m.cursorPos = start
		m.notifyEdited()
		return
	}
	runes = append(runes[:pos-1], runes[pos:]...)
	m.text = string(runes)
	m.cursorPos = pos - 1
	m.notifyEdited()
}

func (m *multilineInput) delete() {
	runes := []rune(m.text)
	pos := m.clampCursor()
	if pos >= len(runes) {
		return
	}
	m.snapshotUndo()
	if end, ok := m.markerStartingAt(pos); ok {
		marker := string(runes[pos:end])
		m.text = string(append(runes[:pos:pos], runes[end:]...))
		m.removePaste(marker)
		m.notifyEdited()
		return
	}
	runes = append(runes[:pos], runes[pos+1:]...)
	m.text = string(runes)
	m.notifyEdited()
}

func (m *multilineInput) moveLeft() {
	if pos := m.clampCursor(); pos > 0 {
		if start, ok := m.markerEndingAt(pos); ok {
			m.cursorPos = start
		} else {
			m.cursorPos = pos - 1
		}
		m.markDirty()
	}
}
func (m *multilineInput) moveRight() {
	if pos := m.clampCursor(); pos < utf8.RuneCountInString(m.text) {
		if end, ok := m.markerStartingAt(pos); ok {
			m.cursorPos = end
		} else {
			m.cursorPos = pos + 1
		}
		m.markDirty()
	}
}

// moveHome and moveEnd go to the start and end of the cursor's line (Pi's
// cursorLineStart/cursorLineEnd).
func (m *multilineInput) moveHome() { m.cursorPos = m.lineStart(m.clampCursor()); m.markDirty() }
func (m *multilineInput) moveEnd()  { m.cursorPos = m.lineEnd(m.clampCursor()); m.markDirty() }

func (m *multilineInput) lineStart(pos int) int {
	runes := []rune(m.text)
	for pos > 0 && runes[pos-1] != '\n' {
		pos--
	}
	return pos
}

func (m *multilineInput) lineEnd(pos int) int {
	runes := []rune(m.text)
	for pos < len(runes) && runes[pos] != '\n' {
		pos++
	}
	return pos
}

// moveWordLeft and moveWordRight are Pi's moveWordBackwards/Forwards: at a
// line edge they cross the line break; otherwise Pi's findWordBackward/
// findWordForward within the line.
func (m *multilineInput) moveWordLeft() {
	m.cursorPos = m.wordBackward(m.clampCursor())
	m.markDirty()
}

func (m *multilineInput) moveWordRight() {
	m.cursorPos = m.wordForward(m.clampCursor())
	m.markDirty()
}

func (m *multilineInput) wordBackward(pos int) int {
	if start := m.lineStart(pos); pos == start {
		return max(0, pos-1)
	}
	return findWordBackward([]rune(m.text), m.lineStart(pos), pos, m.markerSpans())
}

func (m *multilineInput) wordForward(pos int) int {
	runes := []rune(m.text)
	if end := m.lineEnd(pos); pos == end {
		return min(len(runes), pos+1)
	}
	return findWordForward(runes, m.lineEnd(pos), pos, m.markerSpans())
}

// piPunctuation is pi-tui's PUNCTUATION_REGEX.
const piPunctuation = "(){}[]<>.,;:'\"!?+-=*/\\|&%^$#@~`"

// wordLike approximates Intl.Segmenter's word-like segments split at Pi's
// punctuation, which is what Pi's word navigation stops at.
func wordLike(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r)
}

// findWordBackward is pi-tui's findWordBackward over runes[lo:pos]: skip
// whitespace, then one paste marker (atomic), a word, or a run of other
// characters.
func findWordBackward(runes []rune, lo, pos int, markers [][2]int) int {
	markerEnding := func(i int) (int, bool) {
		for _, s := range markers {
			if s[1] == i {
				return s[0], true
			}
		}
		return 0, false
	}
	for pos > lo && unicode.IsSpace(runes[pos-1]) {
		pos--
	}
	if pos == lo {
		return pos
	}
	if start, ok := markerEnding(pos); ok {
		return start
	}
	if wordLike(runes[pos-1]) {
		for pos > lo && wordLike(runes[pos-1]) {
			pos--
		}
		return pos
	}
	for pos > lo && !wordLike(runes[pos-1]) && !unicode.IsSpace(runes[pos-1]) {
		if _, ok := markerEnding(pos); ok {
			break
		}
		pos--
	}
	return pos
}

// findWordForward is pi-tui's findWordForward over runes[pos:hi].
func findWordForward(runes []rune, hi, pos int, markers [][2]int) int {
	markerStarting := func(i int) (int, bool) {
		for _, s := range markers {
			if s[0] == i {
				return s[1], true
			}
		}
		return 0, false
	}
	for pos < hi && unicode.IsSpace(runes[pos]) {
		pos++
	}
	if pos == hi {
		return pos
	}
	if end, ok := markerStarting(pos); ok {
		return end
	}
	if wordLike(runes[pos]) {
		for pos < hi && wordLike(runes[pos]) {
			pos++
		}
		return pos
	}
	for pos < hi && !wordLike(runes[pos]) && !unicode.IsSpace(runes[pos]) {
		if _, ok := markerStarting(pos); ok {
			break
		}
		pos++
	}
	return pos
}

// The kills are Pi's: within the cursor's line, removing the line break at
// its edge; the text goes to the kill ring, accumulating while the previous
// action was a kill (backward kills prepend, forward kills append).

func (m *multilineInput) deleteWordBackward() {
	end := m.clampCursor()
	if lineStart := m.lineStart(end); end == lineStart {
		m.killRange(end-1, end, true)
	} else {
		m.killRange(findWordBackward([]rune(m.text), lineStart, end, m.markerSpans()), end, true)
	}
}

func (m *multilineInput) deleteWordForward() {
	start := m.clampCursor()
	if lineEnd := m.lineEnd(start); start == lineEnd {
		m.killRange(start, start+1, false)
	} else {
		m.killRange(start, findWordForward([]rune(m.text), lineEnd, start, m.markerSpans()), false)
	}
}

func (m *multilineInput) deleteToLineStart() {
	pos := m.clampCursor()
	if start := m.lineStart(pos); start < pos {
		m.killRange(start, pos, true)
	} else {
		m.killRange(pos-1, pos, true)
	}
}

func (m *multilineInput) deleteToLineEnd() {
	pos := m.clampCursor()
	if end := m.lineEnd(pos); end > pos {
		m.killRange(pos, end, false)
	} else {
		m.killRange(pos, pos+1, false)
	}
}

// killRange deletes [start,end) into the kill ring.
func (m *multilineInput) killRange(start, end int, backward bool) {
	runes := []rune(m.text)
	if start < 0 || end > len(runes) || start >= end {
		return
	}
	m.snapshotUndo()
	m.pushKill(string(runes[start:end]), backward, m.lastAction == "kill")
	m.lastAction = "kill"
	m.text = string(append(runes[:start:start], runes[end:]...))
	m.cursorPos = start
	m.notifyEdited()
}

// pushKill is Pi's KillRing.push.
func (m *multilineInput) pushKill(text string, prepend, accumulate bool) {
	if text == "" {
		return
	}
	switch {
	case accumulate && m.yankText != "" && prepend:
		m.yankText = text + m.yankText
	case accumulate && m.yankText != "":
		m.yankText += text
	default:
		if m.yankText != "" {
			m.killOlder = append(m.killOlder, m.yankText)
		}
		m.yankText = text
	}
}

func (m *multilineInput) killRingLen() int {
	if m.yankText == "" {
		return 0
	}
	return len(m.killOlder) + 1
}

// rotateKills is Pi's KillRing.rotate: the newest entry moves to the front.
func (m *multilineInput) rotateKills() {
	if len(m.killOlder) == 0 {
		return
	}
	ring := append([]string{m.yankText}, m.killOlder...)
	m.yankText, m.killOlder = ring[len(ring)-1], ring[:len(ring)-1]
}

func isWordSpace(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }

func (m *multilineInput) complete() {
	m.jumpMode, m.lastAction = "", ""
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
	m.notifyEdited()
}

func (m *multilineInput) enter(ke gotui.KeyEvent) {
	m.jumpMode, m.lastAction = "", ""
	if ke.Mod&gotui.ModShift != 0 {
		if m.onShiftEnter != nil {
			m.onShiftEnter()
			return
		}
		m.insertLiteral('\n')
		return
	}
	if ke.Mod&gotui.ModAlt != 0 && m.onFollowUp != nil {
		m.onFollowUp(m.ExpandedText())
		return
	}
	if m.onSubmit != nil {
		m.onSubmit(m.ExpandedText())
	}
}

func (m *multilineInput) insertLiteral(r rune) {
	m.jumpMode, m.lastAction = "", ""
	m.snapshotUndo()
	runes := []rune(m.text)
	pos := m.clampCursor()
	runes = append(runes[:pos], append([]rune{r}, runes[pos:]...)...)
	m.text = string(runes)
	m.cursorPos = pos + 1
	m.notifyEdited()
}

func (m *multilineInput) snapshotUndo() {
	if m.hasUndo {
		m.undoOlder = append(m.undoOlder, inputSnapshot{m.undoText, m.undoCursor, m.undoPastes, m.undoCounter})
	}
	m.undoText = m.text
	m.undoCursor = m.clampCursor()
	m.undoPastes, m.undoCounter = copyPastes(m.pastes), m.pasteCounter
	m.hasUndo = true
}

// undo is Pi's: restore the newest snapshot and drop it.
func (m *multilineInput) undo() {
	if !m.hasUndo {
		return
	}
	m.text, m.cursorPos, m.pastes, m.pasteCounter = m.undoText, m.undoCursor, m.undoPastes, m.undoCounter
	if n := len(m.undoOlder); n > 0 {
		top := m.undoOlder[n-1]
		m.undoText, m.undoCursor, m.undoPastes, m.undoCounter = top.text, top.cursor, top.pastes, top.counter
		m.undoOlder = m.undoOlder[:n-1]
	} else {
		m.undoText, m.undoCursor, m.undoPastes, m.undoCounter, m.hasUndo = "", 0, nil, 0, false
	}
	m.lastAction = ""
	m.notifyEdited()
}

// yank is Pi's: insert the newest kill.
func (m *multilineInput) yank() {
	if m.yankText == "" {
		return
	}
	m.snapshotUndo()
	m.insertText(m.yankText)
	m.lastAction = "yank"
}

// yankPop is Pi's: right after a yank, replace the yanked text with the
// previous kill.
func (m *multilineInput) yankPop() {
	if m.lastAction != "yank" || m.killRingLen() <= 1 {
		return
	}
	m.snapshotUndo()
	runes := []rune(m.text)
	pos := m.clampCursor()
	start := max(0, pos-utf8.RuneCountInString(m.yankText))
	m.text = string(append(runes[:start:start], runes[pos:]...))
	m.cursorPos = start
	m.rotateKills()
	m.insertText(m.yankText)
	m.lastAction = "yank"
}

func (m *multilineInput) insertText(text string) {
	runes := []rune(m.text)
	pos := m.clampCursor()
	inserted := []rune(text)
	m.text = string(append(runes[:pos:pos], append(inserted, runes[pos:]...)...))
	m.cursorPos = pos + len(inserted)
	m.notifyEdited()
}

// toggleJump starts Pi's character jump (the next printable key is the
// target), or cancels it when its key is pressed again.
func (m *multilineInput) toggleJump(direction string) {
	m.lastAction = ""
	if m.jumpMode != "" {
		m.jumpMode = ""
		return
	}
	m.jumpMode = direction
}

// jumpToChar is Pi's: the first occurrence after (or before) the cursor,
// across lines; the cursor stays when there is none.
func (m *multilineInput) jumpToChar(target rune, direction string) {
	runes := []rune(m.text)
	pos := m.clampCursor()
	if direction == "forward" {
		for i := pos + 1; i < len(runes); i++ {
			if runes[i] == target {
				m.cursorPos = i
				m.markDirty()
				return
			}
		}
		return
	}
	for i := pos - 1; i >= 0; i-- {
		if runes[i] == target {
			m.cursorPos = i
			m.markDirty()
			return
		}
	}
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

func (m *multilineInput) notifyEdited() {
	if m.onEdit != nil {
		m.onEdit()
	}
	m.notifyChanged()
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
