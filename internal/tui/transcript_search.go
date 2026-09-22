package tui

import (
	"fmt"
	"reflect"
	"strings"

	gotui "github.com/grindlemire/go-tui"
)

type transcriptSearchRow struct {
	text   string
	spans  []gotui.TextSpan
	prompt bool
}

type transcriptSearch struct {
	active      bool
	input       *multilineInput
	rows        []transcriptSearchRow
	matches     []int
	selected    int
	width       int
	viewport    int
	source      []string
	expanded    map[string]bool
	savedScroll int
	savedFollow bool
	query       string
}

// Index the same retained, wrapped and collapsed output the fullscreen renderer
// displays. Tool metadata and hidden lines never become phantom search matches.
func (c *chatTUI) renderedTranscriptRows(width int) []transcriptSearchRow {
	rows := c.transcriptRowsAtWidth(width)
	// go-tui reserves one column whenever vertical content overflows. Match
	// that layout even when no custom scrollbar style was chosen.
	if len(rows) > c.transcriptViewportHeight() && width > 1 {
		return c.transcriptRowsAtWidth(width - 1)
	}
	return rows
}

func (c *chatTUI) transcriptRowsAtWidth(width int) []transcriptSearchRow {
	width = max(1, width)
	refs := c.transcriptBlockRefs
	defer func() { c.transcriptBlockRefs = refs }()
	var rows []transcriptSearchRow
	for _, block := range c.buildTranscriptRenderableBlocks(c.visibleTranscript()) {
		el := c.renderTranscriptBlock(block)
		height := max(1, el.HeightForWidth(width))
		root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(height))
		root.AddChild(el)
		buf := gotui.NewBuffer(width, height)
		root.Render(buf, width, height)
		for y := 0; y < height; y++ {
			var text strings.Builder
			spans := make([]gotui.TextSpan, 0, width)
			for x := 0; x < width; x++ {
				cell := buf.Cell(x, y)
				if cell.Width == 0 {
					continue
				}
				value := string(cell.Rune) + cell.Combining
				if cell.Rune == 0 {
					value = " "
				}
				text.WriteString(value)
				spans = append(spans, gotui.TextSpan{Text: value, Style: cell.Style})
			}
			rows = append(rows, transcriptSearchRow{text: strings.TrimRight(text.String(), " "), spans: spans, prompt: block.Kind == "user" && y == 0})
		}
	}
	return rows
}

func (c *chatTUI) toggleTranscriptSearch() {
	if c.regularMode {
		return
	}
	if c.search.active {
		c.closeTranscriptSearch()
		return
	}
	if c.modelMenuOpen || c.editorAskActive {
		return
	}
	c.ensureInput()
	offset := c.transcriptScroll
	if c.transcriptRef != nil && c.transcriptRef.El() != nil {
		_, offset = c.transcriptRef.El().ScrollOffset()
	}
	input := c.search.input
	c.search = transcriptSearch{active: true, input: input, selected: -1, savedScroll: offset, savedFollow: c.stickToBottom}
	if c.search.input == nil {
		c.search.input = newMultilineInput(c.currentContentWidth(), "", func(string) { c.moveTranscriptSearch(1) }, func(query string) { c.updateTranscriptSearchQuery(query) })
	}
	c.search.input.SetText("")
	c.search.input.onEscape = func() bool { c.closeTranscriptSearch(); return true }
	c.search.input.onShiftEnter = func() { c.moveTranscriptSearch(-1) }
	c.search.input.onNewline = func() { c.moveTranscriptSearch(1) }
	c.stickToBottom = false
	c.search.input.Focus()
	if c.app != nil {
		if c.app.Focused() == nil {
			c.app.FocusNext()
		}
		c.app.MarkDirty()
	}
}

