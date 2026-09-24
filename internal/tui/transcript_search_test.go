package tui

import (
	"fmt"
	gotui "github.com/grindlemire/go-tui"
	"reflect"
	"strings"
	"testing"
)

func TestTranscriptSearchIndexesVisibleRenderingAndRestoresEditor(t *testing.T) {
	for _, width := range []int{60, 100, 140} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			c := sessionTestChat(t)
			c.outputWidth = width
			c.transcript = nil
			c.appendTranscript("you: ALPHA 中文🙂", "assistant: beta Alpha")
			c.appendTranscriptBlock(transcriptBlockMeta{Key: "tool", Kind: "tool", Title: "read", Status: "ok"}, []string{"visible one", "visible two", "hidden-only", "tail"})
			c.input.SetText("unsent\n中文🙂")
			c.input.cursorPos = 3
			c.input.undoText = "undo"
			c.input.hasUndo = true
			c.input.yankText = "yank"
			c.transcriptScroll, c.stickToBottom = 1, false
			c.toggleTranscriptSearch()
			c.refreshTranscriptSearch(width)
			c.search.input.SetText("alpha")
			if len(c.search.matches) != 2 {
				t.Fatal("case-insensitive rendered matches", c.search.matches)
			}
			first := c.transcriptScroll
			c.moveTranscriptSearch(1)
			if c.transcriptScroll <= first {
				t.Fatal("next match")
			}
			c.moveTranscriptSearch(-1)
			if c.transcriptScroll != first {
				t.Fatal("previous match")
			}
			c.search.input.SetText("hidden-only")
			if len(c.search.matches) != 0 {
				t.Fatal("searched hidden tool output")
			}
			c.transcriptExpanded["tool"] = true
			c.refreshTranscriptSearch(width)
			if len(c.search.matches) != 1 {
				t.Fatal("expanded output not indexed")
			}
			c.search.input.SetText("中文🙂")
			if len(c.search.matches) != 1 {
				t.Fatal("unicode search")
			}
			el := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(30))
			c.renderTranscriptSearchRows(el)
			buf := gotui.NewBuffer(width, 30)
			el.Render(buf, width, 30)
			match := c.search.matches[0]
			if buf.Cell(match.start, match.row).Style.Bg != piText {
				t.Fatal("current match highlight absent")
			}
			c.closeTranscriptSearch()
			if c.input.Text() != "unsent\n中文🙂" || c.input.cursorPos != 3 || c.input.undoText != "undo" || c.input.yankText != "yank" || !c.input.hasUndo {
				t.Fatal("search damaged editor")
			}
			if c.transcriptScroll != 1 || c.stickToBottom {
				t.Fatal("search did not restore reader")
			}
			old := c.search.input
			c.toggleTranscriptSearch()
			if c.search.input != old || c.search.input.Text() != "" {
				t.Fatal("search reopen reused stale mounted input")
			}
			c.switchSession("B")
			if c.search.active || c.input.Text() != "" {
				t.Fatal("search leaked into session")
			}
			c.switchSession("A")
			if c.input.Text() != "unsent\n中文🙂" {
				t.Fatal("origin draft lost")
			}
		})
	}
}

func TestTranscriptSearchRefreshAndPromptJumps(t *testing.T) {
	c := &chatTUI{draftLineIndex: -1}
	c.cfg.AssistantName = "Gi"
	c.ensureInput()
	c.transcript = []string{"you: first", strings.Repeat("wrapped ", 80), "you: second", "Gi: response", "you: third"}
	rows := c.renderedTranscriptRows(60)
	var markers []int
	for i, row := range rows {
		if row.prompt {
			markers = append(markers, i)
		}
	}
	if len(markers) != 3 || markers[1] < 5 {
		t.Fatal("rendered prompt index", markers)
	}
	c.outputWidth = 60
	c.transcriptScroll = markers[2]
	c.input.SetText("untouched")
	c.jumpTranscriptPrompt(-1)
	if c.transcriptScroll != markers[1] {
		t.Fatal("previous prompt", c.transcriptScroll)
	}
	c.jumpTranscriptPrompt(1)
	if c.transcriptScroll != markers[2] || c.input.Text() != "untouched" {
		t.Fatal("next prompt changed draft")
	}
	c.toggleTranscriptSearch()
	c.refreshTranscriptSearch(60)
	c.search.input.SetText("fresh")
	c.appendTranscript("Gi: fresh arrival")
	c.refreshTranscriptSearch(60)
	if len(c.search.matches) != 1 {
		t.Fatal("new native output missing")
	}
	c.refreshTranscriptSearch(100)
	if c.search.width != 100 || len(c.search.matches) != 1 {
		t.Fatal("resize search index")
	}
	c.closeTranscriptSearch()
	c.regularMode = true
	c.toggleTranscriptSearch()
	if c.search.active {
		t.Fatal("regular owns native search")
	}
}

