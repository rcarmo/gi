package tui

import (
	"context"
	"strings"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/inference"
)

type modelPickerMetadata struct{ search, unavailable string }

// Snapshot metadata while Alt-M is open or an unavailable row is retried.
// Enabled acceptance revalidates credentials/context through chooseSessionModel.
func (c *chatTUI) captureModelPickerMetadata() {
	c.modelMenuSession = c.selectionScope()
	c.modelMenuMetadata = map[string]modelPickerMetadata{}
	usage, err := c.store.LatestContextMeasurement(context.Background(), c.sessionID)
	for _, option := range c.sessionModelCatalogue() {
		search := []string{option.Label, option.Provider, option.ID, option.Name}
		if option.ContextWindow > 0 {
			search = append(search, formatTokenCount(option.ContextWindow)+" ctx")
		}
		if option.Reasoning {
			search = append(search, "reasoning")
		}
		reason := ""
		switch {
		case err != nil:
			reason = "context unavailable; Enter to retry"
		case !inference.UsableSessionModel(option):
			reason = "unavailable or lacks credentials"
		case usage != nil && option.ContextWindow > 0 && usage.Tokens > option.ContextWindow:
			reason = "context too small"
		}
		c.modelMenuMetadata[option.Label] = modelPickerMetadata{strings.Join(search, " "), reason}
	}
}

func (c *chatTUI) modelPickerUnavailable(label string) string {
	if c.modelMenuKind != "model" {
		return ""
	}
	metadata, ok := c.modelMenuMetadata[label]
	if !ok {
		return "unavailable or lacks credentials"
	}
	return metadata.unavailable
}

func (c *chatTUI) enabledModelMenuIndices() []int {
	indices := make([]int, 0, len(c.modelMenuChoices))
	for i, label := range c.modelMenuChoices {
		if c.modelPickerUnavailable(label) == "" {
			indices = append(indices, i)
		}
	}
	return indices
}

// Keep go-tui's inline renderer while using a temporary screen; switching the
// renderer itself invokes a full clear that can erase terminal-owned history.
func (c *chatTUI) openModelPickerScreen() {
	if !c.regularMode || c.app == nil || c.modelMenuAltScreen {
		return
	}
	c.modelMenuAltScreen, c.modelMenuResized = true, false
	c.modelMenuInlineHeight = c.app.InlineHeight()
	c.app.Terminal().EnterAltScreen()
	w, h := c.app.Size()
	c.app.Dispatch(gotui.ResizeEvent{Width: w, Height: h})
}

func (c *chatTUI) closeModelPickerScreen() {
	if !c.modelMenuAltScreen || c.app == nil {
		return
	}
	c.modelMenuAltScreen = false
	// Shrink on the temporary screen before restoring main-screen history.
	c.app.SetInlineHeight(c.modelMenuInlineHeight)
	c.app.Terminal().ExitAltScreen()
	w, h := c.app.Size()
	c.app.Dispatch(gotui.ResizeEvent{Width: w, Height: h})
	if c.modelMenuResized {
		c.app.PrintAboveln("sys: terminal resized to %dx%d", w, h)
	}
	c.modelMenuResized = false
}

func (c *chatTUI) resetModelMenuMetadata() {
	c.modelMenuMetadata = nil
	c.modelMenuSession = sessionScope{}
}
