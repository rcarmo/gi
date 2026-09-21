package tui

import (
	"strings"
	"unicode"

	gotui "github.com/grindlemire/go-tui"
)

// Selectors are single-line, terminal-cell bounded views of user-controlled
// titles. Never byte-slice UTF-8 or allow control sequences into the layout.
func selectorText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	text = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, text)
	if gotui.StringWidth(text) <= width {
		return text
	}
	var out strings.Builder
	used := 0
	for _, r := range text {
		n := gotui.StringWidth(string(r))
		if used+n > width-1 {
			break
		}
		out.WriteRune(r)
		used += n
	}
	return out.String() + "…"
}
