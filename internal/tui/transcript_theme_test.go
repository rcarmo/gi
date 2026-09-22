package tui

import (
	"fmt"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
)

func TestPiTranscriptOutcomeBands(t *testing.T) {
	for _, size := range [][2]int{{60, 18}, {100, 22}, {140, 36}} {
		for _, tc := range []struct {
			kind, status string
			bg           gotui.Color
		}{
			{"user", "", piUserBg}, {"tool", "running", piToolPendingBg},
			{"tool", "ok", piToolSuccessBg}, {"tool", "error", piToolErrorBg},
			{"bash", "failed", piToolErrorBg}, {"bash", "cancelled", piToolErrorBg},
		} {
			t.Run(fmt.Sprintf("%dx%d/%s/%s", size[0], size[1], tc.kind, tc.status), func(t *testing.T) {
				c := &chatTUI{}
				head, body, hint, _, _ := transcriptBlockPalette(tc.kind, tc.status, false)
				block := transcriptRenderableBlock{Kind: tc.kind, Status: tc.status, Header: "Title", HeaderStyle: head, BodyStyle: body, HintStyle: hint}
				if tc.kind != "user" {
					block.Body = []string{"output line"}
				}
				el := c.renderTranscriptBlock(block)
				buf := gotui.NewBuffer(size[0], size[1])
				root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(size[0]), gotui.WithHeight(size[1]))
				root.AddChild(el)
				root.Render(buf, size[0], size[1])
				wantHeight, separator := 5, 1 // external blank + top/bottom pad + title/body
				if tc.kind == "user" {
					wantHeight, separator = 3, 0
				}
				if el.Rect().Height != wantHeight {
					t.Fatalf("message padding height %d != %d", el.Rect().Height, wantHeight)
				}
				for y := 0; y < el.Rect().Height; y++ {
					for _, x := range []int{0, size[0] / 2, size[0] - 1} {
						want := tc.bg
						if y < separator {
							want = gotui.Color{}
						}
						if got := buf.Cell(x, y).Style.Bg; got != want {
							t.Fatalf("band/separator mismatch at %d,%d: %v", x, y, got)
						}
					}
					if y <= separator || y == wantHeight-1 {
						for x := 0; x < size[0]; x++ {
							if cell := buf.Cell(x, y); cell.Rune != ' ' && cell.Rune != 0 {
								t.Fatalf("padding contains border/text at %d,%d: %c", x, y, cell.Rune)
							}
						}
					}
				}
			})
		}
	}
	if _, ok := transcriptBand("assistant", ""); ok {
		t.Fatal("assistant should use terminal background")
	}
}

func TestPiTranscriptRenderedScrollBounds(t *testing.T) {
	for _, width := range []int{60, 100, 140} {
		c := &chatTUI{transcript: []string{"user: " + strings.Repeat("wrapped output ", 100)}, transcriptRef: gotui.NewRef()}
		root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(10), gotui.WithScrollable(gotui.ScrollVertical))
		root.AddChild(c.renderInlineStyledLine(c.transcript[0], gotui.NewStyle()))
		root.Render(gotui.NewBuffer(width, 10), width, 10)
		c.transcriptRef.Set(root)
		if c.transcriptMaxScroll() <= 0 {
			t.Fatal("wrapped text cannot scroll", width)
		}
		c.scrollTranscriptToBottom()
		_, maxY := root.MaxScroll()
		if c.transcriptScroll != maxY || !c.stickToBottom {
			t.Fatal("bottom uses source line count")
		}
		root.ScrollTo(0, maxY)
		c.pageTranscript(-1)
		if c.transcriptScroll != max(0, maxY-9) || c.stickToBottom {
			t.Fatal("page lost rendered offset", c.transcriptScroll, maxY)
		}
		c.scrollTranscriptToTop()
		if c.transcriptScroll != 0 || c.stickToBottom {
			t.Fatal("top following")
		}
	}
}

