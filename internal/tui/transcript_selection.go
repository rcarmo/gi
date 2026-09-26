package tui

import (
	"reflect"
	"strings"
	"time"

	gotui "github.com/grindlemire/go-tui"
)

type transcriptPoint struct{ row, col int }
type transcriptSelection struct {
	active, dragging, moved       bool
	rows                          []transcriptSearchRow
	source                        []string
	expanded                      map[string]bool
	scope                         sessionScope
	width, height                 int
	contentWidth                  int
	terminalWidth, terminalHeight int
	query                         string
	searchActive                  bool
	searchMatch                   int
	start, end                    transcriptPoint
	granularity                   int // 0 character, 1 word, 2 rendered line
	initial                       transcriptSelectionRange
	pressX, pressY                int
	clickKey                      string
	pressLink                     string
	previousFollow                bool
	pointerX, pointerY            int
	generation                    uint64
	notice                        string
	noticeUntil                   time.Time
}

func (c *chatTUI) clearTranscriptSelection() {
	next := c.textSelection.generation + 1
	c.textSelection = transcriptSelection{generation: next}
	c.selectionClicks = transcriptClickSequence{}
	c.selectionClickSnapshot = transcriptSelection{}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) selectionCurrent() bool {
	return c.selectionSnapshotCurrent(&c.textSelection)
}

func (c *chatTUI) selectionSnapshotCurrent(s *transcriptSelection) bool {
	if !s.active || !c.ownsScope(s.scope) || c.regularMode || c.modelMenuOpen {
		return false
	}
	if !reflect.DeepEqual(s.source, c.transcript) || !reflect.DeepEqual(s.expanded, c.transcriptExpanded) {
		return false
	}
	if s.searchActive != c.search.active || s.query != c.search.query || s.searchMatch != c.search.selected {
		return false
	}
	if c.transcriptRegion == nil {
		return false
	}
	if c.app != nil {
		w, h := c.app.Size()
		if w != s.terminalWidth || h != s.terminalHeight {
			return false
		}
	}
	r := c.transcriptRegion.Rect()
	return r.Width == s.width && r.Height == s.height
}

func (c *chatTUI) validateTranscriptSelection(width, height int) {
	// During a held press the current selection snapshot owns validation; the
	// released-click snapshot is not installed until release.
	if !c.textSelection.dragging && c.selectionClicks.valid && (width != c.selectionClickSnapshot.width || height != c.selectionClickSnapshot.height || !c.selectionSnapshotCurrent(&c.selectionClickSnapshot)) {
		c.selectionClicks = transcriptClickSequence{}
		c.selectionClickSnapshot = transcriptSelection{}
	}
	if !c.textSelection.active {
		return
	}
	if width != c.textSelection.width || height != c.textSelection.height || !c.selectionCurrent() {
		c.clearTranscriptSelection()
	}
}

func (c *chatTUI) selectionPoint(x, y int) transcriptPoint {
	r := c.transcriptRegion.Rect()
	_, offset := c.transcriptRegion.ScrollOffset()
	rows := len(c.textSelection.rows)
	if rows == 0 {
		return transcriptPoint{}
	}
	width := c.textSelection.contentWidth
	if width <= 0 {
		width = c.textSelection.width
	}
	col := min(max(0, x-r.X), width)
	// Interior endpoints retain the existing half-open cell convention. The
	// final content cell acts as the outer boundary so it can be selected
	// without an out-of-window mouse coordinate. Reverse drags use it too.
	if col >= width-1 {
		col = width
	}
	return transcriptPoint{row: min(rows-1, max(0, offset+min(r.Height-1, max(0, y-r.Y)))), col: col}
}

// Return ordered half-open positions. Display columns are converted to complete
// grapheme cells at copy/render time; no ANSI or partial wide glyph is copied.
func (s *transcriptSelection) bounds() (transcriptPoint, transcriptPoint) {
	a, b := s.start, s.end
	if a.row > b.row || (a.row == b.row && a.col > b.col) {
		return b, a
	}
	return a, b
}
func selectionSpanRange(row int, a, b transcriptPoint, width int) (int, int) {
	if row < a.row || row > b.row {
		return 0, 0
	}
	start, end := 0, width
	if row == a.row {
		start = a.col
	}
	if row == b.row {
		end = b.col
	}
	return start, end
}
func (s *transcriptSelection) text() string {
	if !s.active || !s.moved {
		return ""
	}
	a, b := s.bounds()
	var lines []string
	for row := a.row; row <= b.row && row < len(s.rows); row++ {
		start, end := selectionSpanRange(row, a, b, s.width)
		var text strings.Builder
		col := 0
		for _, span := range s.rows[row].spans {
			width := gotui.StringWidth(span.Text)
			if col+width > start && col < end {
				text.WriteString(span.Text)
			}
			col += width
		}
		lines = append(lines, strings.TrimRight(text.String(), " "))
	}
	return strings.Join(lines, "\n")
}

