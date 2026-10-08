package tui

import (
	"strings"
	"testing"
	"time"

	gotui "github.com/grindlemire/go-tui"
)

func renderToolForTest(t *testing.T, c *chatTUI, meta transcriptBlockMeta, body []string, width int) (*gotui.Buffer, string, int) {
	t.Helper()
	lines := []string{encodeTranscriptBlockMarker(meta)}
	for _, l := range body {
		lines = append(lines, "│ "+l)
	}
	blocks := c.buildTranscriptRenderableBlocks(lines)
	el := c.renderTranscriptBlock(blocks[0])
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(width), gotui.WithHeight(30))
	root.AddChild(el)
	buf := gotui.NewBuffer(width, 30)
	root.RenderTo(buf, width, 30)
	return buf, buf.StringTrimmed(), el.Rect().Height
}

func TestPiToolBandFollowsStatus(t *testing.T) {
	c := &chatTUI{transcriptExpanded: map[string]bool{}}
	start := time.Now().Add(-1500 * time.Millisecond).UTC().Format(time.RFC3339Nano)
	for status, want := range map[string]gotui.Color{"running": piToolPendingBg, "ok": piToolSuccessBg, "error": piToolErrorBg} {
		meta := transcriptBlockMeta{Key: "t", Kind: "tool", Title: "bash", Detail: "ls -la", Status: status, StartedAt: start}
		if status != "running" {
			meta.EndedAt = time.Now().UTC().Format(time.RFC3339Nano)
			ms := int64(1500)
			meta.DurationMS = &ms
		}
		buf, screen, height := renderToolForTest(t, c, meta, []string{"a", "b"}, 40)
		// Row 0 is Pi's Spacer; the band (padding included) fills every other cell.
		for y := 1; y < height; y++ {
			for x := 0; x < 40; x++ {
				if bg := buf.Cell(x, y).Style.Bg; bg != want {
					t.Fatalf("%s: bg at %d,%d = %v want %v\n%s", status, x, y, bg, want, screen)
				}
			}
		}
		label := "Took"
		if status == "running" {
			label = "Elapsed"
		}
		if !strings.Contains(screen, " $ ls -la") || !strings.Contains(screen, label+" 1.") {
			t.Fatalf("%s: header/timing differ from Pi: %q", status, screen)
		}
	}
}

func TestPiToolPreviewHints(t *testing.T) {
	c := &chatTUI{transcriptExpanded: map[string]bool{}}
	body := make([]string, 12)
	for i := range body {
		body[i] = string(rune('a' + i))
	}
	_, screen, _ := renderToolForTest(t, c, transcriptBlockMeta{Key: "r", Kind: "tool", Title: "read", Detail: "main.go", Status: "ok"}, body, 60)
	if !strings.Contains(screen, "read main.go") || !strings.Contains(screen, "... (2 more lines, ctrl+o to expand)") || strings.Contains(screen, "\n l\n") {
		t.Fatalf("read preview: %q", screen)
	}
	_, screen, _ = renderToolForTest(t, c, transcriptBlockMeta{Key: "s", Kind: "tool", Title: "shell", Detail: "make", Status: "ok"}, body, 60)
	if !strings.Contains(screen, "... (7 earlier lines, ctrl+o to expand)") || strings.Contains(screen, " a\n") || !strings.Contains(screen, " l") {
		t.Fatalf("shell preview: %q", screen)
	}
}

func TestPiBashExecutionBlock(t *testing.T) {
	c := &chatTUI{transcriptExpanded: map[string]bool{}, outputWidth: 40}
	lines := c.bashBlockLines("false", "out", "error", &exitErr{}, time.Now(), time.Now())
	blocks := c.buildTranscriptRenderableBlocks(lines)
	el := c.renderTranscriptBlock(blocks[0])
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidth(40), gotui.WithHeight(12))
	root.AddChild(el)
	buf := gotui.NewBuffer(40, 12)
	root.RenderTo(buf, 40, 12)
	screen := buf.StringTrimmed()
	for _, want := range []string{"────", " $ false", " out", " (exit 1)"} {
		if !strings.Contains(screen, want) {
			t.Fatalf("missing %q: %q", want, screen)
		}
	}
}

type exitErr struct{}

func (*exitErr) Error() string { return "exit status 1" }
