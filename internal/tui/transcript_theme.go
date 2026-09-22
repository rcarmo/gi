package tui

import gotui "github.com/grindlemire/go-tui"

// Pi 0.85.1 built-in dark theme. Bands describe content/outcomes, not alternating
// row numbers. Keep these values independent of the terminal's ANSI palette.
var (
	piUserBg        = gotui.RGBColor(0x34, 0x35, 0x41)
	piToolPendingBg = gotui.RGBColor(0x28, 0x28, 0x32)
	piToolSuccessBg = gotui.RGBColor(0x28, 0x32, 0x28)
	piToolErrorBg   = gotui.RGBColor(0x3c, 0x28, 0x28)
	piText          = gotui.RGBColor(0xd4, 0xd4, 0xd4)
	piError         = gotui.RGBColor(0xf4, 0x87, 0x71)
)

func transcriptBand(kind, status string) (gotui.Color, bool) {
	switch kind {
	case "user":
		return piUserBg, true
	case "error":
		return piToolErrorBg, true
	case "tool", "bash", "local":
		switch status {
		case "ok", "completed", "success":
			return piToolSuccessBg, true
		case "error", "failed", "cancelled", "aborted":
			return piToolErrorBg, true
		default:
			return piToolPendingBg, true
		}
	}
	return gotui.Color{}, false
}

func applyTranscriptBand(element *gotui.Element, block transcriptRenderableBlock) {
	if bg, ok := transcriptBand(block.Kind, block.Status); ok {
		style := gotui.NewStyle().Background(bg)
		element.SetBackground(&style)
	}
}

// The fullscreen transcript navigates rendered rows, including wrapped and
// expanded output. The fallback is only for tests/before the first layout.
func (c *chatTUI) transcriptMaxScroll() int {
	if c.transcriptRef != nil && c.transcriptRef.El() != nil {
		_, maxY := c.transcriptRef.El().MaxScroll()
		return maxY
	}
	return max(0, len(c.visibleTranscript())-c.transcriptViewportHeight())
}

func (c *chatTUI) toggleToolOutput() {
	blocks := c.buildTranscriptRenderableBlocks(c.visibleTranscript())
	expand := false
	for _, b := range blocks {
		if (b.Kind == "tool" || b.Kind == "bash" || b.Kind == "local") && b.Expandable && !b.Expanded {
			expand = true
			break
		}
	}
	for _, b := range blocks {
		if (b.Kind == "tool" || b.Kind == "bash" || b.Kind == "local") && b.Expandable {
			c.transcriptExpanded[b.Key] = expand
		}
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) setTranscriptPosition(row int) {
	c.transcriptScroll = max(0, row)
	if c.transcriptRef != nil && c.transcriptRef.El() != nil {
		c.transcriptRef.El().ScrollTo(0, c.transcriptScroll)
	}
}
