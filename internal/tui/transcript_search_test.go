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
				{"i", []transcriptSearchMatch{{row: 0, start: 5, end: 6}, {row: 0, start: 11, end: 12}, {row: 0, start: 43, end: 44}, {row: 0, start: 48, end: 49}}},
				{"界", []transcriptSearchMatch{{row: 0, start: 6, end: 8}, {row: 0, start: 12, end: 14}}},
				{"e\u0301", []transcriptSearchMatch{{row: 0, start: 9, end: 10}, {row: 0, start: 15, end: 16}}},
				{"\u0301", []transcriptSearchMatch{{row: 0, start: 9, end: 10}, {row: 0, start: 15, end: 16}}},
				{"aa", []transcriptSearchMatch{{row: 0, start: 17, end: 19}, {row: 0, start: 19, end: 21}, {row: 0, start: 51, end: 53}}},
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
			if c.search.selected != 1 || !reflect.DeepEqual(c.search.matches[1], second) {
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

func TestTranscriptSearchAcrossSoftWrapsOnly(t *testing.T) {
	for _, width := range []int{18, 60, 100, 140} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			c := &chatTUI{outputWidth: width, outputHeight: 18, inputActive: true, stickToBottom: true}
			c.ensureInput()
			c.input.SetText("retained 中文🙂 draft")
			c.input.cursorPos = 4
			c.transcript = []string{"sys: " + strings.Repeat("a", width*2) + "界e\u0301suffix", "sys: hard\nbreak", "sys: left", "sys: right"}
			c.toggleTranscriptSearch()
			c.refreshTranscriptSearch(width)
			var runs []transcriptSearchRun
			for _, r := range c.search.rows {
				runs = append(runs, r.wrapped...)
			}
			if len(runs) != 1 {
				t.Fatalf("width%d runs=%d rows=%#v", width, len(runs), c.search.rows)
			}
			query := strings.Repeat("a", width*2) + "界e\u0301suffix"
			c.updateTranscriptSearchQuery(query)
			if len(c.search.matches) != 1 || len(c.search.matches[0].continuation) < 1 {
				t.Fatalf("missing cross wrap %#v", c.search.matches)
			}
			el := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(len(c.search.rows)))
			c.renderTranscriptSearchRows(el)
			buf := gotui.NewBuffer(width, len(c.search.rows))
			el.Render(buf, width, len(c.search.rows))
			match := c.search.matches[0]
			parts := append([]transcriptSearchCell{{match.row, match.start, match.end}}, match.continuation...)
			for _, p := range parts {
				if buf.Cell(p.start, p.row).Style.Bg != piText {
					t.Fatalf("unhighlighted segment %#v", p)
				}
			}
			for _, q := range []string{"hardbreak", "hard break", "leftsys: right", "left right"} {
				c.updateTranscriptSearchQuery(q)
				if len(c.search.matches) != 0 {
					t.Fatal("joined hard boundary", q)
				}
			}
			c.updateTranscriptSearchQuery("aa")
			if len(c.search.matches) != width {
				t.Fatalf("non-overlapping wrapped count%d want%d", len(c.search.matches), width)
			}
			c.closeTranscriptSearch()
			if c.input.Text() != "retained 中文🙂 draft" || c.input.cursorPos != 4 || !c.stickToBottom {
				t.Fatal("editor/follow changed")
			}
		})
	}
}

func TestTranscriptSearchWrappedWordsLinksAndCollapsedBoundaries(t *testing.T) {
	c, _, _ := modelTestChat(t)
	c.outputHeight = 18
	query := "visible phrase 中文🙂 e\u0301 ending"
	c.transcript = []string{"sys: " + query, "sys: [link](https://example.invalid/" + strings.Repeat("long", 25) + ")", "sys: hidden", "sys: neighbour"}
	c.toggleTranscriptSearch()
	for _, width := range []int{18, 24, 60, 140} {
		c.outputWidth = width
		c.refreshTranscriptSearch(width)
		c.updateTranscriptSearchQuery(query)
		if len(c.search.matches) != 1 {
			t.Fatalf("width %d missing phrase: %#v", width, c.search.matches)
		}
		if width < 30 && len(c.search.matches[0].continuation) == 0 {
			t.Fatal("expected wrapped phrase")
		}
	}
	c.transcript = []string{encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "wrap-tool", Kind: "tool", Title: "shell command", Status: "ok"}), "│ visible top", "│ second line", "│ " + strings.Repeat("x", 80) + "secret"}
	c.outputWidth = 18
	c.refreshTranscriptSearch(18)
	c.updateTranscriptSearchQuery(strings.Repeat("x", 80) + "secret")
	if len(c.search.matches) != 0 {
		t.Fatal("hidden wrapped text matched")
	}
	c.transcriptExpanded = map[string]bool{"wrap-tool": true}
	c.refreshTranscriptSearch(18)
	if len(c.search.matches) != 1 || len(c.search.matches[0].continuation) == 0 {
		t.Fatal("expanded wrap not matched", c.search.matches)
	}
	// Renderer-source mismatch is rejected rather than guessed (prewrapped or
	// clipped content must never acquire phantom continuations).
	el := gotui.New(gotui.WithText("abcdef"), gotui.WithWidth(3), gotui.WithHeight(2))
	root := gotui.New(gotui.WithWidth(3), gotui.WithHeight(2))
	root.AddChild(el)
	root.Render(gotui.NewBuffer(3, 2), 3, 2)
	rows := []transcriptSearchRow{{spans: []gotui.TextSpan{{Text: "abc"}}}, {spans: []gotui.TextSpan{{Text: "xyz"}}}}
	if len(transcriptWrapRuns(el, rows)) != 0 {
		t.Fatal("mismatched projection admitted")
	}
}
