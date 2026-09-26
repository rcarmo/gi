package tui

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	gotui "github.com/grindlemire/go-tui"
)

func selectionFixture(t *testing.T, width, height int) *chatTUI {
	t.Helper()
	c := sessionTestChat(t)
	c.outputWidth = width
	c.cfg.AssistantName = "Gi"
	c.transcript = nil
	for i := 0; i < 40; i++ {
		c.appendTranscript(fmt.Sprintf("you: row-%02d 中文🙂 e\u0301", i))
	}
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(height), gotui.WithScrollable(gotui.ScrollVertical))
	c.transcriptRef = gotui.NewRef()
	c.transcriptRef.Set(root)
	c.transcriptRegion = root
	for _, b := range c.buildTranscriptRenderableBlocks(c.visibleTranscript()) {
		root.AddChild(c.renderTranscriptBlock(b))
	}
	root.Render(gotui.NewBuffer(width, height), width, height)
	c.input.SetText("newer draft")
	c.input.cursorPos = 3
	return c
}

func TestTranscriptSelectionWideCellsReverseDragAndCopyPolicy(t *testing.T) {
	for _, size := range [][2]int{{60, 18}, {100, 22}, {140, 36}} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			c := selectionFixture(t, size[0], size[1]-5)
			send := func(action gotui.MouseAction, x, y int) {
				if !c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: action, X: x, Y: y}) {
					t.Fatal("unhandled selection event")
				}
			}
			var clipboard bytes.Buffer
			c.osc52Writer = &clipboard
			send(gotui.MousePress, 1, 1)
			send(gotui.MouseDrag, 27, 4)
			send(gotui.MouseRelease, 27, 4)
			selected := c.textSelection.text()
			if !strings.Contains(selected, "row-00 中文🙂 e\u0301\n\n\n you: row-01 中文🙂 e\u0301") {
				t.Fatal("wide/combining rows damaged", selected)
			}
			if clipboard.Len() != 0 || !strings.Contains(c.textSelection.notice, "Clipboard off") {
				t.Fatal("clipboard opt-out ignored")
			}
			c.cfg.TUIClipboardMode = "osc52"
			c.copyTranscriptSelection()
			seq, err := osc52Sequence(selected)
			if err != nil || clipboard.String() != seq {
				t.Fatal("OSC 52 selection mismatch")
			}
			send(gotui.MousePress, 27, 4)
			send(gotui.MouseDrag, 1, 1)
			send(gotui.MouseRelease, 1, 1)
			if c.textSelection.text() != selected {
				t.Fatal("reverse drag differs")
			}
			if c.input.Text() != "newer draft" || c.input.cursorPos != 3 {
				t.Fatal("selection changed editor")
			}
			el := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(size[0]), gotui.WithHeight(size[1]))
			c.renderTranscriptSelectionRows(el)
			buf := gotui.NewBuffer(size[0], size[1])
			el.Render(buf, size[0], size[1])
			if buf.Cell(1, 1).Style.Bg != piText {
				t.Fatal("selection highlight missing")
			}
			c.clearTranscriptSelection()
			send(gotui.MousePress, 14, 1)
			send(gotui.MouseDrag, 15, 1)
			send(gotui.MouseRelease, 15, 1)
			if c.textSelection.text() != "中" {
				t.Fatal("half-wide selection split glyph", c.textSelection.text())
			}
		})
	}
}
func TestTranscriptSelectionInvalidationAndEdgeScroll(t *testing.T) {
	c := selectionFixture(t, 60, 13)
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 0, Y: 2})
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MouseDrag, X: 20, Y: 15})
	for i := 0; i < 4; i++ {
		c.tickTranscriptSelection()
	}
	if c.transcriptScroll != 4 || c.textSelection.end.row < 16 || c.stickToBottom {
		t.Fatal("edge drag did not advance", c.transcriptScroll, c.textSelection.end)
	}
	c.transcript = append(c.transcript, "new output")
	c.tickTranscriptSelection()
	if c.textSelection.active {
		t.Fatal("new output retained stale selection")
	}
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 0, Y: 1})
	c.validateTranscriptSelection(100, 13)
	if c.textSelection.active {
		t.Fatal("resize retained stale selection")
	}
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 0, Y: 1})
	c.switchSession("B")
	if c.textSelection.active {
		t.Fatal("session selection leaked")
	}
}
func TestTranscriptSelectionClickStillTogglesTool(t *testing.T) {
	c := selectionFixture(t, 60, 13)
	c.transcript = nil
	c.transcriptBlockRefs = nil
	c.appendTranscriptBlock(transcriptBlockMeta{Key: "tool", Kind: "bash", Title: "shell", Status: "ok"}, strings.Split(strings.Repeat("line\n", 15), "\n"))
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(60), gotui.WithHeight(13), gotui.WithScrollable(gotui.ScrollVertical))
	for _, b := range c.buildTranscriptRenderableBlocks(c.transcript) {
		root.AddChild(c.renderTranscriptBlock(b))
	}
	root.Render(gotui.NewBuffer(60, 13), 60, 13)
	c.transcriptRegion = root
	c.transcriptRef.Set(root)
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 3, Y: 1})
	c.transcriptBlockRefs = nil // rendering the selection replaces the normal tree
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MouseRelease, X: 3, Y: 1})
	if !c.transcriptExpanded["tool"] || c.textSelection.active {
		t.Fatal("click failed to toggle")
	}
}

