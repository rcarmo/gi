package tui

import (
	"fmt"
	"testing"

	gotui "github.com/grindlemire/go-tui"
)

// Colored message bands have one full-width padding row above/below, and
// exactly one uncolored separator between bands. Verify pixels, not just text,
// across status changes, expansion, wrapping and both transcript render paths.
func TestColoredBlockSpacingConsistent(t *testing.T) {
	for _, width := range []int{60, 100, 140} {
		for _, status := range []string{"running", "ok", "error", "failed"} {
			for _, expanded := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%s/expanded=%v", width, status, expanded), func(t *testing.T) {
					c := &chatTUI{outputWidth: width, outputHeight: 36, transcriptExpanded: map[string]bool{"tool": expanded}}
					c.cfg.AssistantName = "Gi"
					c.appendUserPrompt("First colored message", false)
					c.appendUserPrompt("Second colored message", false)
					c.appendTranscriptBlock(transcriptBlockMeta{Key: "tool", Kind: "tool", Title: "shell", Status: status}, []string{"tool result"})
					c.appendUserPrompt("After the tool", false)
					rows := c.transcriptRowsAtWidth(width)
					c.regularPrinted = len(c.transcript)
					buf, height := c.regularTranscriptRows(width)
					if buf == nil || height != len(rows) {
						t.Fatalf("regular/fullscreen height differs: %d / %d", height, len(rows))
					}
					var bands []struct {
						start, end int
						bg         gotui.Color
					}
					for y, row := range rows {
						bg := row.spans[0].Style.Bg
						for x := 0; x < width; x++ {
							if row.spans[x].Style.Bg != bg {
								t.Fatalf("nonuniform background at %d,%d", x, y)
							}
							if buf.Cell(x, y).Style.Bg != bg {
								t.Fatalf("regular/fullscreen background differs at %d,%d", x, y)
							}
						}
						if bg == (gotui.Color{}) {
							continue
						}
						if len(bands) == 0 || bands[len(bands)-1].end != y-1 || bands[len(bands)-1].bg != bg {
							bands = append(bands, struct {
								start, end int
								bg         gotui.Color
							}{y, y, bg})
						} else {
							bands[len(bands)-1].end = y
						}
					}
					if len(bands) != 4 {
						t.Fatalf("colored bands merged or disappeared: %+v", bands)
					}
					for i, band := range bands {
						if band.end-band.start < 2 {
							t.Fatal("missing colored padding", band)
						}
						if rows[band.start].text != "" || rows[band.end].text != "" {
							t.Fatalf("padding row contains text: %+v", band)
						}
						if i > 0 && band.start-bands[i-1].end != 2 {
							t.Fatalf("expected one uncolored separator: %+v", bands)
						}
					}
				})
			}
		}
	}
}
