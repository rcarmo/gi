package tui

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/rcarmo/gi/internal/topics"
)

// Ownership is captured at subscription/submission time, not on delivery. The
// generation distinguishes A -> B -> A from the original visit to A.
type sessionScope struct {
	id         string
	generation uint64
}

type sessionEvent struct {
	scope   sessionScope
	payload map[string]any
}

type sessionTopicEvent struct {
	scope    sessionScope
	envelope topics.Envelope
}

type sessionEditorState struct {
	text               string
	cursor             int
	undoText           string
	undoCursor         int
	hasUndo            bool
	yank               string
	history            []string
	histIdx            int
	historySearchQuery string
	historySearchIdx   int
	queuedDrafts       []string
}

func (c *chatTUI) selectionScope() sessionScope {
	return sessionScope{c.sessionID, c.sessionGeneration}
}

func (c *chatTUI) ownsScope(scope sessionScope) bool {
	return scope == c.selectionScope()
}

func (c *chatTUI) handleSessionEvent(event sessionEvent) {
	if c.ownsScope(event.scope) {
		c.handleEvent(event.payload)
	}
}

func (c *chatTUI) handleSessionTopicEvent(event sessionTopicEvent) {
	if c.ownsScope(event.scope) {
		c.handleTopicEvent(event.envelope)
	}
}

func (c *chatTUI) stopSessionSubscription() {
	if c.subscriptionCancel != nil {
		c.subscriptionCancel()
		c.subscriptionCancel = nil
	}
	if c.subscribedCh != nil && c.engine != nil {
		c.engine.Unsubscribe(c.sessionID, c.subscribedCh)
		c.subscribedCh = nil
	}
	if c.topicUnsubscribe != nil {
		c.topicUnsubscribe()
		c.topicUnsubscribe = nil
	}
}

func forwardSessionEvents(ctx context.Context, source <-chan map[string]any, target chan<- sessionEvent, scope sessionScope) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-source:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return
			case target <- sessionEvent{scope, event}:
			}
		}
	}
}

func forwardSessionTopics(ctx context.Context, source <-chan topics.Envelope, target chan<- sessionTopicEvent, scope sessionScope) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-source:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return
			case target <- sessionTopicEvent{scope, event}:
			}
		}
	}
}

func (c *chatTUI) saveSessionEditor() {
	if c.sessionID == "" {
		return
	}
	c.ensureInput()
	if c.sessionEditors == nil {
		c.sessionEditors = map[string]sessionEditorState{}
	}
	c.sessionEditors[c.sessionID] = sessionEditorState{
		text: c.input.Text(), cursor: c.input.cursorPos,
		undoText: c.input.undoText, undoCursor: c.input.undoCursor, hasUndo: c.input.hasUndo, yank: c.input.yankText,
		history: append([]string(nil), c.history...), histIdx: c.histIdx,
		historySearchQuery: c.historySearchQuery, historySearchIdx: c.historySearchIdx,
		queuedDrafts: append([]string(nil), c.queuedDrafts...),
	}
}

func (c *chatTUI) restoreSessionEditor() {
	c.ensureInput()
	state, ok := c.sessionEditors[c.sessionID]
	if !ok {
		state = sessionEditorState{history: c.loadCommandHistory(), histIdx: -1, historySearchIdx: -1}
	}
	c.input.SetText(state.text)
	c.input.cursorPos = min(max(0, state.cursor), utf8.RuneCountInString(c.input.Text()))
	c.input.undoText, c.input.undoCursor, c.input.hasUndo = state.undoText, state.undoCursor, state.hasUndo
	c.input.yankText = state.yank
	c.history = append([]string(nil), state.history...)
	c.histIdx, c.historySearchIdx, c.historySearchQuery = state.histIdx, state.historySearchIdx, state.historySearchQuery
	c.queuedDrafts = append([]string(nil), state.queuedDrafts...)
}

// Guard UI completions at application time too: QueueUpdate may run after a
// switch, and a successful routed submission must not steal the new selection.
func (c *chatTUI) applySessionCompletion(scope sessionScope, apply func()) {
	guarded := func() {
		if !c.ownsScope(scope) {
			return
		}
		apply()
		if c.app != nil {
			c.app.MarkDirty()
		}
	}
	if c.app != nil {
		c.app.QueueUpdate(guarded)
	} else {
		guarded()
	}
}

func sessionEventMatchesID(ev map[string]any, id string) bool {
	if eventID, _ := ev["session_id"].(string); eventID != "" && eventID != id {
		return false
	}
	if jid, _ := ev["chat_jid"].(string); strings.HasPrefix(jid, "gi:") && strings.TrimPrefix(jid, "gi:") != id {
		return false
	}
	return true
}