func TestTranscriptSelectionClipboardFailuresAndLimits(t *testing.T) {
	c := selectionFixture(t, 60, 13)
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 0, Y: 1})
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MouseDrag, X: 24, Y: 2})
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MouseRelease, X: 24, Y: 2})
	c.cfg.TUIClipboardMode = "osc52"
	c.osc52Writer = failingSelectionWriter{}
	c.copyTranscriptSelection()
	if c.textSelection.notice != "Copy failed" || !c.textSelection.active {
		t.Fatal("failure discarded selection")
	}
	c.textSelection.rows = []transcriptSearchRow{{spans: []gotui.TextSpan{{Text: strings.Repeat("x", osc52PayloadLimit+1)}}}}
	c.textSelection.start = transcriptPoint{0, 0}
	c.textSelection.end = transcriptPoint{0, osc52PayloadLimit + 1}
	c.textSelection.width = osc52PayloadLimit + 1
	// Test the existing encoder's no-partial-write boundary independently of UI width.
	if _, err := osc52Sequence(c.textSelection.text()); err == nil {
		t.Fatal("oversized selection not rejected")
	}
}

type failingSelectionWriter struct{}

func (failingSelectionWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("terminal disconnected")
}

func TestTranscriptSelectionSearchViewChangesInvalidateCopy(t *testing.T) {
	c := selectionFixture(t, 60, 13)
	c.toggleTranscriptSearch()
	c.refreshTranscriptSearch(60)
	c.search.input.SetText("row-01")
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 0, Y: 1})
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MouseDrag, X: 20, Y: 2})
	c.search.input.SetText("row-02")
	var clipboard bytes.Buffer
	c.osc52Writer = &clipboard
	c.cfg.TUIClipboardMode = "osc52"
	c.copyTranscriptSelection()
	if c.textSelection.active || clipboard.Len() != 0 {
		t.Fatal("changed query copied wrong rows")
	}
}

func TestTranscriptSelectionLateCopyCannotOverwriteNewSelection(t *testing.T) {
	c := selectionFixture(t, 60, 13)
	start := func() {
		c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 0, Y: 1})
		c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MouseDrag, X: 20, Y: 2})
	}
	start()
	old := c.textSelection.generation
	scope := c.selectionScope()
	c.clearTranscriptSelection()
	start()
	c.selectionNotice("new selection")
	c.applySessionCompletion(scope, func() { c.finishTranscriptCopy(old, fmt.Errorf("old error")) })
	if c.textSelection.notice != "new selection" {
		t.Fatal("old copy mutated new selection")
	}
	current := c.textSelection.generation
	c.transcript = append(c.transcript, "changed")
	c.finishTranscriptCopy(current, nil)
	if c.textSelection.notice != "new selection" {
		t.Fatal("changed output accepted old success")
	}
}

