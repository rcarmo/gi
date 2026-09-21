package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func collectElementTexts(element *gotui.Element) []string {
	texts := []string{}
	if element.Text() != "" {
		texts = append(texts, element.Text())
	}
	for _, child := range element.Children() {
		texts = append(texts, collectElementTexts(child)...)
	}
	return texts
}

func sessionTestChat(t *testing.T) *chatTUI {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "gi.db"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(context.Background(), id, "@"+id, map[string]any{"status": "idle", "model": "bootstrap"}); err != nil {
			t.Fatal(err)
		}
	}
	e := turn.New(s)
	c := &chatTUI{store: s, engine: e, cfg: config.RuntimeConfig{DefaultModel: "bootstrap"}, draftLineIndex: -1}
	c.ensureInput()
	c.bindSession("A")
	t.Cleanup(func() { c.stopSessionSubscription(); e.Close(); s.Close() })
	return c
}

func TestSessionEditorRoundTripAndFailedSwitch(t *testing.T) {
	c := sessionTestChat(t)
	c.input.SetText("A unsent\n中文🙂 /attach photo.png")
	c.input.cursorPos = 4
	c.input.undoText, c.input.undoCursor, c.input.hasUndo, c.input.yankText = "previous A", 3, true, "cut A"
	c.history, c.histIdx, c.historySearchQuery, c.historySearchIdx = []string{"history A"}, 0, "A", 0
	c.queuedDrafts = []string{"queued A"}
	c.draft, c.draftLineCount = "streamed A", 2
	c.extensionWidgets = map[string][]string{"origin": {"A widget"}}
	c.lastContextTokens = 100
	if !c.switchSession("B") {
		t.Fatal("switch B")
	}
	if c.input.Text() != "" || c.draft != "" || c.draftLineCount != 0 || c.lastContextTokens != 0 || len(c.extensionWidgets) != 0 || len(c.queuedDrafts) != 0 || c.input.hasUndo || c.input.yankText != "" {
		t.Fatalf("A leaked to B: %+v", c)
	}
	c.input.SetText("B unsent")
	c.input.cursorPos = 2
	c.history = []string{"history B"}
	c.switchSession("A")
	if c.input.Text() != "A unsent\n中文🙂 /attach photo.png" || c.input.cursorPos != 4 || c.input.undoText != "previous A" || c.input.yankText != "cut A" || !reflect.DeepEqual(c.history, []string{"history A"}) || c.histIdx != 0 || c.historySearchQuery != "A" || c.historySearchIdx != 0 || !reflect.DeepEqual(c.queuedDrafts, []string{"queued A"}) {
		t.Fatal("A editor/history not restored")
	}
	scope := c.selectionScope()
	if c.switchSession("missing") || c.selectionScope() != scope || c.input.cursorPos != 4 {
		t.Fatal("failed switch damaged origin")
	}
	c.switchSession("B")
	if c.input.Text() != "B unsent" || c.input.cursorPos != 2 {
		t.Fatal("B editor not restored")
	}
	for _, id := range []string{"A", "B"} {
		turns, err := c.store.ListTurns(context.Background(), id)
		if err != nil || len(turns) != 0 {
			t.Fatalf("switch submitted model work: %s %v", id, err)
		}
	}
}

func TestSessionEventsRejectSupersededGenerationAndSession(t *testing.T) {
	c := sessionTestChat(t)
	old := c.selectionScope()
	c.switchSession("B")
	for _, returnToA := range []bool{false, true} {
		if returnToA {
			c.switchSession("A")
		}
		before := append([]string(nil), c.transcript...)
		c.handleSessionEvent(sessionEvent{old, map[string]any{"type": "agent_draft_delta", "delta": "stale"}})
		for _, topic := range []string{"turn.draft", "turn.thought", "runtime.tool", "runtime.turn", "runtime.session", "extension.status", "extension.widget", "extension.editor"} {
			c.handleSessionTopicEvent(sessionTopicEvent{old, topics.Envelope{SessionID: "A", Topic: topic, Payload: map[string]any{"delta": "stale", "type": "tool_started", "status": "running", "prompt": "stale prompt", "text": "stale", "key": "stale", "lines": []string{"stale"}}}})
		}
		c.applySessionCompletion(old, func() { c.input.SetText("late completion"); c.switchSession("B") })
		if strings.Join(before, "\n") != strings.Join(c.transcript, "\n") || c.draft != "" || c.running || c.editorAskActive || c.input.Text() != "" || len(c.extensionWidgets) != 0 {
			t.Fatal("stale event/completion mutated selection")
		}
	}
	current := c.selectionScope()
	c.handleSessionEvent(sessionEvent{current, map[string]any{"type": "agent_draft_delta", "session_id": "B", "delta": "wrong"}})
	c.handleSessionTopicEvent(sessionTopicEvent{current, topics.Envelope{SessionID: "B", Topic: "turn.draft", Payload: map[string]any{"delta": "wrong"}}})
	if c.draft != "" {
		t.Fatal("wrong session accepted")
	}
	c.handleSessionEvent(sessionEvent{current, map[string]any{"type": "agent_draft_delta", "delta": "fresh"}})
	if c.draft != "fresh" {
		t.Fatal("current event lost")
	}
}

func TestSessionForwardersCancelWhenTargetIsFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	legacy := make(chan map[string]any, 1)
	legacy <- map[string]any{"type": "agent_status"}
	topic := make(chan topics.Envelope, 1)
	topic <- topics.Envelope{Topic: "turn.status"}
	doneLegacy, doneTopic := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(doneLegacy)
		forwardSessionEvents(ctx, legacy, make(chan sessionEvent), sessionScope{"A", 1})
	}()
	go func() {
		defer close(doneTopic)
		forwardSessionTopics(ctx, topic, make(chan sessionTopicEvent), sessionScope{"A", 1})
	}()
	cancel()
	for _, done := range []chan struct{}{doneLegacy, doneTopic} {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("forwarder remained blocked")
		}
	}
}

