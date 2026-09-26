package tui

import (
	"fmt"
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

func TestTranscriptLinksSurviveMarkdownSoftWrapping(t *testing.T) {
	target := "https://example.invalid/" + strings.Repeat("long-path-", 18) + "終端?q=β"
	for _, width := range []int{30, 60, 100, 140} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			projected := renderMarkdownTranscript("Gi: ", "# Reference\n\nOpen [manual]("+target+").", width)
			c := sessionTestChat(t)
			c.cfg.AssistantName = "Gi"
			c.outputWidth = width
			c.outputHeight = 24
			c.transcript = projected
			rows := c.renderedTranscriptRows(width)
			var joined strings.Builder
			linkedRows := 0
			for _, row := range rows {
				linked := false
				for _, span := range row.spans {
					if span.Link != "" {
						if span.Link != target {
							t.Fatal("partial target", span.Link)
						}
						joined.WriteString(span.Text)
						linked = true
					}
				}
				if linked {
					linkedRows++
				}
			}
			if joined.String() != target || linkedRows < 2 {
				t.Fatalf("width%d linkedrows%d got%q", width, linkedRows, joined.String())
			}
			c.toggleTranscriptSearch()
			c.refreshTranscriptSearch(width)
			c.updateTranscriptSearchQuery("long-path-long-path")
			if len(c.search.matches) == 0 {
				t.Fatal("linked search missing")
			}
			root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(len(c.search.rows)))
			c.renderTranscriptSearchRows(root)
			buf := gotui.NewBuffer(width, len(c.search.rows))
			root.Render(buf, width, len(c.search.rows))
			linked := false
			for y := 0; y < len(c.search.rows); y++ {
				for x := 0; x < width; x++ {
					cell := buf.Cell(x, y)
					if cell.Link != "" {
						if cell.Link != target {
							t.Fatal("search target changed")
						}
						linked = true
					}
				}
			}
			if !linked {
				t.Fatal("search dropped links")
			}
		})
	}
	for _, target := range []string{"https://u:p@example.invalid/" + strings.Repeat("x", 90), "https://example.invalid/%1b" + strings.Repeat("x", 90), "https://example.invalid/" + strings.Repeat("x", 2049)} {
		if intactTranscriptLinkToken("(" + target + ")") {
			t.Fatal("unsafe exempt from projection", target)
		}
	}
	for _, suffix := range []string{".", ",", ";", ":", "!", "?", "!?"} {
		if !intactTranscriptLinkToken("(" + target + ")" + suffix) {
			t.Fatal("punctuated target split", suffix)
		}
		spans := transcriptLinkSpans("("+target+")"+suffix, gotui.NewStyle())
		if spans[len(spans)-1].Link != "" || !strings.HasSuffix(spans[len(spans)-1].Text, suffix) {
			t.Fatal("punctuation became link", spans)
		}
	}
	for _, token := range []string{"(https://example.invalid/unfinished", "https://example.invalid/", "(javascript:alert)", "prefix(https://example.invalid/)suffix"} {
		if intactTranscriptLinkToken(token) {
			t.Fatal("incomplete token", token)
		}
	}
}

func TestTranscriptWrappedLinksKeepClickOwnershipAndVisibleCopy(t *testing.T) {
	target := "https://example.invalid/" + strings.Repeat("wrapped-", 12) + "end"
	c := sessionTestChat(t)
	c.outputWidth = 40
	c.cfg.AssistantName = "Gi"
	c.input.SetText("kept link draft")
	c.transcript = []string{encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "wrapped-tool", Kind: "tool", Title: "read", Status: "ok"}), "│ (" + target + ")", "│ second", "│ hidden"}
	layoutSearchChat(t, c, 40, 25)
	rows := c.renderedTranscriptRows(40)
	linked := []transcriptPoint{}
	for row, r := range rows {
		for col := 0; col < 40; col++ {
			if transcriptRowLinkAt(r, col) == target {
				linked = append(linked, transcriptPoint{row, col})
				break
			}
		}
	}
	if len(linked) < 2 {
		t.Fatal("no wrapped target")
	}
	p := linked[1]
	rect := c.transcriptRegion.Rect()
	_, offset := c.transcriptRegion.ScrollOffset()
	x, y := rect.X+p.col, rect.Y+p.row-offset
	for range 3 {
		c.handleTranscriptSelection(gotui.MouseEvent{Action: gotui.MousePress, Button: gotui.MouseLeft, X: x, Y: y})
		c.handleTranscriptSelection(gotui.MouseEvent{Action: gotui.MouseRelease, Button: gotui.MouseLeft, X: x, Y: y})
	}
	if c.textSelection.active || c.transcriptExpanded["wrapped-tool"] {
		t.Fatal("wrapped link click changed selection/tool")
	}
	c.handleTranscriptSelection(gotui.MouseEvent{Action: gotui.MousePress, Button: gotui.MouseLeft, X: x, Y: y})
	c.handleTranscriptSelection(gotui.MouseEvent{Action: gotui.MouseDrag, Button: gotui.MouseLeft, X: x + 5, Y: y})
	c.handleTranscriptSelection(gotui.MouseEvent{Action: gotui.MouseRelease, Button: gotui.MouseLeft, X: x + 5, Y: y})
	if copied := c.textSelection.text(); copied == target || gotui.StringWidth(copied) != 5 || strings.Contains(copied, "\x1b") {
		t.Fatal("copy not displayed slice", copied)
	}
	if c.input.Text() != "kept link draft" {
		t.Fatal("draft changed")
	}
}