func TestTranscriptSelectionStationaryEdgePressDoesNotScroll(t *testing.T) {
	c := selectionFixture(t, 60, 13)
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 2, Y: 12})
	for i := 0; i < 10; i++ {
		c.tickTranscriptSelection()
	}
	if c.transcriptScroll != 0 || c.textSelection.moved {
		t.Fatal("stationary click became edge drag")
	}
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MouseRelease, X: 2, Y: 12})
	if c.textSelection.active {
		t.Fatal("stationary click retained selection")
	}
}

func TestTranscriptSelectionNativeCopyHasSinglePendingSlot(t *testing.T) {
	c := selectionFixture(t, 60, 13)
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 1, Y: 1})
	c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MouseDrag, X: 25, Y: 4})
	c.nativeSelectionCopyPending = true
	c.cfg.TUIClipboardMode = "native"
	c.copyTranscriptSelection()
	if c.textSelection.notice != "Clipboard busy · retry when ready" || !c.nativeSelectionCopyPending {
		t.Fatal("parallel clipboard writes permitted")
	}
	c.clearTranscriptSelection()
	if !c.nativeSelectionCopyPending {
		t.Fatal("selection reset freed in-flight clipboard slot")
	}
}

func TestTranscriptSelectionIncludesFinalContentCellWithoutScrollbarPress(t *testing.T) {
	for _, width := range []int{60, 100, 140} {
		for _, scrollbar := range []bool{false, true} {
			for _, suffix := range []string{"Z", "界", "e\u0301"} {
				t.Run(fmt.Sprintf("%d-scroll%v-%s", width, scrollbar, suffix), func(t *testing.T) {
					c := sessionTestChat(t)
					c.outputWidth = width
					c.cfg.TUIClipboardMode = "osc52"
					var clipboard bytes.Buffer
					c.osc52Writer = &clipboard
					contentWidth := width
					if scrollbar {
						contentWidth--
					}
					text := strings.Repeat("a", contentWidth-gotui.StringWidth(suffix)) + suffix
					c.transcript = []string{text}
					if scrollbar {
						for i := 0; i < 20; i++ {
							c.transcript = append(c.transcript, text)
						}
					}
					layoutSearchChat(t, c, width, 6)
					r := c.transcriptRegion.Rect()
					view, _ := c.transcriptRegion.ViewportSize()
					_, overflow := c.transcriptRegion.MaxScroll()
					if overflow > 0 {
						view--
					}
					if view != contentWidth {
						t.Fatal(view, contentWidth)
					}
					send := func(action gotui.MouseAction, x int) bool {
						return c.handleTranscriptSelection(gotui.MouseEvent{Action: action, Button: gotui.MouseLeft, X: r.X + x, Y: r.Y + 2})
					}
					y := r.Y
					if scrollbar {
						y = r.Y + 2
					}
					point := func(action gotui.MouseAction, x int) bool {
						return c.handleTranscriptSelection(gotui.MouseEvent{Action: action, Button: gotui.MouseLeft, X: r.X + x, Y: y})
					}
					point(gotui.MousePress, 0)
					point(gotui.MouseDrag, contentWidth-1)
					point(gotui.MouseRelease, contentWidth-1)
					if got := c.textSelection.text(); got != text {
						t.Fatalf("forward %q want %q", got, text)
					}
					sequence, err := osc52Sequence(text)
					if err != nil || clipboard.String() != sequence {
						t.Fatal("clipboard differs from selected cells")
					}
					c.clearTranscriptSelection()
					clipboard.Reset()
					point(gotui.MousePress, contentWidth-1)
					point(gotui.MouseDrag, 0)
					point(gotui.MouseRelease, 0)
					if got := c.textSelection.text(); got != text {
						t.Fatalf("reverse %q want %q", got, text)
					}
					if clipboard.String() != sequence {
						t.Fatal("reverse release did not copy selected cells")
					}
					c.clearTranscriptSelection()
					clipboard.Reset()
					point(gotui.MousePress, contentWidth-1)
					point(gotui.MouseRelease, contentWidth-1)
					if c.textSelection.active || clipboard.Len() != 0 {
						t.Fatal("stationary edge click selected or copied text")
					}
					// A new press in the actual scrollbar remains go-tui's responsibility.
					if scrollbar && send(gotui.MousePress, contentWidth) {
						t.Fatal("scrollbar press consumed")
					}
					c.clearTranscriptSelection()
					point(gotui.MousePress, 0)
					point(gotui.MouseDrag, 5)
					point(gotui.MouseRelease, 5)
					if got := c.textSelection.text(); got != "aaaaa" {
						t.Fatal("interior half-open convention changed", got)
					}
					c.clearTranscriptSelection()
					point(gotui.MousePress, 0)
					point(gotui.MouseDrag, width+10)
					point(gotui.MouseRelease, width+10)
					if got := c.textSelection.text(); got != text {
						t.Fatal("outside drag did not clamp", got)
					}
				})
			}
		}
	}
}

