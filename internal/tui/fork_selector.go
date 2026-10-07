package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/store"
)

// Pi's /fork (UserMessageSelectorComponent): pick an earlier user message,
// copy the history before it into a new session, switch to it and put the
// message text back in the editor. Messages are listed oldest first with the
// most recent selected; Up/Down wrap, Enter forks, Escape cancels.

const forkMaxVisible = 10

func (c *chatTUI) openForkSelector() {
	if c.store == nil || c.sessionID == "" {
		c.appendTranscript("sys: No messages to fork from")
		return
	}
	messages, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		c.appendTranscript(fmt.Sprintf("error: fork: %v", err))
		return
	}
	var texts []string
	ids := map[string]string{} // choice index -> message ID (texts may repeat)
	for _, m := range messages {
		if m.Role == "user" && strings.TrimSpace(m.Content) != "" {
			ids[strconv.Itoa(len(texts))] = m.ID
			texts = append(texts, m.Content)
		}
	}
	if len(texts) == 0 {
		c.appendTranscript("sys: No messages to fork from")
		return
	}
	c.modelMenuOpen = true
	c.modelMenuKind = "fork"
	c.modelMenuSession = c.selectionScope()
	c.modelMenuAll = texts
	c.modelMenuChoices = texts
	c.modelMenuValues = ids
	c.modelMenuQuery = ""
	c.modelMenuSelected = len(texts) - 1
	c.modelMenuScroll = 0
	c.modelMenuError = ""
	c.inputActive = false
	if c.app != nil {
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
}

func (c *chatTUI) forkSelectorKeys() gotui.KeyMap {
	move := func(delta int) {
		n := len(c.modelMenuChoices)
		if n == 0 {
			return
		}
		c.modelMenuSelected = ((c.modelMenuSelected+delta)%n + n) % n // wraps like Pi
		if c.app != nil {
			c.app.MarkDirty()
		}
	}
	return gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyCtrlC, func(gotui.KeyEvent) { c.closeModelMenu() }),
		gotui.OnPreemptStop(gotui.KeyEscape, func(gotui.KeyEvent) { c.closeModelMenu() }),
		gotui.OnPreemptStop(gotui.KeyUp, func(gotui.KeyEvent) { move(-1) }),
		gotui.OnPreemptStop(gotui.KeyDown, func(gotui.KeyEvent) { move(1) }),
		gotui.OnPreemptStop(gotui.KeyEnter, func(gotui.KeyEvent) { c.acceptForkSelection() }),
	}
}

func (c *chatTUI) piForkSelectorRows(width int) spanRows {
	rows := spanRows{{},
		{{Text: " " + truncateCells("Fork from Message", max(1, width-1)), Style: gotui.NewStyle().Bold()}},
		{{Text: " " + truncateCells("Select a user message to copy the active path up to that point into a new session", max(1, width-1)), Style: piFg(piMuted)}},
		{}, piRule(width, piBorder), {}}
	n := len(c.modelMenuChoices)
	start := max(0, min(c.modelMenuSelected-forkMaxVisible/2, n-forkMaxVisible))
	end := min(start+forkMaxVisible, n)
	for i := start; i < end; i++ {
		text := strings.TrimSpace(strings.ReplaceAll(c.modelMenuChoices[i], "\n", " "))
		text = truncateCells(text, max(1, width-2))
		if i == c.modelMenuSelected {
			rows = append(rows, []gotui.TextSpan{{Text: "› ", Style: piFg(piAccent)}, {Text: text, Style: gotui.NewStyle().Bold()}})
		} else {
			rows = append(rows, []gotui.TextSpan{{Text: "  " + text}})
		}
		rows = append(rows, []gotui.TextSpan{{Text: fmt.Sprintf("  Message %d of %d", i+1, n), Style: piFg(piMuted)}}, []gotui.TextSpan{})
	}
	if start > 0 || end < n {
		rows = append(rows, []gotui.TextSpan{{Text: fmt.Sprintf("  (%d/%d)", c.modelMenuSelected+1, n), Style: piFg(piMuted)}})
	}
	if c.modelMenuError != "" {
		rows = append(rows, []gotui.TextSpan{{Text: truncateCells("  "+c.modelMenuError, width), Style: piFg(piError)}})
	}
	return append(rows, spanRows{{}, piRule(width, piBorder)}...)
}

func (c *chatTUI) acceptForkSelection() {
	i := c.modelMenuSelected
	if i < 0 || i >= len(c.modelMenuChoices) || c.modelMenuValues[strconv.Itoa(i)] == "" {
		return
	}
	if !c.ownsScope(c.modelMenuSession) {
		c.modelMenuError = "session changed; reopen /fork"
		return
	}
	messageID, text := c.modelMenuValues[strconv.Itoa(i)], c.modelMenuChoices[i]
	forked, err := c.store.CloneSessionBefore(context.Background(), c.sessionID, store.NowID("session"), "", "", messageID)
	if err != nil {
		c.modelMenuError = "fork failed: " + err.Error()
		if c.app != nil {
			c.app.MarkDirty()
		}
		return
	}
	c.closeModelMenu()
	c.switchSession(forked.ID)
	c.ensureInput()
	c.input.SetText(text) // Pi puts the selected message back in the editor
	c.appendTranscript(fmt.Sprintf("sys: Forked to new session @%s (%s)", c.agentIDForSession(forked), forked.ID))
	if c.app != nil {
		c.app.MarkDirty()
	}
}
