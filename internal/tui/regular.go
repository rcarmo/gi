package tui

import (
	"context"
	"database/sql"
	"errors"

	gotui "github.com/grindlemire/go-tui"
)

func (c *chatTUI) bindTranscriptNavigation() {
	if c.regularMode {
		c.input.onTranscriptTop, c.input.onTranscriptEnd = nil, nil
	} else {
		c.input.onTranscriptTop, c.input.onTranscriptEnd = c.scrollTranscriptToTop, c.scrollTranscriptToBottom
	}
}

// Regular mode never rewrites terminal-owned rows. Wait for native claim release
// before baking mutable draft/tool blocks into the terminal history.
func (c *chatTUI) regularBusy() bool {
	if c.running || c.compaction.active {
		return true
	}
	if c.store != nil && c.sessionID != "" {
		if _, _, err := c.store.GetSessionActiveTurn(context.Background(), c.sessionID); !errors.Is(err, sql.ErrNoRows) {
			return true // Unknown activity must not commit a mutable preview.
		}
	}
	return false
}

func (c *chatTUI) regularPending() []string {
	start := min(max(0, c.regularPrinted), len(c.transcript))
	return c.transcript[start:]
}

func (c *chatTUI) flushRegularTranscript() {
	if !c.regularMode || c.app == nil || c.workspaceIndex.active {
		return
	}
	if c.regularSessionPending {
		c.regularSessionPending = false
		c.app.PrintAboveln("sys: session %s", c.sessionID)
	}
	if c.regularBusy() {
		return
	}
	// Native completion and legacy final-response delivery are separate queued
	// events. Never bake a draft span before its final replacement arrives.
	end := c.regularStableEnd()
	if end <= c.regularPrinted {
		return
	}
	lines := c.transcript[c.regularPrinted:end]
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
	// Print complete retained output. Baked scrollback cannot later expand in place.
	for _, block := range c.buildTranscriptRenderableBlocks(lines) {
		if block.Kind == "thinking_indicator" {
			continue
		}
		block.Expanded = true
		root.AddChild(c.renderTranscriptBlock(block))
	}
	c.regularPrinted = end
	c.app.PrintAboveElement(root)
}

func (c *chatTUI) renderRegular(app *gotui.App) *gotui.Element {
	w, h := app.Size()
	if c.workspaceIndex.active && c.regularWidth != 0 && (w != c.regularWidth || h != c.regularHeight) {
		c.workspaceIndex.resized = true
	}
	if !c.workspaceIndex.active && c.regularWidth != 0 && (w != c.regularWidth || h != c.regularHeight) {
		// The inline renderer invalidates history geometry on width changes.
		// Re-establish it conservatively before any dynamic dock growth.
		app.PrintAboveln("sys: terminal resized to %dx%d", w, h)
	}
	c.regularWidth, c.regularHeight = w, h
	c.outputWidth, c.outputHeight = w, h
	if c.workspaceIndex.active {
		// Retain the previous inline height/layout while the alternate screen is
		// temporary, including multiline editor height. Only five rows contain UI.
		root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithHeight(5))
		root.AddChild(c.renderWorkspaceIndex(w))
		return root
	}
	c.ensureInput()
	c.input.width = w
	c.input.suspended = c.modelMenuOpen
	input := app.MountPersistent(c, 0, func() gotui.Component { return c.input })
	inputHeight := max(1, input.HeightForWidth(w))
	footer := c.footerLines(w)
	widgets := c.extensionWidgetLines()
	if c.editorAskActive {
		widgets = append(widgets, "? "+c.editorAskPrompt+" (Enter submit · Esc cancel)")
	}
	menuHeight := c.modelMenuHeight()
	// Active output is temporary and bounded; the idle dock has only editor,
	// separators and existing footer. Leave at least one terminal-owned history row.
	previewHeight := 0
	pending := c.regularPending()
	if len(pending) > 0 && c.regularBusy() {
		previewHeight = min(3, len(pending))
	}
	dock := min(h-1, 2+inputHeight+len(footer)+len(widgets)+menuHeight+previewHeight)
	app.SetInlineHeight(max(1, dock))
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithHeight(dock))
	c.transcriptBlockRefs = nil
	if previewHeight > 0 {
		preview := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithHeight(previewHeight), gotui.WithScrollable(gotui.ScrollVertical))
		for _, block := range c.buildTranscriptRenderableBlocks(pending) {
			preview.AddChild(c.renderTranscriptBlock(block))
		}
		preview.ScrollToBottom()
		root.AddChild(preview)
	}
	if c.modelMenuOpen {
		root.AddChild(c.renderModelMenu(w))
	}
	if len(widgets) > 0 {
		root.AddChild(c.renderLineBlock(widgets, gotui.NewStyle().Foreground(gotui.Blue)))
	}
	separator := func() *gotui.Element {
		return gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1), gotui.WithText(c.horizontalRule(w)), gotui.WithTextStyle(gotui.NewStyle().Dim()))
	}
	root.AddChild(separator())
	root.AddChild(input)
	root.AddChild(separator())
	root.AddChild(c.renderLineBlock(footer, gotui.NewStyle().Dim()))
	c.inputRegion = input
	return root
}

// Terminal-owned scrollback cannot be navigated or mutated by app shortcuts.
// Leave Home/End to the editor; the terminal/multiplexer owns history keys.
func regularKeyMap(bindings gotui.KeyMap) gotui.KeyMap {
	out := make(gotui.KeyMap, 0, len(bindings))
	for _, b := range bindings {
		switch b.Pattern.Key {
		case gotui.KeyPageUp, gotui.KeyPageDown, gotui.KeyHome, gotui.KeyEnd, gotui.KeyF6, gotui.KeyF7, gotui.KeyF8:
			continue
		}
		if b.Pattern.Rune == 'o' && b.Pattern.Mod == gotui.ModCtrl {
			continue
		}
		out = append(out, b)
	}
	return out
}

func (c *chatTUI) regularStableEnd() int {
	end := len(c.transcript)
	if c.draftLineIndex >= 0 && c.draftLineCount > 0 {
		end = min(end, c.draftLineIndex)
	}
	for i := c.regularPrinted; i < end; i++ {
		if meta, ok := parseTranscriptBlockMarker(c.transcript[i]); ok && meta.Status == "running" {
			end = i
			break
		}
	}
	return max(c.regularPrinted, end)
}