func TestTranscriptWordSegmentsAndClickSequence(t *testing.T) {
	c := &chatTUI{transcript: []string{"sys: file-name/path.txt e\u0301clair 中文 🙂 end"}, outputHeight: 20}
	rows := c.renderedTranscriptRows(100)
	row := rows[0]
	selected := func(col int) string {
		r, ok := transcriptWordRange(0, row, col)
		if !ok {
			return ""
		}
		s := transcriptSelection{active: true, moved: true, rows: rows, width: 100, start: r.start, end: r.end}
		return s.text()
	}
	for _, check := range []struct {
		col  int
		want string
	}{{6, "file-name/path.txt"}, {11, "file-name/path.txt"}, {24, "e\u0301clair"}, {36, "🙂"}, {0, "sys"}, {3, ":"}} {
		if got := selected(check.col); got != check.want {
			t.Fatalf("col%d got%q want%q row%q", check.col, got, check.want, row.text)
		}
	}
	var seq transcriptClickSequence
	r, _ := transcriptWordRange(0, row, 6)
	now := time.Now()
	for _, want := range []int{1, 2, 3, 1} {
		if got := seq.next(r, now); got != want {
			t.Fatal(got, want)
		}
		now = now.Add(100 * time.Millisecond)
	}
	if seq.next(r, now.Add(time.Second)) != 1 {
		t.Fatal("expired")
	}
	r.start.row++
	r.end.row++
	if seq.next(r, now.Add(time.Second)) != 1 {
		t.Fatal("new row")
	}
}

