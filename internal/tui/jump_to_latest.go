package tui

import (
	gotui "github.com/grindlemire/go-tui"
)

// Pi's fullscreen "Jump to latest message" indicator (pi-tui TuiAltScreen
// scrollToEndIndicator, as configured by pi-coding-agent's tui-renderer):
// while the transcript is scrolled away from its end, a label is drawn
// centred on the transcript's last row, in the text colour on selectedBg.
// Clicking it, or pressing Ctrl+End (tui.altScreen.bottom), scrolls to the
// bottom and resumes following new output.

// jumpToLatestLabel is Pi's label with the default tui.altScreen.bottom key.
const jumpToLatestLabel = " ↓ Jump to latest message · Ctrl+End "

type jumpToLatestRect struct{ row, column, width int }

// compositeJumpToLatest draws the indicator over the laid-out frame (go-tui
// pre-flush hook) and records where it is for mouse clicks.
func (c *chatTUI) compositeJumpToLatest(buf *gotui.Buffer) {
	c.jumpToLatest = jumpToLatestRect{}
	if c.regularMode || c.stickToBottom || c.transcriptRegion == nil || c.transcriptRegion.IsAtBottom() {
		return
	}
	clip := c.transcriptRegion.Rect()
	if clip.Width <= 0 || clip.Height <= 0 {
		return
	}
	row := clip.Y + clip.Height - 1
	if row < 0 || row >= buf.Height() {
		return
	}
	label := truncateToDisplayWidth(jumpToLatestLabel, clip.Width)
	width := gotui.StringWidth(label)
	if width == 0 {
		return
	}
	column := clip.X + (clip.Width-width)/2
	buf.SetString(column, row, label, gotui.NewStyle().Foreground(piText).Background(piSelectedBg))
	c.jumpToLatest = jumpToLatestRect{row: row, column: column, width: width}
}

// handleJumpToLatestClick scrolls to the bottom when the indicator is
// clicked (a left press, as Pi).
func (c *chatTUI) handleJumpToLatestClick(me gotui.MouseEvent) bool {
	r := c.jumpToLatest
	if r.width == 0 || me.Button != gotui.MouseLeft || me.Action != gotui.MousePress {
		return false
	}
	if me.Y != r.row || me.X < r.column || me.X >= r.column+r.width {
		return false
	}
	c.scrollTranscriptToBottom()
	return true
}

// truncateToDisplayWidth cuts s to at most width terminal columns, without
// an ellipsis (Pi's truncateToWidth(label, width, "")).
func truncateToDisplayWidth(s string, width int) string {
	if gotui.StringWidth(s) <= width {
		return s
	}
	out := []rune{}
	used := 0
	for _, r := range s {
		w := gotui.StringWidth(string(r))
		if used+w > width {
			break
		}
		out = append(out, r)
		used += w
	}
	return string(out)
}