// Columns are display cells, including both cells of a wide grapheme.
func transcriptRowLinkAt(row transcriptSearchRow, column int) string {
	if column < 0 {
		return ""
	}
	x := 0
	for _, span := range row.spans {
		width := gotui.StringWidth(span.Text)
		if column >= x && column < x+width {
			return span.Link
		}
		x += width
	}
	return ""
}

func (c *chatTUI) handleTranscriptSelection(me gotui.MouseEvent) bool {
	if c.regularMode || c.modelMenuOpen || c.transcriptRegion == nil {
		return false
	}
	s := &c.textSelection
	if s.dragging && !c.selectionCurrent() {
		c.clearTranscriptSelection()
		return true
	}
	if me.Action == gotui.MousePress && me.Button == gotui.MouseLeft {
		if !c.transcriptRegion.ContainsPoint(me.X, me.Y) {
			if s.active || c.selectionClicks.valid {
				c.clearTranscriptSelection()
			}
			return false
		}
		r := c.transcriptRegion.Rect()
		viewWidth, _ := c.transcriptRegion.ViewportSize()
		// go-tui's ViewportSize includes its visible scrollbar column even
		// though child layout reserves that cell when vertical overflow exists.
		_, maxScroll := c.transcriptRegion.MaxScroll()
		if maxScroll > 0 {
			viewWidth--
		}
		if viewWidth < 1 || me.X >= r.X+viewWidth {
			c.selectionClicks = transcriptClickSequence{}
			c.selectionClickSnapshot = transcriptSelection{}
			return false
		} // scrollbar retains its own hit region
		clicks := c.selectionClicks
		if !c.selectionSnapshotCurrent(&c.selectionClickSnapshot) || me.Mod != 0 {
			clicks = transcriptClickSequence{}
		}
		c.clearTranscriptSelection()
		c.selectionClicks = clicks
		s = &c.textSelection
		rows := c.renderedTranscriptRows(r.Width)
		if c.search.active {
			rows = c.search.rows
		}
		if len(rows) == 0 {
			return true
		}
		s.active, s.dragging = true, true
		s.rows = rows
		s.scope = c.selectionScope()
		s.width, s.height = r.Width, r.Height
		s.contentWidth = viewWidth
		if c.app != nil {
			s.terminalWidth, s.terminalHeight = c.app.Size()
		}
		s.source = append([]string(nil), c.transcript...)
		s.expanded = map[string]bool{}
		for k, v := range c.transcriptExpanded {
			s.expanded[k] = v
		}
		s.searchActive, s.query, s.searchMatch = c.search.active, c.search.query, c.search.selected
		// Snapshot actual layout scroll before disabling following; the model's
		// offset may lag a post-layout ScrollToBottom request.
		_, c.transcriptScroll = c.transcriptRegion.ScrollOffset()
		s.pressX, s.pressY = me.X, me.Y
		s.pointerX, s.pointerY = me.X, me.Y
		s.start = c.selectionPoint(me.X, me.Y)
		s.clickKey = rows[s.start.row].blockKey
		s.pressLink = transcriptRowLinkAt(rows[s.start.row], min(s.start.col, s.contentWidth-1))
		s.end = s.start
		// Links and single-click tool controls keep their existing ownership.
		if s.pressLink == "" && s.clickKey == "" && me.Mod == 0 {
			if word, ok := transcriptWordRange(s.start.row, rows[s.start.row], min(s.start.col, s.contentWidth-1)); ok {
				count := c.selectionClicks.next(word, time.Now())
				if count > 1 {
					s.granularity = count - 1
					s.initial = word
					if count == 3 {
						s.initial = transcriptLineRange(s.start.row, rows[s.start.row])
					}
					s.start, s.end = s.initial.start, s.initial.end
					s.moved = s.start != s.end
				}
			}
		}
		s.previousFollow = c.stickToBottom
		c.stickToBottom = false
		return true
	}
	if !s.dragging {
		return false
	}
	switch me.Action {
	case gotui.MouseDrag:
		if me.X != s.pressX || me.Y != s.pressY {
			c.selectionClicks = transcriptClickSequence{}
			c.selectionClickSnapshot = transcriptSelection{}
		}
		s.pointerX, s.pointerY = me.X, me.Y
		s.extend(c.selectionPoint(me.X, me.Y))
		s.moved = s.moved || s.end != s.start
		if c.app != nil {
			c.app.MarkDirty()
		}
		return true
	case gotui.MouseRelease:
		if me.X != s.pressX || me.Y != s.pressY {
			c.selectionClicks = transcriptClickSequence{}
		}
		s.extend(c.selectionPoint(me.X, me.Y))
		s.moved = s.moved || s.end != s.start
		s.dragging = false
		if !s.moved {
			key, follow, linked := s.clickKey, s.previousFollow, s.pressLink != ""
			clicks, snapshot := c.selectionClicks, *s
			c.clearTranscriptSelection()
			if key == "" && !linked {
				c.selectionClicks = clicks
				c.selectionClickSnapshot = snapshot
			}
			c.stickToBottom = follow
			// OSC 8 activation belongs to the terminal client (usually a
			// modified click). Never also toggle the containing tool block.
			if key != "" && !linked {
				c.toggleTranscriptBlock(key)
			}
		} else {
			if c.selectionClicks.valid {
				c.selectionClickSnapshot = *s
			}
			c.copyTranscriptSelection()
		}
		return true
	}
	return false
}