func TestTranscriptSearchRowsMatchActualScrollLayout(t *testing.T) {
	for _, width := range []int{60, 100, 140} {
		c := &chatTUI{draftLineIndex: -1, transcriptRef: gotui.NewRef()}
		c.cfg.AssistantName = "Gi"
		c.transcript = []string{"you: first", strings.Repeat("wrapped ", 70), "you: second", "Gi: final"}
		c.appendTranscriptBlock(transcriptBlockMeta{Key: "tool", Kind: "tool", Title: "shell", Status: "ok"}, []string{"long " + strings.Repeat("output ", 40), "tail", "hidden"})
		el := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(8), gotui.WithScrollable(gotui.ScrollVertical))
		for _, block := range c.buildTranscriptRenderableBlocks(c.visibleTranscript()) {
			el.AddChild(c.renderTranscriptBlock(block))
		}
		el.Render(gotui.NewBuffer(width, 8), width, 8)
		c.transcriptRef.Set(el)
		rows := c.renderedTranscriptRows(width)
		_, contentHeight := el.ContentSize()
		if len(rows) != contentHeight {
			t.Fatalf("%d: indexed %d != rendered %d", width, len(rows), contentHeight)
		}
		var promptRows []int
		for i, row := range rows {
			if row.prompt {
				promptRows = append(promptRows, i)
			}
		}
		c.outputWidth = width
		c.setTranscriptPosition(promptRows[1])
		c.jumpTranscriptPrompt(-1)
		if c.transcriptScroll != promptRows[0] {
			t.Fatal("prompt position not synchronized")
		}
	}
}

func TestTranscriptSearchOccurrencesDisplayCellsAndNavigation(t *testing.T) {
	for _, width := range []int{60, 100, 140} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			c := sessionTestChat(t)
			c.outputWidth = width
			c.ensureInput()
			c.input.SetText("draft 中文")
			c.input.cursorPos = 3
			c.transcript = []string{"sys: İ界 e\u0301 İ界 e\u0301 aaAAA [x](https://example.invalid/aaa)"}
			c.toggleTranscriptSearch()
			c.refreshTranscriptSearch(width)
			cases := []struct {
				query string
				want  []transcriptSearchMatch
			}{
				{"i", []transcriptSearchMatch{{0, 5, 6}, {0, 11, 12}, {0, 43, 44}, {0, 48, 49}}},
				{"界", []transcriptSearchMatch{{0, 6, 8}, {0, 12, 14}}},
				{"e\u0301", []transcriptSearchMatch{{0, 9, 10}, {0, 15, 16}}},
				{"\u0301", []transcriptSearchMatch{{0, 9, 10}, {0, 15, 16}}},
				{"aa", []transcriptSearchMatch{{0, 17, 19}, {0, 19, 21}, {0, 51, 53}}},
			}
			// The short prefix remains on one row even at the smallest width.
			for _, tc := range cases {
				c.updateTranscriptSearchQuery(tc.query)
				if !reflect.DeepEqual(c.search.matches, tc.want) {
					t.Fatalf("%q: %#v want %#v", tc.query, c.search.matches, tc.want)
				}
			}
			c.updateTranscriptSearchQuery("界")
			first := c.search.matches[0]
			c.moveTranscriptSearch(1)
			second := c.search.matches[1]
			if c.search.selected != 1 || first.row != second.row {
				t.Fatal("same-row next skipped")
			}
			el := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(4))
			c.renderTranscriptSearchRows(el)
			buf := gotui.NewBuffer(width, 4)
			el.Render(buf, width, 4)
			if buf.Cell(second.start, 0).Style.Bg != piText || buf.Cell(second.start+1, 0).Style.Bg != piText {
				t.Fatal("wide active match incomplete")
			}
			if buf.Cell(first.start, 0).Style.Bg != piUserBg {
				t.Fatal("inactive occurrence not distinguished")
			}
			if buf.Cell(0, 0).Style != c.search.rows[0].spans[0].Style {
				t.Fatal("unmatched cells restyled")
			}
			if buf.Cell(29, 0).Link != "https://example.invalid/aaa" {
				t.Fatal("link metadata lost")
			}
			c.updateTranscriptSearchQuery("example")
			linked := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(4))
			c.renderTranscriptSearchRows(linked)
			linked.Render(buf, width, 4)
			match := c.search.matches[0]
			if buf.Cell(match.start, 0).Link != "https://example.invalid/aaa" || buf.Cell(match.start, 0).Style.Bg != piText {
				t.Fatal("highlighted URL lost link metadata")
			}
			c.updateTranscriptSearchQuery("界")
			c.moveTranscriptSearch(1)
			// Appending output preserves the second occurrence on this unchanged row.
			c.transcript = append(c.transcript, "sys: later 界")
			c.refreshTranscriptSearch(width)
			if c.search.selected != 1 || c.search.matches[1] != second {
				t.Fatal("append reset occurrence")
			}
			c.moveTranscriptSearch(-1)
			if c.search.selected != 0 {
				t.Fatal("previous failed")
			}
			c.moveTranscriptSearch(-1)
			if c.search.selected != 2 {
				t.Fatal("reverse wrap failed")
			}
			c.moveTranscriptSearch(1)
			if c.search.selected != 0 {
				t.Fatal("forward wrap failed")
			}
			c.closeTranscriptSearch()
			if c.input.Text() != "draft 中文" || c.input.cursorPos != 3 {
				t.Fatal("editor changed")
			}
		})
	}
}

func TestTranscriptSearchOccurrencesIgnorePaddingAndCrossRows(t *testing.T) {
	c := sessionTestChat(t)
	c.outputWidth = 60
	c.transcript = []string{"sys: AAAAA", "sys: z"}
	c.toggleTranscriptSearch()
	c.refreshTranscriptSearch(60)
	c.updateTranscriptSearchQuery("aa")
	if len(c.search.matches) != 2 {
		t.Fatal("non-overlapping matches", c.search.matches)
	}
	c.updateTranscriptSearchQuery("     ")
	if len(c.search.matches) != 0 {
		t.Fatal("phantom right padding matches")
	}
	c.updateTranscriptSearchQuery("A\nsys:")
	if len(c.search.matches) != 0 {
		t.Fatal("cross-row result advertised")
	}
	c.updateTranscriptSearchQuery("")
	if len(c.search.matches) != 0 || c.search.selected != -1 {
		t.Fatal("empty query did not reset")
	}
}
