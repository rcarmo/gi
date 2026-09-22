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

// Match Pi's message-level spacing: user Box(outputPad,1); assistant
// Spacer(1)+Markdown(outputPad,0); default tools Spacer(1)+Box(1,1).
// The external spacer is deliberately not part of the outcome background.
func transcriptSpacing(kind string) (separator, vertical, horizontal int) {
	switch kind {
	case "user":
		return 0, 1, 1
	case "assistant":
		return 1, 0, 1
	case "tool", "bash", "local", "error":
		return 1, 1, 1
	default:
		return 0, 0, 0
	}
}

func padTranscriptBlock(content *gotui.Element, block transcriptRenderableBlock) *gotui.Element {
	separator, vertical, horizontal := transcriptSpacing(block.Kind)
	if separator == 0 && vertical == 0 && horizontal == 0 {
		return content
	}
	band := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithPaddingTRBL(vertical, horizontal, vertical, horizontal))
	band.AddChild(content)
	applyTranscriptBand(band, block)
	if separator == 0 {
		return band
	}
	wrapper := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
	wrapper.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(separator)))
	wrapper.AddChild(band)
	return wrapper
}
