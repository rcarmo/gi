package tui

import gotui "github.com/grindlemire/go-tui"

// Pi 1.5's getSelectPageIndex: page by the list's visible item count,
// clamp at either end (unlike arrow keys), and never select an empty list.
func selectPageIndex(selected, count, pageSize, direction int) int {
	return max(0, min(count-1, selected+direction*max(1, pageSize)))
}

func selectorPageKeys(page func(direction int)) gotui.KeyMap {
	return gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyPageUp, func(gotui.KeyEvent) { page(-1) }),
		gotui.OnPreemptStop(gotui.KeyPageDown, func(gotui.KeyEvent) { page(1) }),
	}
}

func (c *chatTUI) pageModelMenuSelection(direction int) {
	size := c.modelMenuVisibleRows()
	if c.modelMenuKind == "model" {
		size = c.piModelSelectorListRows()
	}
	if len(c.modelMenuChoices) == 0 {
		return
	}
	index := selectPageIndex(c.modelMenuSelected, len(c.modelMenuChoices), size, direction)
	// Keep Gi's unavailable-model admission guard. Paging is in visible rows,
	// not in the number of enabled models skipped by arrow navigation.
	if c.modelMenuKind == "model" {
		for index >= 0 && index < len(c.modelMenuChoices) && c.modelPickerUnavailable(c.modelMenuChoices[index]) != "" {
			index += direction
		}
		if index < 0 || index >= len(c.modelMenuChoices) {
			return
		}
	}
	c.setModelMenuSelection(index)
}

func (c *chatTUI) pageSlashSelection(direction int) {
	if len(c.slash.items) > 0 {
		c.slash.selected = selectPageIndex(c.slash.selected, len(c.slash.items), slashMaxVisible, direction)
		c.markDirty()
	}
}