func (c *chatTUI) tickTranscriptSelection() {
	s := &c.textSelection
	if !s.dragging || !s.moved || (s.pointerX == s.pressX && s.pointerY == s.pressY) {
		return
	}
	if !c.selectionCurrent() {
		c.clearTranscriptSelection()
		return
	}
	r := c.transcriptRegion.Rect()
	delta := 0
	if s.pointerY <= r.Y {
		delta = -1
	}
	if s.pointerY >= r.Y+r.Height-1 {
		delta = 1
	}
	if delta == 0 {
		return
	}
	_, offset := c.transcriptRegion.ScrollOffset()
	next := min(c.transcriptMaxScroll(), max(0, offset+delta))
	if next == offset {
		return
	}
	c.setTranscriptPosition(next)
	c.stickToBottom = false
	s.extend(c.selectionPoint(s.pointerX, min(r.Y+r.Height-1, max(r.Y, s.pointerY))))
	s.moved = true
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) selectionNotice(text string) {
	c.textSelection.notice = text
	c.textSelection.noticeUntil = time.Now().Add(4 * time.Second)
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) copyTranscriptSelection() {
	if !c.selectionCurrent() {
		c.clearTranscriptSelection()
		return
	}
	text := c.textSelection.text()
	if text == "" {
		return
	}
	mode, _, _ := c.copyModeFromArgs(nil)
	switch mode {
	case "osc52":
		if len(text) > osc52PayloadLimit {
			c.selectionNotice("Selection too large for OSC 52")
			return
		}
		if err := c.writeOSC52(text); err != nil {
			c.selectionNotice("Copy failed")
		} else {
			c.selectionNotice("Selection copied (OSC 52)")
		}
	case "native", "auto":
		if c.nativeSelectionCopyPending {
			c.selectionNotice("Clipboard busy · retry when ready")
			return
		}
		c.nativeSelectionCopyPending = true
		generation, scope := c.textSelection.generation, c.selectionScope()
		c.selectionNotice("Copying selection…")
		go func() {
			err := c.copyNative(text)
			finish := func() {
				c.nativeSelectionCopyPending = false
				if c.ownsScope(scope) {
					c.finishTranscriptCopy(generation, err)
				}
			}
			if c.app != nil {
				c.app.QueueUpdate(finish)
			} else {
				finish()
			}
		}()
	default:
		c.selectionNotice("Clipboard off · /copy --osc52 --persist")
	}
}

func (c *chatTUI) selectionSeparator(width int) string {
	text := "Selection · drag edges to scroll · Ctrl-C/X copy · Esc clear"
	if time.Now().Before(c.textSelection.noticeUntil) {
		text = c.textSelection.notice
	}
	return truncate(text, width)
}

func (c *chatTUI) renderTranscriptSelectionRows(root *gotui.Element) {
	a, b := c.textSelection.bounds()
	for row, value := range c.textSelection.rows {
		start, end := selectionSpanRange(row, a, b, c.textSelection.width)
		spans := append([]gotui.TextSpan(nil), value.spans...)
		col := 0
		for i := range spans {
			width := gotui.StringWidth(spans[i].Text)
			if c.textSelection.moved && col+width > start && col < end {
				spans[i].Style = spans[i].Style.Background(piText).Foreground(piUserBg)
			}
			col += width
		}
		root.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1), gotui.WithWrap(false), gotui.WithRichText(spans...)))
	}
}

func (c *chatTUI) handleTranscriptEscape() bool {
	c.selectionClicks = transcriptClickSequence{}
	c.selectionClickSnapshot = transcriptSelection{}
	if c.textSelection.active {
		c.clearTranscriptSelection()
		return true
	}
	return c.handleCompactionEscape()
}

func (c *chatTUI) finishTranscriptCopy(generation uint64, err error) {
	if c.textSelection.generation != generation || !c.selectionCurrent() {
		return
	}
	if err != nil {
		c.selectionNotice("Copy failed")
	} else {
		c.selectionNotice("Selection copied")
	}
}
