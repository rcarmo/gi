package tui

import (
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
)

func TestTranscriptLinksKeepVisibleTextAndRejectUnsafeTargets(t *testing.T) {
	for _, raw := range []string{"javascript:alert(1)", "file:///tmp/x", "https://u:p@example.com/x", "https://example.com/a\\b", "https://example.com/%1b]8;", "https://example.com/\x1b", "https://example.com/" + strings.Repeat("a", 2049)} {
		if safeTranscriptURL(raw) {
			t.Fatalf("unsafe %q", raw)
		}
	}
	text := "see (https://example.invalid/docs?q=β) then (http://example.invalid)"
	spans := transcriptLinkSpans(text, gotui.NewStyle())
	var joined strings.Builder
	links := 0
	for _, s := range spans {
		joined.WriteString(s.Text)
		if s.Link != "" {
			links++
		}
	}
	if joined.String() != text || links != 2 {
		t.Fatal(joined.String(), spans)
	}
	for _, text := range []string{"https://example.invalid/clipped", "(https://example.invalid/unfinished", "(https://name:secret@example.invalid/x)"} {
		for _, s := range transcriptLinkSpans(text, gotui.NewStyle()) {
			if s.Link != "" {
				t.Fatal(text, s)
			}
		}
	}
	row := transcriptSearchRow{spans: []gotui.TextSpan{{Text: "界", Link: "https://example.invalid"}, {Text: "x"}}}
	if transcriptRowLinkAt(row, 0) == "" || transcriptRowLinkAt(row, 1) == "" || transcriptRowLinkAt(row, 2) != "" {
		t.Fatal("wide-cell hit testing")
	}
}

func TestTranscriptLinkPrecedenceSearchAndVisibleTextSelection(t *testing.T) {
	for _, size := range [][2]int{{60, 18}, {100, 22}, {140, 36}} {
		c := sessionTestChat(t)
		c.outputWidth = size[0]
		c.input.SetText("draft β")
		c.input.cursorPos = len([]rune(c.input.Text()))
		c.transcript = []string{encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "tool-link", Kind: "tool", Title: "read", Status: "ok"}), "│ reference (https://example.invalid/docs)", "│ second output", "│ third output"}
		layoutSearchChat(t, c, size[0], size[1])
		rows := c.renderedTranscriptRows(c.currentContentWidth())
		linkRow, col := -1, 0
		for i, row := range rows {
			for j := 0; j < c.currentContentWidth(); j++ {
				if transcriptRowLinkAt(row, j) == "https://example.invalid/docs" {
					linkRow, col = i, j
					break
				}
			}
			if linkRow >= 0 {
				break
			}
		}
		if linkRow < 0 {
			t.Fatal("rendered link metadata missing", rows)
		}
		r := c.transcriptRegion.Rect()
		_, offset := c.transcriptRegion.ScrollOffset()
		x, y := r.X+col, r.Y+linkRow-offset
		send := func(action gotui.MouseAction, x, y int) {
			if !c.handleTranscriptSelection(gotui.MouseEvent{Action: action, Button: gotui.MouseLeft, X: x, Y: y}) {
				t.Fatal("mouse not handled")
			}
		}
		send(gotui.MousePress, x, y)
		send(gotui.MouseRelease, x, y)
		if c.transcriptExpanded["tool-link"] || c.textSelection.active {
			t.Fatal("link toggled tool block or left selection")
		}
		send(gotui.MousePress, x, y)
		send(gotui.MouseDrag, x+5, y)
		send(gotui.MouseRelease, x+5, y)
		if c.textSelection.text() != "https" {
			t.Fatalf("selection copied hidden URL/markup: %q", c.textSelection.text())
		}
		if c.transcriptExpanded["tool-link"] || c.input.Text() != "draft β" || c.input.cursorPos != len([]rune(c.input.Text())) {
			t.Fatal("drag altered block/editor")
		}
		c.clearTranscriptSelection()
		c.toggleTranscriptSearch()
		c.refreshTranscriptSearch(size[0])
		c.search.input.SetText("reference")
		layoutSearchChat(t, c, size[0], size[1])
		linked := false
		for _, row := range c.search.rows {
			for _, span := range row.spans {
				if span.Link == "https://example.invalid/docs" {
					linked = true
				}
			}
		}
		if !linked {
			t.Fatal("search dropped link target")
		}
		// A stationary click outside the link retains the existing block action.
		c.closeTranscriptSearch()
		layoutSearchChat(t, c, size[0], size[1])
		r = c.transcriptRegion.Rect()
		_, offset = c.transcriptRegion.ScrollOffset()
		send(gotui.MousePress, r.X+1, r.Y+linkRow-offset)
		send(gotui.MouseRelease, r.X+1, r.Y+linkRow-offset)
		if !c.transcriptExpanded["tool-link"] {
			t.Fatal("ordinary block click lost")
		}
	}
}

func layoutSearchChat(t *testing.T, c *chatTUI, width, height int) {
	t.Helper()
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(height), gotui.WithScrollable(gotui.ScrollVertical))
	c.transcriptRef = gotui.NewRef()
	c.transcriptRef.Set(root)
	c.transcriptRegion = root
	if c.search.active {
		c.renderTranscriptSearchRows(root)
	} else {
		for _, block := range c.buildTranscriptRenderableBlocks(c.visibleTranscript()) {
			root.AddChild(c.renderTranscriptBlock(block))
		}
	}
	root.Render(gotui.NewBuffer(width, height), width, height)
}
