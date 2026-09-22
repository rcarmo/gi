package tui

import (
	"fmt"
	gotui "github.com/grindlemire/go-tui"
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
			if buf.Cell(0, c.search.matches[0]).Style.Bg != piText {
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
		if c.transcriptScroll != 0 {
			t.Fatal("prompt position not synchronized")
		}
	}
}
