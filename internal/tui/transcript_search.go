package tui

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	gotui "github.com/grindlemire/go-tui"
)

type transcriptSearchRow struct {
	text     string
	spans    []gotui.TextSpan
	prompt   bool
	blockKey string                // pointer hit identity for tool rows, including their padding
	wrapped  []transcriptSearchRun // validated soft-wrap runs starting in this row
}

// Match columns are half-open display-cell boundaries, not UTF-8 offsets.
// A substring inside a combining/wide grapheme highlights the complete cell.
type transcriptSearchMatch struct {
	row, start, end int
	continuation    []transcriptSearchCell
}

type transcriptSearch struct {
	active      bool
	input       *multilineInput
	rows        []transcriptSearchRow
	matches     []transcriptSearchMatch
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
		baseRow := len(rows)
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
				spans = append(spans, gotui.TextSpan{Text: value, Style: cell.Style, Link: cell.Link})
			}
			key := ""
			separator, _, _ := transcriptSpacing(block.Kind)
			if y >= separator && (len(block.Body) > 0 || block.Subheader != "") && block.Kind != "user" && block.Kind != "assistant" {
				key = block.Key
			}
			rows = append(rows, transcriptSearchRow{text: strings.TrimRight(text.String(), " "), spans: spans, prompt: block.Kind == "user" && y == 1, blockKey: key})
		}
		for _, run := range transcriptWrapRuns(el, rows[baseRow:]) {
			for i := range run.cells {
				run.cells[i].row += baseRow
			}
			row := run.cells[0].row
			rows[row].wrapped = append(rows[row].wrapped, run)
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
	c.clearTranscriptSelection()
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
	c.search.input.onEscape = func() bool {
		if c.textSelection.active {
			c.clearTranscriptSelection()
		} else {
			c.closeTranscriptSearch()
		}
		return true
	}
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
	c.clearTranscriptSelection()
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
	needle := strings.ToLower(query)
	if needle != "" {
		var runs []transcriptSearchRun
		for row, value := range c.search.rows {
			c.search.matches = append(c.search.matches, transcriptRowMatches(row, value, needle)...)
			runs = append(runs, value.wrapped...)
		}
		// A validated paragraph owns all its occurrences (including same-row
		// ones), so non-overlap does not restart at a soft line boundary.
		c.search.matches = transcriptMergeWrappedMatches(c.search.matches, runs, needle)
		sort.SliceStable(c.search.matches, func(i, j int) bool {
			a, b := c.search.matches[i], c.search.matches[j]
			return a.row < b.row || a.row == b.row && a.start < b.start
		})
	}
	if len(c.search.matches) > 0 {
		c.search.selected = 0
		c.setTranscriptPosition(c.search.matches[0].row)
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
	previous := transcriptSearchMatch{row: -1}
	if c.search.selected >= 0 && c.search.selected < len(c.search.matches) {
		previous = c.search.matches[c.search.selected]
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
	if previous.row >= 0 && len(c.search.matches) > 0 {
		for i, match := range c.search.matches {
			if match.row > previous.row || (match.row == previous.row && match.start >= previous.start) {
				c.search.selected = i
				break
			}
		}
		c.setTranscriptPosition(c.search.matches[c.search.selected].row)
	}
}

func (c *chatTUI) moveTranscriptSearch(delta int) {
	if !c.search.active || len(c.search.matches) == 0 {
		return
	}
	c.search.selected = (c.search.selected + delta + len(c.search.matches)) % len(c.search.matches)
	c.setTranscriptPosition(c.search.matches[c.search.selected].row)
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

// Lowercasing can change UTF-8 byte lengths (for example İ -> i). Map every
// folded byte back to its rendered grapheme's full display-cell interval.
// Search remains row-local and non-overlapping; trimmed padding is not content.
func transcriptRowMatches(row int, value transcriptSearchRow, needle string) []transcriptSearchMatch {
	if needle == "" {
		return nil
	}
	var folded strings.Builder
	var starts, ends []int
	col := 0
	for _, span := range value.spans {
		text, width := strings.ToLower(span.Text), gotui.StringWidth(span.Text)
		folded.WriteString(text)
		for range len(text) {
			starts = append(starts, col)
			ends = append(ends, col+width)
		}
		col += width
	}
	text := strings.TrimRight(folded.String(), " ")
	var matches []transcriptSearchMatch
	for offset := 0; offset < len(text); {
		i := strings.Index(text[offset:], needle)
		if i < 0 {
			break
		}
		i += offset
		matches = append(matches, transcriptSearchMatch{row: row, start: starts[i], end: ends[i+len(needle)-1]})
		offset = i + len(needle)
	}
	return matches
}

func (c *chatTUI) renderTranscriptSearchRows(transcript *gotui.Element) {
	matches := map[int][]transcriptSearchMatch{}
	for _, match := range c.search.matches {
		matches[match.row] = append(matches[match.row], match)
		for _, part := range match.continuation {
			matches[part.row] = append(matches[part.row], transcriptSearchMatch{row: part.row, start: part.start, end: part.end})
		}
	}
	selected := transcriptSearchMatch{row: -1}
	if c.search.selected >= 0 && c.search.selected < len(c.search.matches) {
		selected = c.search.matches[c.search.selected]
	}
	selectedParts := append([]transcriptSearchCell{{row: selected.row, start: selected.start, end: selected.end}}, selected.continuation...)
	for i, row := range c.search.rows {
		spans := append([]gotui.TextSpan(nil), row.spans...)
		col := 0
		for j := range spans {
			end := col + gotui.StringWidth(spans[j].Text)
			for _, match := range matches[i] {
				if col < match.end && end > match.start {
					spans[j].Style = spans[j].Style.Background(piUserBg).Underline()
					break
				}
			}
			for _, part := range selectedParts {
				if part.row == i && col < part.end && end > part.start {
					spans[j].Style = spans[j].Style.Background(piText).Foreground(piUserBg).Bold()
					break
				}
			}
			col = end
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
		gotui.OnPreemptStop(gotui.KeyEscape, func(gotui.KeyEvent) {
			if c.textSelection.active {
				c.clearTranscriptSelection()
			} else {
				c.closeTranscriptSearch()
			}
		}),
		gotui.OnPreemptStop(gotui.KeyCtrlC, func(gotui.KeyEvent) {
			if c.textSelection.active {
				c.copyTranscriptSelection()
			} else {
				c.closeTranscriptSearch()
			}
		}),
		gotui.OnPreemptStop(gotui.Rune('x').Ctrl(), func(gotui.KeyEvent) { c.copyTranscriptSelection() }),
		gotui.OnPreemptStop(gotui.Rune('g').Ctrl(), func(gotui.KeyEvent) { c.moveTranscriptSearch(1) }),
		gotui.OnPreemptStop(gotui.Rune('g').Ctrl().Shift(), func(gotui.KeyEvent) { c.moveTranscriptSearch(-1) }),
	}
}
