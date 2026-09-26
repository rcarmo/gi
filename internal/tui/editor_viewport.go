package tui

// Pi's editor uses at most 30% of terminal rows with a five-line minimum.
// Gi also reserves its existing dock, menu and transcript/history area; tiny
// terminals may therefore use fewer than five lines. No text is discarded.
func editorViewportRows(terminalRows, available int) int {
	return max(1, min(max(5, terminalRows*3/10), available))
}

func (c *chatTUI) boundEditor(terminalRows, padding, footerRows, widgetRows, menuRows int, regular bool) {
	reserve := 4
	if regular {
		reserve = 1
	}
	available := terminalRows - 2*padding - footerRows - widgetRows - 2 - menuRows - reserve
	c.input.maxLines = editorViewportRows(terminalRows, available)
}