func TestPiToolOutputToggleRetainsEditor(t *testing.T) {
	c := &chatTUI{}
	c.ensureInput()
	c.input.SetText("newer draft")
	c.input.cursorPos = 3
	c.appendTranscriptBlock(transcriptBlockMeta{Key: "tool", Kind: "tool", Title: "read", Status: "ok"}, []string{"one", "two", "three", "four"})
	c.toggleToolOutput()
	if !c.transcriptExpanded["tool"] {
		t.Fatal("tool not expanded")
	}
	c.toggleToolOutput()
	if c.transcriptExpanded["tool"] {
		t.Fatal("tool not collapsed")
	}
	if c.input.Text() != "newer draft" || c.input.cursorPos != 3 {
		t.Fatal("expand changed editor")
	}
}

func TestPiFullscreenWheelOverEditorScrollsTranscript(t *testing.T) {
	c := &chatTUI{transcriptRef: gotui.NewRef()}
	c.ensureInput()
	c.input.SetText("draft")
	c.input.cursorPos = 2
	el := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(60), gotui.WithHeight(10), gotui.WithScrollable(gotui.ScrollVertical))
	for i := 0; i < 30; i++ {
		el.AddChild(gotui.New(gotui.WithText(fmt.Sprint(i)), gotui.WithHeight(1)))
	}
	el.Render(gotui.NewBuffer(60, 18), 60, 10)
	el.ScrollTo(0, 20)
	c.transcriptRef.Set(el)
	c.transcriptRegion = el
	c.stickToBottom = true
	if !c.handleTranscriptScrollEvent(gotui.MouseEvent{Button: gotui.MouseWheelUp, X: 5, Y: 15}) || c.transcriptScroll != 17 || c.stickToBottom {
		t.Fatal("editor wheel failed to scroll history", c.transcriptScroll)
	}
	if c.input.Text() != "draft" || c.input.cursorPos != 2 {
		t.Fatal("wheel altered editor")
	}
	c.modelMenuOpen = true
	if c.handleTranscriptScrollEvent(gotui.MouseEvent{Button: gotui.MouseWheelUp, X: 5, Y: 15}) {
		t.Fatal("wheel stole selector input")
	}
}

func TestPiMessageSpacingGroupsMarkdownContinuationRows(t *testing.T) {
	for _, width := range []int{60, 100, 140} {
		c := &chatTUI{}
		c.cfg.AssistantName = "Gi"
		lines := renderMarkdownTranscript("you: ", "First paragraph\n\n- list one\n- list two\n\n```go\nfmt.Println(\"hello\")\n```", width-2)
		lines = append(lines, renderMarkdownTranscript("Gi: ", "Answer paragraph\n\nSecond paragraph\n\n- detail", width-2)...)
		lines = append(lines, "you: another message", "sys: independent notice")
		blocks := c.buildTranscriptRenderableBlocks(lines)
		if len(blocks) != 4 || blocks[0].Kind != "user" || blocks[1].Kind != "assistant" || len(blocks[0].Body) < 4 || len(blocks[1].Body) < 3 {
			t.Fatalf("%d: continuation rows escaped parent message: %+v", width, blocks)
		}
		for _, block := range blocks[:2] {
			el := c.renderTranscriptBlock(block)
			height := el.HeightForWidth(width)
			root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(height))
			root.AddChild(el)
			buf := gotui.NewBuffer(width, height)
			root.Render(buf, width, height)
			text := buf.StringTrimmed()
			if block.Kind == "user" && !strings.Contains(text, "list two") {
				t.Fatal("lost user body", text)
			}
			if block.Kind == "assistant" && !strings.Contains(text, "Second paragraph") {
				t.Fatal("lost assistant body", text)
			}
			if block.Expandable {
				t.Fatal("ordinary multiline message collapsed")
			}
			first := strings.TrimSpace(strings.Split(text, "\n")[0])
			if first != "" {
				t.Fatalf("missing message top blank: %q", first)
			}
		}
	}
}
