package tui

import (
	"strings"

	gotui "github.com/grindlemire/go-tui"
)

// Pi's built-in dark theme (theme/dark.json), resolved to truecolor by Pi's
// own okhsl conversion. Keep these values in sync with Pi, not the ANSI
// palette, so gi and Pi render the same colors in the same terminal.
// The user message alone gets a background; transcript output stays on the
// terminal background.
var (
	piText            = piRGB(222, 224, 225)
	piMuted           = piRGB(157, 165, 169)
	piDim             = piRGB(126, 136, 142)
	piAccent          = piRGB(167, 152, 215)
	piError           = piRGB(234, 127, 129)
	piWarning         = piRGB(205, 154, 34)
	piSuccess         = piRGB(104, 183, 141)
	piThinkingText    = piRGB(150, 160, 164)
	piUserBg          = piRGB(33, 59, 73)
	piSelectedBg      = piRGB(33, 59, 73)
	piSearchMatchBg   = piRGB(78, 47, 27)
	piMdCode          = piAccent
	piBashMode        = piRGB(94, 178, 134)
	piBorderMuted     = piRGB(118, 129, 134)
	piThinkingOff     = piRGB(108, 118, 123)
	piThinkingMinimal = piRGB(104, 128, 141)
	piThinkingLow     = piRGB(84, 137, 164)
	piThinkingMedium  = piRGB(97, 133, 204)
	piThinkingHigh    = piRGB(151, 118, 229)
	piThinkingXhigh   = piRGB(222, 84, 193)
	piThinkingMax     = piRGB(254, 84, 98)
)

// Keep distinct Pi roles independent even when the built-ins share RGB values.
var (
	piMdHeading     = piWarning
	piMdCodeBlock   = piSuccess
	piMdCodeBorder  = piMuted
	piMdQuote       = piMuted
	piMdQuoteBorder = piMuted
	piMdHr          = piMuted
	piMdLinkUrl     = piMuted
	piMdListBullet  = piAccent
	piToolTitle     = piText
	piToolOutput    = piMuted
	piDiffAdded     = piSuccess
	piDiffRemoved   = piError
	piDiffContext   = piMuted
	piUserText      = piText
	piCustomLabel   = piAccent
	piCustomText    = piMuted
)

func piFg(c gotui.Color) gotui.Style { return gotui.NewStyle().Foreground(c) }

// Pi colors the editor border (and the embedded working status) by thinking
// level, or with bashMode while the draft is a `!` shell command.
func piThinkingBorderColor(level string) gotui.Color {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "minimal":
		return piThinkingMinimal
	case "low":
		return piThinkingLow
	case "medium":
		return piThinkingMedium
	case "high":
		return piThinkingHigh
	case "xhigh":
		return piThinkingXhigh
	case "max":
		return piThinkingMax
	}
	return piThinkingOff
}

func transcriptBand(kind, status string) (gotui.Color, bool) {
	switch kind {
	case "user":
		return piUserBg, true
	case "tool":
		return toolBandColor(status), true
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
	// Ctrl+O also expands the startup header, as Pi's app.tools.expand does.
	toggles := func(b transcriptRenderableBlock) bool {
		return (b.Kind == "tool" || b.Kind == "bash" || b.Kind == "local" || b.Kind == startupHeaderKind) && b.Expandable
	}
	for _, b := range blocks {
		if toggles(b) && !b.Expanded {
			expand = true
			break
		}
	}
	for _, b := range blocks {
		if toggles(b) {
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

// User messages and tool calls have padded backgrounds, as in Pi. Other
// transcript content has a separating blank row, but no boxed padding.
func transcriptSpacing(kind string) (separator, vertical, horizontal int) {
	switch kind {
	case "user":
		return 0, 1, 1
	case "assistant":
		return 1, 0, 1
	case "tool":
		// Pi's ToolExecutionComponent: Spacer(1) + Box(paddingX 1, paddingY 1).
		return 1, 1, 1
	case "bash":
		// BashExecutionComponent: Spacer(1) + full-width borders.
		return 1, 0, 0
	case "local", "error", "thought", "thinking", "hook", "route", "dispatcher", "subturn", "compact":
		return 1, 0, 1
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

// Pi inserts a Spacer(1) before every subsequent user message, not just
// after an assistant: adjacent colored bands must never touch. Other blocks
// usually provide their own leading separator. Keep gaps outside the band.
func assistantGapBefore(previousKind, nextKind string) int {
	if previousKind != "" && nextKind == "user" {
		return 1
	}
	separator, _, _ := transcriptSpacing(nextKind)
	if previousKind == "assistant" && separator == 0 {
		return 1
	}
	return 0
}

func (c *chatTUI) renderTranscriptBlockAfter(block transcriptRenderableBlock, previousKind string) *gotui.Element {
	if block.Kind == "thinking_indicator" {
		return c.renderTranscriptBlock(block)
	}
	content := c.renderTranscriptBlock(block)
	if assistantGapBefore(previousKind, block.Kind) == 0 {
		return content
	}
	wrapper := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
	wrapper.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1)))
	wrapper.AddChild(content)
	return wrapper
}

// Projection and layout must agree on the width inside the message band.
func (c *chatTUI) transcriptBlockContentWidth(kind string) int {
	_, _, horizontal := transcriptSpacing(kind)
	return max(1, c.currentContentWidth()-2*horizontal)
}