func TestTranscriptMultiClickWordLineDragAndInvalidation(t *testing.T) {
	c := selectionFixture(t, 60, 18)
	mouse := func(c *chatTUI, action gotui.MouseAction, x, y int) {
		t.Helper()
		if !c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: action, X: x, Y: y}) {
			t.Fatal("unhandled mouse")
		}
	}
	before := *c.input
	click := func(x, y int) {
		mouse(c, gotui.MousePress, x, y)
		// Real go-tui may redraw between press and release.
		c.validateTranscriptSelection(60, 18)
		mouse(c, gotui.MouseRelease, x, y)
		c.validateTranscriptSelection(60, 18)
	}
	// Same word, different cell: 500ms sequence follows word bounds, not pixels.
	click(7, 1)
	click(8, 1)
	if got := c.textSelection.text(); got != "row-00" || c.textSelection.granularity != 1 {
		t.Fatalf("double:%q %#v", got, c.selectionClicks)
	}
	if !strings.Contains(c.textSelection.notice, "Clipboard off") {
		t.Fatal("clipboard policy")
	}
	click(8, 1)
	if got := c.textSelection.text(); got != c.textSelection.rows[1].text || c.textSelection.granularity != 2 {
		t.Fatalf("triple:%q", got)
	}
	c.clearTranscriptSelection()
	click(7, 7)
	mouse(c, gotui.MousePress, 8, 7)
	mouse(c, gotui.MouseDrag, 8, 1)
	mouse(c, gotui.MouseRelease, 8, 1)
	if got := c.textSelection.text(); !strings.HasPrefix(got, "row-00") || !strings.HasSuffix(got, "row-02") {
		t.Fatalf("reverse word drag:%q", got)
	}
	if before.text != c.input.text || before.cursorPos != c.input.cursorPos || before.undoText != c.input.undoText || before.yankText != c.input.yankText {
		t.Fatal("editor mutated")
	}
	c.clearTranscriptSelection()
	click(7, 1)
	c.transcript = append(c.transcript, "sys: new output")
	click(7, 1)
	if c.textSelection.active {
		t.Fatal("stale output counted as second click")
	}
	c.clearTranscriptSelection()
	click(7, 1)
	c.handleTranscriptEscape()
	click(7, 1)
	if c.textSelection.active {
		t.Fatal("Escape retained multiclick sequence")
	}
	c.clearTranscriptSelection()
	click(7, 1)
	c.selectionClicks.at = time.Now().Add(-time.Second)
	click(7, 1)
	if c.textSelection.active {
		t.Fatal("slow clicks became double")
	}
	c.clearTranscriptSelection()
	click(7, 1)
	mouse(c, gotui.MousePress, 8, 1)
	c.tickTranscriptSelection()
	if c.transcriptScroll != 0 {
		t.Fatal("stationary word click scrolled")
	}
}

func TestTranscriptMultiClickSequenceOwnership(t *testing.T) {
	for _, kind := range []string{"session", "generation", "query", "resize", "outside", "wheel", "modifier", "drag"} {
		t.Run(kind, func(t *testing.T) {
			c := selectionFixture(t, 60, 13)
			event := func(action gotui.MouseAction, x, y int) {
				c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: action, X: x, Y: y})
			}
			click := func() { event(gotui.MousePress, 8, 1); event(gotui.MouseRelease, 8, 1) }
			click()
			if !c.selectionClicks.valid {
				t.Fatal("first click missing")
			}
			switch kind {
			case "session":
				c.sessionID = "different"
			case "generation":
				c.sessionGeneration++
			case "query":
				c.search.query = "changed"
			case "resize":
				c.validateTranscriptSelection(58, 13)
			case "outside":
				event(gotui.MousePress, 70, 20)
			case "wheel":
				c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseWheelDown, Action: gotui.MousePress, X: 8, Y: 1})
			case "modifier":
				c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 8, Y: 1, Mod: gotui.ModShift})
				event(gotui.MouseRelease, 8, 1)
			case "drag":
				event(gotui.MousePress, 8, 1)
				event(gotui.MouseDrag, 18, 1)
				event(gotui.MouseRelease, 18, 1)
			}
			click()
			if c.textSelection.granularity != 0 {
				t.Fatalf("%s retained multiclick", kind)
			}
		})
	}
}

func TestTranscriptWordSelectionWideHalfAndEdgeDrag(t *testing.T) {
	c := selectionFixture(t, 60, 13)
	event := func(action gotui.MouseAction, x, y int) {
		c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: action, X: x, Y: y})
	}
	// First and second cells of the same CJK word segment identify one click unit.
	event(gotui.MousePress, 13, 1)
	event(gotui.MouseRelease, 13, 1)
	event(gotui.MousePress, 14, 1)
	event(gotui.MouseRelease, 14, 1)
	if got := c.textSelection.text(); got != "中" {
		t.Fatalf("wide selection:%q", got)
	}
	c.clearTranscriptSelection()
	event(gotui.MousePress, 8, 4)
	event(gotui.MouseRelease, 8, 4)
	event(gotui.MousePress, 8, 4)
	event(gotui.MouseDrag, 8, 12)
	for range 3 {
		c.tickTranscriptSelection()
	}
	if c.transcriptScroll < 1 || c.textSelection.granularity != 1 {
		t.Fatal("word edge scroll")
	}
	if c.textSelection.end.col < 1 {
		t.Fatal("word edge lost range")
	}
}