func (c *chatTUI) closeTranscriptSearch() {
	if !c.search.active {
		return
	}
	c.transcriptScroll, c.stickToBottom = c.search.savedScroll, c.search.savedFollow
	c.search = transcriptSearch{input: c.search.input}
	c.input.Focus()
	c.focusInput()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) updateTranscriptSearchQuery(query string) {
	c.search.query = query
	c.search.selected = -1
	c.search.matches = nil
	if query != "" {
		for row, value := range c.search.rows {
			if strings.Contains(strings.ToLower(value.text), strings.ToLower(query)) {
				c.search.matches = append(c.search.matches, row)
			}
		}
	}
	if len(c.search.matches) > 0 {
		c.search.selected = 0
		c.setTranscriptPosition(c.search.matches[0])
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) refreshTranscriptSearch(width int) {
	if !c.search.active {
		return
	}
	if c.search.width == width && c.search.viewport == c.transcriptViewportHeight() && reflect.DeepEqual(c.search.source, c.transcript) && reflect.DeepEqual(c.search.expanded, c.transcriptExpanded) {
		return
	}
	previousRow := -1
	if c.search.selected >= 0 && c.search.selected < len(c.search.matches) {
		previousRow = c.search.matches[c.search.selected]
	}
	c.search.rows = c.renderedTranscriptRows(width)
	c.search.width = width
	c.search.viewport = c.transcriptViewportHeight()
	c.search.source = append([]string(nil), c.transcript...)
	c.search.expanded = map[string]bool{}
	for key, value := range c.transcriptExpanded {
		c.search.expanded[key] = value
	}
	c.updateTranscriptSearchQuery(c.search.query)
	if previousRow >= 0 && len(c.search.matches) > 0 {
		for i, row := range c.search.matches {
			if row >= previousRow {
				c.search.selected = i
				break
			}
		}
		c.setTranscriptPosition(c.search.matches[c.search.selected])
	}
}

func (c *chatTUI) moveTranscriptSearch(delta int) {
	if !c.search.active || len(c.search.matches) == 0 {
		return
	}
	c.search.selected = (c.search.selected + delta + len(c.search.matches)) % len(c.search.matches)
	c.setTranscriptPosition(c.search.matches[c.search.selected])
	c.stickToBottom = false
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) transcriptSearchLabel(width int) string {
	label := fmt.Sprintf("Search %d/%d · Enter next · Shift-Enter prev · Esc close", max(0, c.search.selected+1), len(c.search.matches))
	if c.search.query != "" && len(c.search.matches) == 0 {
		label = "Search: no matches · Esc close"
	}
	return truncate(label, width)
}

func (c *chatTUI) renderTranscriptSearchRows(transcript *gotui.Element) {
	match := map[int]bool{}
	for _, row := range c.search.matches {
		match[row] = true
	}
	for i, row := range c.search.rows {
		spans := append([]gotui.TextSpan(nil), row.spans...)
		selected := c.search.selected >= 0 && c.search.selected < len(c.search.matches) && c.search.matches[c.search.selected] == i
		for j := range spans {
			if match[i] {
				spans[j].Style = spans[j].Style.Background(piUserBg).Underline()
			}
			if selected {
				spans[j].Style = spans[j].Style.Background(piText).Foreground(piUserBg).Bold()
			}
		}
		transcript.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1), gotui.WithWrap(false), gotui.WithRichText(spans...)))
	}
}

func (c *chatTUI) jumpTranscriptPrompt(direction int) {
	if c.regularMode || c.search.active || c.modelMenuOpen {
		return
	}
	rows := c.renderedTranscriptRows(c.currentContentWidth())
	offset := c.transcriptScroll
	if c.transcriptRef != nil && c.transcriptRef.El() != nil {
		_, offset = c.transcriptRef.El().ScrollOffset()
	}
	target := -1
	for row, value := range rows {
		if value.prompt {
			if direction < 0 && row < offset {
				target = row
			}
			if direction > 0 && row > offset {
				target = row
				break
			}
		}
	}
	if target < 0 {
		return
	}
	c.setTranscriptPosition(target)
	c.stickToBottom = false
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) transcriptSearchKeys() gotui.KeyMap {
	return gotui.KeyMap{
		gotui.OnPreemptStop(gotui.Rune('f').Ctrl().Shift(), func(gotui.KeyEvent) { c.closeTranscriptSearch() }),
		gotui.OnPreemptStop(gotui.KeyEscape, func(gotui.KeyEvent) { c.closeTranscriptSearch() }),
		gotui.OnPreemptStop(gotui.KeyCtrlC, func(gotui.KeyEvent) { c.closeTranscriptSearch() }),
		gotui.OnPreemptStop(gotui.Rune('g').Ctrl(), func(gotui.KeyEvent) { c.moveTranscriptSearch(1) }),
		gotui.OnPreemptStop(gotui.Rune('g').Ctrl().Shift(), func(gotui.KeyEvent) { c.moveTranscriptSearch(-1) }),
	}
}