func TestBufferedTopicEventCarriesSubscriptionGeneration(t *testing.T) {
	c := sessionTestChat(t)
	original := c.selectionScope()
	c.engine.Topics().Publish(topics.Envelope{SessionID: "A", Topic: "turn.draft", Payload: map[string]any{"delta": "held"}})
	var held sessionTopicEvent
	select {
	case held = <-c.topicEventCh:
	case <-time.After(time.Second):
		t.Fatal("missing bus event")
	}
	if held.scope != original {
		t.Fatal("ownership not captured")
	}
	c.switchSession("B")
	c.switchSession("A")
	c.handleSessionTopicEvent(held)
	if c.draft != "" {
		t.Fatal("buffered old A event accepted on revisit")
	}
}

func TestSessionSwitchCancelsExtensionQuestionWithoutLosingDraft(t *testing.T) {
	c := sessionTestChat(t)
	c.input.SetText("original draft")
	c.input.cursorPos = 3
	c.setEditorAsk("question", "Answer?", "partial answer")
	c.switchSession("B")
	if c.editorAskActive || c.input.Text() != "" {
		t.Fatal("extension prompt followed selection")
	}
	c.switchSession("A")
	if c.input.Text() != "original draft" || c.input.cursorPos != 3 {
		t.Fatal("extension cancellation lost origin draft")
	}
}

func TestSessionPickerFootprintAndUnicodeAtTargetSizes(t *testing.T) {
	c := sessionTestChat(t)
	for i := 0; i < 18; i++ {
		if _, err := c.store.CreateSession(context.Background(), fmt.Sprintf("long-session-%02d", i), strings.Repeat("中文🙂", 20), map[string]any{"status": "idle"}); err != nil {
			t.Fatal(err)
		}
	}
	c.input.SetText("unsent\nsecond line")
	for _, size := range [][2]int{{60, 18}, {100, 22}, {140, 36}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			c.outputWidth, c.outputHeight = size[0], size[1]
			idleFooter := append([]string(nil), c.footerLines(size[0])...)
			before := append([]string(nil), c.transcript...)
			c.openSessionMenu()
			c.setModelMenuSelection(len(c.modelMenuChoices) - 1)
			menu := c.renderModelMenu(size[0])
			if c.modelMenuVisibleRows() > 6 || c.modelMenuHeight() > 8 {
				t.Fatal("selector exceeds footprint")
			}
			if c.modelMenuSelected < c.modelMenuScroll || c.modelMenuSelected >= c.modelMenuScroll+c.modelMenuVisibleRows() {
				t.Fatal("resize hides selection")
			}
			for _, line := range collectElementTexts(menu) {
				if !utf8.ValidString(line) || gotui.StringWidth(line) > size[0] {
					t.Fatalf("invalid/overflowing line: %q", line)
				}
			}
			c.modelMenuQuery = "absent"
			c.applyModelMenuFilter()
			if !strings.Contains(strings.Join(collectElementTexts(c.renderModelMenu(size[0])), "\n"), "no matching sessions") {
				t.Fatal("incorrect empty state")
			}
			c.closeModelMenu()
			if c.modelMenuHeight() != 0 || !reflect.DeepEqual(idleFooter, c.footerLines(size[0])) || !reflect.DeepEqual(before, c.transcript) || c.input.Text() != "unsent\nsecond line" || !c.inputActive {
				t.Fatal("cancel changed idle footprint/draft/focus")
			}
		})
	}
}

func TestSessionMissingModelMetadataDoesNotInheritPreviousSession(t *testing.T) {
	c := sessionTestChat(t)
	if err := c.store.TouchSessionState(context.Background(), "B", map[string]any{"model": "other-model", "provider": "other", "thinking_level": "high"}); err != nil {
		t.Fatal(err)
	}
	c.switchSession("B")
	if c.cfg.DefaultModel != "other-model" || c.cfg.DefaultThinkingLevel != "high" {
		t.Fatal("target model not restored")
	}
	c.switchSession("A")
	if c.cfg.DefaultModel != "bootstrap" || c.cfg.DefaultProvider != "" || c.cfg.DefaultThinkingLevel != "" {
		t.Fatal("missing metadata inherited previous session")
	}
}

func TestSessionQueueRestoreDoesNotReplaceAssistantStream(t *testing.T) {
	c := sessionTestChat(t)
	c.draft = "assistant stream"
	c.queuedDrafts = []string{"followup"}
	c.restoreQueuedDraft()
	if c.input.Text() != "followup" || c.draft != "assistant stream" {
		t.Fatal("editor draft contaminated assistant stream")
	}
}

func TestSessionPickerUsesFullIdentityAndWraps(t *testing.T) {
	c := sessionTestChat(t)
	c.openSessionMenu()
	c.setModelMenuSelection(0)
	c.moveModelMenuSelection(-1)
	if c.modelMenuSelected != len(c.modelMenuChoices)-1 {
		t.Fatal("up did not wrap")
	}
	c.moveModelMenuSelection(1)
	if c.modelMenuSelected != 0 {
		t.Fatal("down did not wrap")
	}
	c.modelMenuQuery = "@B"
	c.applyModelMenuFilter()
	c.acceptModelMenuSelection()
	if c.sessionID != "B" || c.modelMenuOpen {
		t.Fatal("selection failed")
	}
}
