package tui

import (
	"context"
	"strings"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/inference"
)

type modelPickerMetadata struct{ search, context, unavailable string }

// Snapshot metadata while Alt-M is open or an unavailable row is retried.
// Enabled acceptance revalidates credentials/context through chooseSessionModel.
func (c *chatTUI) captureModelPickerMetadata() {
	c.modelMenuSession = c.selectionScope()
	c.modelMenuMetadata = map[string]modelPickerMetadata{}
	usage, err := c.store.LatestContextMeasurement(context.Background(), c.sessionID)
	for _, option := range c.sessionModelCatalogue() {
		search := []string{option.Label, option.Provider, option.ID, option.Name}
		contextLabel := ""
		if option.ContextWindow > 0 {
			contextLabel = formatTokenCount(option.ContextWindow) + " ctx"
			search = append(search, contextLabel)
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
		c.modelMenuMetadata[option.Label] = modelPickerMetadata{search: strings.Join(search, " "), context: contextLabel, unavailable: reason}
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

// Context is advisory metadata, shown only when the complete identity and
// suffix fit. Never shorten a model key merely to make room for a badge.
func (c *chatTUI) modelPickerRowLabel(label string, width int) string {
	if c.modelMenuKind == "model" {
		metadata := c.modelMenuMetadata[label]
		if metadata.unavailable == "" && metadata.context != "" {
			full := label + " · " + metadata.context
			if gotui.StringWidth(full) <= width {
				return selectorText(full, width)
			}
		}
	}
	return selectorText(label, width)
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

// Model and session selectors share this bounded temporary-screen lifecycle.
// Keep go-tui's inline renderer: switching renderers invokes a full clear that
// can erase terminal-owned history.
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
	c.modelMenuRenderedHeight = 0
}

func (c *chatTUI) resetModelMenuMetadata() {
	c.modelMenuMetadata = nil
	c.modelMenuSession = sessionScope{}
}
