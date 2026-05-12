package tui

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func initialSessionID(ctx context.Context, s *store.Store) (string, error) {
	if sess, err := s.ResolveMainSession(ctx, "agent", "gi", "default"); err == nil {
		return sess.ID, nil
	}
	sessions, err := s.ListSessions(ctx)
	if err != nil {
		return "", fmt.Errorf("list sessions: %w", err)
	}
	if len(sessions) == 0 {
		return "", nil
	}
	return sessions[0].ID, nil
}

func Run(dbPath, workspace, model string) error {
	cfg := config.Load(workspace)
	if model != "" {
		cfg.DefaultModel = model
	}

	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	engine := turn.NewWithRuntimeConfig(s, cfg, cfg.SystemPrompt)

	sessionID, err := initialSessionID(context.Background(), s)
	if err != nil {
		return err
	}
	if sessionID == "" {
		id := store.NowID("session")
		alloc := gisession.AllocateDefaultSession("agent", "gi", "default", id)
		sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(context.Background(), store.ResolveOrCreateSessionFromAllocationInput{ID: id, Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": cfg.DefaultModel, "provider": cfg.DefaultProvider, "thinking_level": cfg.DefaultThinkingLevel}, Allocation: alloc})
		if err != nil {
			return fmt.Errorf("create session: %w", err)
		}
		sessionID = sess.ID
	}

	chat := &chatTUI{
		store:         s,
		engine:        engine,
		sessionID:     sessionID,
		cfg:           cfg,
		transcriptRef: gotui.NewRef(),
		stickToBottom: true,
	}

	app, err := gotui.NewApp(
		gotui.WithMouse(),
		gotui.WithLegacyKeyboard(),
	)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}
	chat.app = app
	app.SetRootComponent(chat)
	return app.Run()
}

type chatTUI struct {
	app              *gotui.App
	store            *store.Store
	engine           *turn.Engine
	sessionID        string
	cfg              config.RuntimeConfig
	history          []string
	histIdx          int
	running          bool
	status           string
	draft            string
	inputActive      bool
	eventCh          chan map[string]any
	topicEventCh     chan topics.Envelope
	subscribedCh     chan map[string]any
	topicUnsubscribe func()
	input            *multilineInput
	inputRegion      *gotui.Element
	transcriptRegion *gotui.Element
	transcriptRef    *gotui.Ref
	transcript       []string
	transcriptScroll int
	stickToBottom    bool
	draftLineIndex   int
}

func (c *chatTUI) ensureInput() {
	if c.input != nil {
		return
	}
	c.input = newMultilineInput(80, "Send a message…", c.onSubmit, nil)
}

func (c *chatTUI) Init() func() {
	c.eventCh = make(chan map[string]any, 64)
	c.bindSession(c.sessionID)
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.histIdx = -1
	c.history = []string{}
	c.inputActive = true
	c.transcript = c.loadTranscript()
	c.draftLineIndex = -1
	if strings.TrimSpace(c.cfg.DefaultModel) == "" {
		c.status = "Select a model with /model <name>"
		c.transcript = append(c.transcript, c.firstUseModelPromptLines()...)
	}
	c.ensureInput()
	c.scrollTranscriptToBottom()

	if c.app != nil {
		time.AfterFunc(200*time.Millisecond, func() {
			c.app.QueueUpdate(func() { c.focusInput() })
		})
	}

	return func() {
		if c.subscribedCh != nil {
			c.engine.Unsubscribe(c.sessionID, c.subscribedCh)
		}
		if c.topicUnsubscribe != nil {
			c.topicUnsubscribe()
			c.topicUnsubscribe = nil
		}
	}
}

func (c *chatTUI) bindSession(sessionID string) {
	if c.eventCh == nil {
		c.eventCh = make(chan map[string]any, 64)
	}
	if c.topicEventCh == nil {
		c.topicEventCh = make(chan topics.Envelope, 64)
	}
	if c.subscribedCh != nil {
		c.engine.Unsubscribe(c.sessionID, c.subscribedCh)
		c.subscribedCh = nil
	}
	if c.topicUnsubscribe != nil {
		c.topicUnsubscribe()
		c.topicUnsubscribe = nil
	}
	c.sessionID = sessionID
	c.subscribedCh = c.engine.Subscribe(sessionID)
	if c.engine.Topics() != nil {
		ch, unsubscribe := c.engine.Topics().Subscribe(context.Background(), "*", topics.SubscribeOptions{Buffer: 64, SessionID: sessionID})
		c.topicUnsubscribe = unsubscribe
		go func(ch <-chan topics.Envelope, target chan topics.Envelope) {
			for env := range ch {
				if target == nil {
					return
				}
				target <- env
			}
		}(ch, c.topicEventCh)
	}
	go func(ch chan map[string]any) {
		for ev := range ch {
			if c.eventCh == nil {
				return
			}
			c.eventCh <- ev
		}
	}(c.subscribedCh)
}

func (c *chatTUI) Watchers() []gotui.Watcher {
	watchers := []gotui.Watcher{gotui.NewChannelWatcher(c.eventCh, c.handleEvent)}
	if c.topicEventCh != nil {
		watchers = append(watchers, gotui.NewChannelWatcher(c.topicEventCh, c.handleTopicEvent))
	}
	return watchers
}

func (c *chatTUI) handleTopicEvent(env topics.Envelope) {
	payload := env.Payload
	switch env.Topic {
	case "turn.status":
		title, _ := payload["title"].(string)
		status, _ := payload["status"].(string)
		if status == "running" {
			c.markRunning()
		}
		if title != "" {
			c.status = title
		} else if status == "idle" {
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
	case "turn.response":
		text := ""
		responseType := ""
		sender, _ := payload["sender"].(string)
		if data, _ := payload["data"].(map[string]any); data != nil {
			text, _ = data["content"].(string)
			responseType, _ = data["type"].(string)
		}
		if text == "" {
			text, _ = payload["content"].(string)
		}
		if text != "" {
			if sender == "system" || responseType == "system_message" {
				c.clearDraftTranscriptLine()
				c.appendTranscript("sys: " + truncate(text, 160))
			} else {
				c.finalizeDraftTranscript(text)
			}
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			c.resetRunningDraftState()
		}
	case "turn.draft":
		delta, _ := payload["delta"].(string)
		if delta != "" {
			c.markRunning()
			c.draft += delta
			c.updateDraftTranscriptLine()
			c.status = fmt.Sprintf("⏳ %s…", truncate(c.draft, 80))
		}
	case "turn.thought":
		c.markRunning()
		c.status = "Thinking…"
	case "runtime.tool":
		toolName, _ := payload["tool"].(string)
		typ, _ := payload["type"].(string)
		switch typ {
		case "tool_started":
			c.markRunning()
			if toolName != "" {
				c.status = fmt.Sprintf("Running: %s", toolName)
			}
		case "tool_finished":
			if toolName != "" {
				c.status = fmt.Sprintf("Tool finished: %s", toolName)
			}
		case "tool_failed":
			errText, _ := payload["error"].(string)
			line := fmt.Sprintf("sys: tool failed: %s", toolName)
			if errText != "" {
				line = fmt.Sprintf("%s: %s", line, truncate(errText, 120))
			}
			c.appendTranscript(line)
			c.status = fmt.Sprintf("Tool failed: %s", toolName)
		case "tool_skipped":
			reason, _ := payload["reason"].(string)
			line := fmt.Sprintf("sys: tool skipped: %s", toolName)
			if reason != "" {
				line = fmt.Sprintf("%s: %s", line, truncate(reason, 120))
			}
			c.appendTranscript(line)
		}
	case "runtime.hook":
		typ, _ := payload["type"].(string)
		hookName, _ := payload["hook"].(string)
		reason, _ := payload["reason"].(string)
		toolName, _ := payload["tool"].(string)
		errText, _ := payload["error"].(string)
		line := ""
		switch typ {
		case "hook_deny", "hook_abort":
			line = fmt.Sprintf("sys: hook %s", typ)
		case "hook_modify":
			line = "sys: hook modified"
		case "hook_respond":
			line = "sys: hook responded directly"
		case "hook_invocation":
			if errText != "" {
				line = "sys: hook invocation error"
			}
		}
		if line != "" {
			if hookName != "" {
				line += " via " + hookName
			}
			if toolName != "" {
				line += " for " + toolName
			}
			if reason != "" {
				line += ": " + truncate(reason, 120)
			} else if errText != "" {
				line += ": " + truncate(errText, 120)
			}
			c.appendTranscript(line)
		}
	case "runtime.turn":
		typ, _ := payload["type"].(string)
		status, _ := payload["status"].(string)
		switch typ {
		case "turn_completed":
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		case "turn_terminal":
			if status == "failed" || status == "aborted" || status == "cancelled" {
				c.resetRunningDraftState()
				c.status = fmt.Sprintf("Turn %s", status)
			}
		case "turn_state":
			phase, _ := payload["phase"].(string)
			if status == "running" && phase == "waiting_on_tools" {
				c.markRunning()
				toolName, _ := payload["tool"].(string)
				if toolName != "" {
					c.status = fmt.Sprintf("Running: %s", toolName)
				}
			}
		}
	case "runtime.session":
		typ, _ := payload["type"].(string)
		status, _ := payload["status"].(string)
		switch typ {
		case "session_running", "session_state":
			if status == "running" {
				c.markRunning()
				if c.status == "" {
					c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				}
			}
			if status == "idle" {
				c.resetRunningDraftState()
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			}
		case "session_idle":
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
	case "runtime.routing":
		typ, _ := payload["type"].(string)
		targetAgent, _ := payload["target_agent_id"].(string)
		targetSession, _ := payload["target_session"].(string)
		if targetSession == "" {
			targetSession, _ = payload["target_session_id"].(string)
		}
		sourceAgent, _ := payload["source_agent_id"].(string)
		switch typ {
		case "routing_decision":
			if targetAgent != "" {
				line := fmt.Sprintf("sys: routed to @%s", targetAgent)
				if targetSession != "" {
					line += fmt.Sprintf(" (%s)", targetSession)
				}
				c.appendTranscript(line)
			}
		case "routing_incoming":
			if sourceAgent != "" {
				line := fmt.Sprintf("sys: incoming route from @%s", sourceAgent)
				if targetSession != "" {
					line += fmt.Sprintf(" (%s)", targetSession)
				}
				c.appendTranscript(line)
			}
		}
	case "runtime.inbound_work":
		typ, _ := payload["type"].(string)
		sourceKind, _ := payload["source_kind"].(string)
		status, _ := payload["status"].(string)
		attemptCount := intFromAny(payload["attempt_count"])
		errText, _ := payload["error"].(string)
		line := ""
		switch typ {
		case "inbound_work_enqueued":
			line = fmt.Sprintf("sys: inbound work queued (%s)", sourceKind)
		case "inbound_work_retry_scheduled":
			line = fmt.Sprintf("sys: inbound work retry scheduled (%s)", sourceKind)
			if attemptCount > 0 {
				line += fmt.Sprintf(" attempt %d", attemptCount)
			}
		case "inbound_work_failed":
			line = fmt.Sprintf("sys: inbound work failed (%s)", sourceKind)
		case "inbound_work_completed":
			line = fmt.Sprintf("sys: inbound work completed (%s)", sourceKind)
		case "inbound_work_requeued":
			line = fmt.Sprintf("sys: inbound work requeued (%s)", sourceKind)
		case "inbound_work_discarded":
			line = fmt.Sprintf("sys: inbound work discarded (%s)", sourceKind)
		}
		if line != "" {
			if status != "" {
				line += fmt.Sprintf(" [%s]", status)
			}
			if errText != "" {
				line += ": " + truncate(errText, 120)
			}
			c.appendTranscript(line)
		}
	case "session.compaction":
		before, _ := payload["messages_before"].(int)
		after, _ := payload["messages_after"].(int)
		tokens, _ := payload["tokens_before"].(int)
		if before == 0 {
			before = intFromAny(payload["messages_before"])
		}
		if after == 0 {
			after = intFromAny(payload["messages_after"])
		}
		if tokens == 0 {
			tokens = intFromAny(payload["tokens_before"])
		}
		c.appendTranscript(fmt.Sprintf("sys: compacted context: messages %d→%d, tokens_before=%d", before, after, tokens))
		c.status = "Compacted context"
	case "session.routing":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		typ, _ := payload["type"].(string)
		targetAgent, _ := payload["target_agent_id"].(string)
		targetSession, _ := payload["target_session"].(string)
		if targetSession == "" {
			targetSession, _ = payload["target_session_id"].(string)
		}
		sourceAgent, _ := payload["source_agent_id"].(string)
		switch typ {
		case "routing_decision":
			if targetAgent != "" {
				line := fmt.Sprintf("sys: routed to @%s", targetAgent)
				if targetSession != "" {
					line += fmt.Sprintf(" (%s)", targetSession)
				}
				c.appendTranscript(line)
			}
		case "routing_incoming":
			if sourceAgent != "" {
				line := fmt.Sprintf("sys: incoming route from @%s", sourceAgent)
				if targetSession != "" {
					line += fmt.Sprintf(" (%s)", targetSession)
				}
				c.appendTranscript(line)
			}
		}
	case "session.steering":
		typ, _ := payload["type"].(string)
		switch typ {
		case "steering_enqueued":
			c.status = "Queued follow-up"
		case "steering_injected":
			c.status = "Injected follow-up"
		case "steering_continued":
			c.status = "Continuing queued follow-up"
		}
	case "turn.subturn":
		typ, _ := payload["type"].(string)
		childTurn, _ := payload["child_turn_id"].(string)
		status, _ := payload["status"].(string)
		switch typ {
		case "subturn_created":
			if childTurn != "" {
				c.appendTranscript(fmt.Sprintf("sys: sub-turn started: %s", childTurn))
			}
		case "subturn_status":
			if childTurn != "" && status != "" {
				c.appendTranscript(fmt.Sprintf("sys: sub-turn %s: %s", childTurn, status))
			}
		case "subturn_result_ready", "subturn_result_delivered", "subturn_orphaned":
			if childTurn != "" && status != "" {
				c.appendTranscript(fmt.Sprintf("sys: sub-turn %s result: %s", childTurn, status))
			}
		}
	}
	if c.stickToBottom {
		c.scrollTranscriptToBottom()
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) useTopicNativeRuntimeStatus() bool {
	return c.topicUnsubscribe != nil
}

func (c *chatTUI) handleEvent(ev map[string]any) {
	evType, _ := ev["type"].(string)
	switch evType {
	case "agent_draft_delta":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		delta, _ := ev["delta"].(string)
		c.markRunning()
		c.draft += delta
		c.updateDraftTranscriptLine()
		c.status = fmt.Sprintf("⏳ %s…", truncate(c.draft, 80))
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "new_post":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		data, _ := ev["data"].(map[string]any)
		if data != nil {
			text, _ := data["content"].(string)
			if text != "" {
				c.finalizeDraftTranscript(text)
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				c.resetRunningDraftState()
				if c.stickToBottom {
					c.scrollTranscriptToBottom()
				}
				if c.app != nil {
					c.app.MarkDirty()
				}
			}
		}
	case "agent_thought_delta":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		c.markRunning()
		c.status = "Thinking…"
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "tool_finished":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		toolName, _ := ev["tool"].(string)
		if toolName != "" {
			c.status = fmt.Sprintf("Tool finished: %s", toolName)
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "tool_failed":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		toolName, _ := ev["tool"].(string)
		errText, _ := ev["error"].(string)
		line := fmt.Sprintf("sys: tool failed: %s", toolName)
		if errText != "" {
			line = fmt.Sprintf("%s: %s", line, truncate(errText, 120))
		}
		c.appendTranscript(line)
		c.status = fmt.Sprintf("Tool failed: %s", toolName)
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "compaction":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		before := intFromEvent(ev, "messages_before")
		after := intFromEvent(ev, "messages_after")
		tokens := intFromEvent(ev, "tokens_before")
		c.appendTranscript(fmt.Sprintf("sys: compacted context: messages %d→%d, tokens_before=%d", before, after, tokens))
		c.status = "Compacted context"
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "agent_status":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		title := ""
		if v, ok := ev["title"].(string); ok {
			title = v
		}
		if title != "" {
			c.markRunning()
			c.status = title
		} else {
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "error":
		msg := ""
		if v, ok := ev["error"].(string); ok {
			msg = v
		}
		if msg == "" {
			if v, ok := ev["message"].(string); ok {
				msg = v
			}
		}
		if msg != "" {
			c.resetRunningDraftState()
			c.appendTranscript("error: " + truncate(msg, 160))
			c.status = "Error"
			if c.app != nil {
				c.app.MarkDirty()
			}
		}
	}
}

func (c *chatTUI) updateDraftTranscriptLine() {
	line := fmt.Sprintf("%s: %s", c.cfg.AssistantName, c.draft)
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		c.transcript[c.draftLineIndex] = line
		return
	}
	c.appendTranscript(line)
	c.draftLineIndex = len(c.transcript) - 1
}

func (c *chatTUI) finalizeDraftTranscript(text string) {
	lines := renderMarkdownTranscript(c.cfg.AssistantName+": ", text, c.transcriptRenderWidth())
	if !looksLikeMarkdown(text) {
		lines = []string{fmt.Sprintf("%s: %s", c.cfg.AssistantName, text)}
	}
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		prefix := append([]string(nil), c.transcript[:c.draftLineIndex]...)
		suffix := append([]string(nil), c.transcript[c.draftLineIndex+1:]...)
		c.transcript = append(prefix, append(lines, suffix...)...)
		c.applyTranscriptLimit()
		c.draftLineIndex = -1
		return
	}
	c.appendTranscript(lines...)
}

func (c *chatTUI) clearDraftTranscriptLine() {
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		c.transcript = append(c.transcript[:c.draftLineIndex], c.transcript[c.draftLineIndex+1:]...)
	}
	c.draft = ""
	c.draftLineIndex = -1
}

func (c *chatTUI) markRunning() {
	c.running = true
}

func (c *chatTUI) resetRunningDraftState() {
	c.clearDraftTranscriptLine()
	c.running = false
	c.draft = ""
	c.draftLineIndex = -1
}

func intFromEvent(ev map[string]any, key string) int {
	return intFromAny(ev[key])
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	default:
		return 0
	}
}

func (c *chatTUI) KeyMap() gotui.KeyMap {
	return gotui.KeyMap{
		gotui.OnStop(gotui.KeyCtrlC, func(ke gotui.KeyEvent) { c.app.Stop() }),
		gotui.OnStop(gotui.KeyCtrlD, func(ke gotui.KeyEvent) {
			if c.input.Text() == "" {
				c.app.Stop()
			}
		}),
		gotui.OnStop(gotui.KeyEscape, func(ke gotui.KeyEvent) {
			c.inputActive = false
			if c.app != nil {
				c.app.BlurFocused()
			}
		}),
		gotui.OnStop(gotui.KeyTab, func(ke gotui.KeyEvent) { c.focusInput() }),
		gotui.OnStop(gotui.KeyPageUp, func(ke gotui.KeyEvent) { c.pageTranscript(-1) }),
		gotui.OnStop(gotui.KeyPageDown, func(ke gotui.KeyEvent) { c.pageTranscript(1) }),
		gotui.OnStop(gotui.KeyHome, func(ke gotui.KeyEvent) { c.scrollTranscriptToTop() }),
		gotui.OnStop(gotui.KeyEnd, func(ke gotui.KeyEvent) { c.scrollTranscriptToBottom() }),
		gotui.OnPreemptStop(gotui.KeyF2, func(ke gotui.KeyEvent) { c.recallHistory(-1) }),
		gotui.OnPreemptStop(gotui.KeyF3, func(ke gotui.KeyEvent) { c.recallHistory(1) }),
		gotui.OnPreemptStop(gotui.Rune('p').Ctrl(), func(ke gotui.KeyEvent) { c.recallHistory(-1) }),
		gotui.OnPreemptStop(gotui.Rune('n').Ctrl(), func(ke gotui.KeyEvent) { c.recallHistory(1) }),
		gotui.OnPreemptStop(gotui.KeyUp, func(ke gotui.KeyEvent) {
			c.recallHistory(-1)
		}),
		gotui.OnPreemptStop(gotui.KeyDown, func(ke gotui.KeyEvent) {
			c.recallHistory(1)
		}),
	}
}

func (c *chatTUI) recallHistory(delta int) {
	if c.input == nil || c.input.Text() != "" || len(c.history) == 0 {
		return
	}
	if delta < 0 {
		if c.histIdx < 0 {
			c.histIdx = len(c.history) - 1
		} else if c.histIdx > 0 {
			c.histIdx--
		}
	} else {
		if c.histIdx < 0 {
			return
		}
		c.histIdx++
		if c.histIdx >= len(c.history) {
			c.histIdx = -1
			c.input.SetText("")
			return
		}
	}
	c.focusInput()
	c.input.SetText(c.history[c.histIdx])
}

func (c *chatTUI) HandleMouse(me gotui.MouseEvent) bool {
	switch me.Button {
	case gotui.MouseWheelUp:
		c.scrollTranscript(-3)
		return true
	case gotui.MouseWheelDown:
		c.scrollTranscript(3)
		return true
	}
	if me.Button != gotui.MouseLeft || me.Action != gotui.MousePress {
		return false
	}
	if c.inputRegion != nil && c.inputRegion.ContainsPoint(me.X, me.Y) {
		c.focusInput()
		return true
	}
	if c.app != nil {
		_, h := c.app.Size()
		if me.Y >= h-4 {
			c.focusInput()
			return true
		}
	}
	return false
}

func (c *chatTUI) focusInput() {
	c.inputActive = true
	if c.app != nil && c.app.Focused() == nil {
		c.app.FocusNext()
	}
}

func (c *chatTUI) onSubmit(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	c.history = append(c.history, text)
	c.histIdx = -1
	c.input.SetText("")
	if strings.HasPrefix(text, "/") {
		c.handleCommand(text)
		return
	}
	if c.running {
		return
	}
	if strings.TrimSpace(c.cfg.DefaultModel) == "" {
		c.transcript = append(c.transcript, c.firstUseModelPromptLines()...)
		c.status = "Select a model with /model <name>"
		c.scrollTranscriptToBottom()
		if c.app != nil {
			c.app.MarkDirty()
		}
		return
	}
	c.running = true
	c.status = "sending…"
	c.draft = ""
	c.draftLineIndex = -1
	c.appendTranscript(fmt.Sprintf("you: %s", text))
	c.stickToBottom = true
	c.scrollTranscriptToBottom()
	if c.app != nil {
		c.app.MarkDirty()
	}

	go func() {
		result, err := c.engine.SubmitPromptRouted(context.Background(), turn.RunInput{
			SessionID: c.sessionID,
			Prompt:    text,
			Intent:    "prompt",
			Model:     c.cfg.DefaultModel,
		})
		if err != nil {
			if c.app != nil {
				c.app.QueueUpdate(func() {
					c.clearDraftTranscriptLine()
					c.appendTranscript(fmt.Sprintf("error: %v", err))
					c.running = false
					c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
					c.app.MarkDirty()
				})
			} else {
				c.clearDraftTranscriptLine()
				c.appendTranscript(fmt.Sprintf("error: %v", err))
				c.running = false
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			}
			return
		}
		if result != nil && result.Routed && result.SessionID != c.sessionID {
			if c.app != nil {
				c.app.QueueUpdate(func() {
					c.clearDraftTranscriptLine()
					c.appendTranscript(fmt.Sprintf("sys: routed to @%s (%s)", result.TargetAgentID, result.SessionID))
					c.running = false
					c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
					c.app.MarkDirty()
				})
			} else {
				c.clearDraftTranscriptLine()
				c.appendTranscript(fmt.Sprintf("sys: routed to @%s (%s)", result.TargetAgentID, result.SessionID))
				c.running = false
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			}
		}
	}()
}

func (c *chatTUI) handleCommand(text string) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return
	}
	switch fields[0] {
	case "/help":
		c.transcript = append(c.transcript, c.helpLines()...)
	case "/tools":
		c.transcript = append(c.transcript, c.toolCommand(fields)...)
	case "/skills":
		query := ""
		if len(fields) > 1 {
			query = strings.Join(fields[1:], " ")
		}
		c.transcript = append(c.transcript, c.skillLines(query)...)
	case "/model":
		c.transcript = append(c.transcript, c.modelCommand(fields)...)
	case "/thinking":
		c.transcript = append(c.transcript, c.thinkingCommand(fields)...)
	case "/compact":
		c.appendTranscript(c.compactLines()...)
	case "/scrollback":
		c.appendTranscript(c.scrollbackCommand(fields)...)
	case "/settings", "/config":
		c.appendTranscript(c.settingsLines()...)
	case "/approvals":
		c.appendTranscript("approvals: no approval gates are configured in gi yet")
	case "/cancel":
		c.appendTranscript(c.cancelCommand())
	case "/agents":
		c.transcript = append(c.transcript, c.listAgentLines()...)
	case "/plugins", "/extensions":
		c.transcript = append(c.transcript, c.pluginLines()...)
	case "/tree":
		c.transcript = append(c.transcript, c.treeLines()...)
	case "/fork":
		target := ""
		if len(fields) > 1 {
			target = strings.TrimPrefix(fields[1], "@")
		}
		if target == "" {
			target = c.nextForkAgentID()
		}
		child, _, err := c.engine.ResolveOrCreatePeerSession(context.Background(), c.sessionID, target)
		if err != nil {
			c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
			break
		}
		c.switchSession(child.ID)
		c.transcript = append(c.transcript, fmt.Sprintf("sys: switched to @%s (%s)", target, child.ID))
	case "/switch":
		if len(fields) < 2 {
			c.transcript = append(c.transcript, "sys: usage /switch @agent|session_id")
			break
		}
		sess, err := c.resolveSessionRef(fields[1])
		if err != nil {
			c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
			break
		}
		c.switchSession(sess.ID)
		c.transcript = append(c.transcript, fmt.Sprintf("sys: switched to @%s (%s)", c.agentIDForSession(sess), sess.ID))
	case "/send":
		if len(fields) < 3 {
			c.transcript = append(c.transcript, "sys: usage /send @agent message")
			break
		}
		target := strings.TrimPrefix(fields[1], "@")
		body := strings.TrimSpace(strings.TrimPrefix(text, fields[0]+" "+fields[1]))
		c.transcript = append(c.transcript, fmt.Sprintf("you → @%s: %s", target, body))
		c.running = true
		go func() {
			result, err := c.engine.SubmitPeerMessage(context.Background(), c.sessionID, target, body, "prompt", c.cfg.DefaultModel, "")
			if err != nil {
				if c.app != nil {
					c.app.QueueUpdate(func() {
						c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
						c.running = false
						c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
						c.app.MarkDirty()
					})
				} else {
					c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
					c.running = false
					c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				}
				return
			}
			if c.app != nil {
				c.app.QueueUpdate(func() {
					c.transcript = append(c.transcript, fmt.Sprintf("sys: delivered to @%s (%s)", target, result.SessionID))
					c.running = false
					c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
					c.app.MarkDirty()
				})
			} else {
				c.transcript = append(c.transcript, fmt.Sprintf("sys: delivered to @%s (%s)", target, result.SessionID))
				c.running = false
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
			}
		}()
		return
	case "/where":
		c.transcript = append(c.transcript, c.contextSummary())
	default:
		c.appendTranscript("sys: commands: /help, /tools [query|active|activate|reset], /skills [query], /model [name], /thinking [level], /compact, /scrollback [n], /settings, /approvals, /cancel, /agents, /tree, /plugins, /fork [@agentN], /switch @agent|session_id, /send @agent message, /where")
	}
	c.running = false
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.stickToBottom = true
	c.scrollTranscriptToBottom()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) firstUseModelPromptLines() []string {
	if len(c.cfg.EnabledModels) > 0 {
		return []string{fmt.Sprintf("sys: no model selected; choose one with /model <name>. available: %s", strings.Join(c.cfg.EnabledModels, ", "))}
	}
	return []string{"sys: no model selected; choose one with /model <provider/model> before sending your first prompt"}
}

func (c *chatTUI) helpLines() []string {
	return []string{
		"sys: gi TUI help",
		"sys: keys: Enter send · Esc blur input · Ctrl-C/Ctrl-D quit · Up/Down history · PgUp/PgDn scroll · Home/End transcript",
		"sys: commands: /help, /tools [query|active|activate|reset], /skills [query], /model [name], /thinking [level], /compact, /scrollback [n], /settings, /approvals, /cancel, /agents, /tree, /plugins, /where",
		"sys: sessions: /fork [@agentN], /switch @agent|session_id, /send @agent message",
	}
}

func (c *chatTUI) modelCommand(fields []string) []string {
	if len(fields) == 1 {
		if strings.TrimSpace(c.cfg.DefaultModel) == "" {
			return c.firstUseModelPromptLines()
		}
		return []string{fmt.Sprintf("sys: model: %s", c.cfg.DefaultModel)}
	}
	model := strings.TrimSpace(strings.Join(fields[1:], " "))
	if model == "" {
		return []string{"sys: usage /model <model>"}
	}
	c.cfg.DefaultModel = model
	if strings.Contains(model, "/") && strings.TrimSpace(c.cfg.DefaultProvider) == "" {
		c.cfg.DefaultProvider = strings.SplitN(model, "/", 2)[0]
	}
	lines := []string{fmt.Sprintf("sys: model set to %s", model)}
	if err := c.store.TouchSessionState(context.Background(), c.sessionID, map[string]any{"model": model}); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist model in session state: %v", err))
	}
	if err := config.PersistModelSelection(c.cfg.WorkspaceRoot, c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel, c.cfg.EnabledModels); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist model selection: %v", err))
	}
	return lines
}

func (c *chatTUI) thinkingCommand(fields []string) []string {
	if len(fields) == 1 {
		return []string{fmt.Sprintf("sys: thinking: %s", c.cfg.DefaultThinkingLevel)}
	}
	level := strings.TrimSpace(fields[1])
	if level == "" {
		return []string{"sys: usage /thinking <low|medium|high>"}
	}
	c.cfg.DefaultThinkingLevel = level
	lines := []string{fmt.Sprintf("sys: thinking set to %s", level)}
	if err := c.store.TouchSessionState(context.Background(), c.sessionID, map[string]any{"thinking_level": level}); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist thinking level in session state: %v", err))
	}
	return lines
}

func (c *chatTUI) currentScrollbackLimit() int {
	if c.cfg.ScrollbackLimit > 0 {
		return c.cfg.ScrollbackLimit
	}
	return 1000
}

func (c *chatTUI) pruneTranscript(lines []string) []string {
	limit := c.currentScrollbackLimit()
	if limit <= 0 || len(lines) <= limit {
		return lines
	}
	return append([]string(nil), lines[len(lines)-limit:]...)
}

func (c *chatTUI) appendTranscript(lines ...string) {
	if len(lines) == 0 {
		return
	}
	c.transcript = append(c.transcript, lines...)
	c.applyTranscriptLimit()
}

func (c *chatTUI) applyTranscriptLimit() {
	limit := c.currentScrollbackLimit()
	if limit <= 0 || len(c.transcript) <= limit {
		return
	}
	drop := len(c.transcript) - limit
	c.transcript = append([]string(nil), c.transcript[drop:]...)
	if c.draftLineIndex >= 0 {
		c.draftLineIndex -= drop
		if c.draftLineIndex < 0 {
			c.draftLineIndex = -1
		}
	}
}

func (c *chatTUI) compactLines() []string {
	messages, _ := c.store.ListMessages(context.Background(), c.sessionID)
	turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
	settings := c.cfg.Compaction
	return []string{
		fmt.Sprintf("compact: enabled=%v threshold_tokens=%d keep_recent_tokens=%d reserve_tokens=%d strategy=%s", settings.Enabled, settings.ThresholdTokens, settings.KeepRecentTokens, settings.ReserveTokens, settings.Strategy),
		fmt.Sprintf("compact: messages=%d turns=%d; use the agent `compact` tool for full JSON preparation", len(messages), len(turns)),
	}
}

func (c *chatTUI) settingsLines() []string {
	return []string{
		"settings: runtime:",
		fmt.Sprintf("- provider=%s model=%s thinking=%s max_iterations=%d", c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel, c.cfg.MaxIterations),
		fmt.Sprintf("- scrollback_limit=%d", c.currentScrollbackLimit()),
		fmt.Sprintf("- workspace=%s", c.cfg.WorkspaceRoot),
		fmt.Sprintf("- compaction enabled=%v threshold_tokens=%d keep_recent_tokens=%d reserve_tokens=%d", c.cfg.Compaction.Enabled, c.cfg.Compaction.ThresholdTokens, c.cfg.Compaction.KeepRecentTokens, c.cfg.Compaction.ReserveTokens),
		fmt.Sprintf("- peering enabled=%v hostname=%s auth_key_env=%s auth_key_keychain=%s", c.cfg.Peering.Enabled, c.cfg.Peering.Hostname, c.cfg.Peering.AuthKeyEnv, c.cfg.Peering.AuthKeyKeychain),
		fmt.Sprintf("- tools active=%s", strings.Join(c.engine.ActiveTools(), ", ")),
	}
}

func (c *chatTUI) scrollbackCommand(fields []string) []string {
	if len(fields) == 1 {
		return []string{fmt.Sprintf("sys: scrollback limit: %d", c.currentScrollbackLimit())}
	}
	limit, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil || limit <= 0 {
		return []string{"sys: usage /scrollback <positive-limit>"}
	}
	c.cfg.ScrollbackLimit = limit
	c.applyTranscriptLimit()
	lines := []string{fmt.Sprintf("sys: scrollback limit set to %d", limit)}
	if err := config.PersistScrollbackLimit(c.cfg.WorkspaceRoot, limit); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist scrollback limit: %v", err))
	}
	return lines
}

func (c *chatTUI) cancelCommand() string {
	turns, err := c.store.ListTurns(context.Background(), c.sessionID)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	for i := len(turns) - 1; i >= 0; i-- {
		if turns[i].Status == "running" || turns[i].Status == "queued" || turns[i].Status == "cancelling" {
			if err := c.engine.CancelTurn(context.Background(), c.sessionID, turns[i].ID); err != nil {
				return fmt.Sprintf("error: %v", err)
			}
			c.running = false
			return fmt.Sprintf("sys: cancellation requested for %s", turns[i].ID)
		}
	}
	c.running = false
	return "sys: no running or queued turn to cancel"
}

func (c *chatTUI) toolCommand(fields []string) []string {
	if len(fields) > 1 {
		switch fields[1] {
		case "active":
			active := c.engine.ActiveTools()
			if len(active) == 0 {
				return []string{"tools: active: (none)"}
			}
			return []string{fmt.Sprintf("tools: active: %s", strings.Join(active, ", "))}
		case "activate":
			if len(fields) < 3 {
				return []string{"tools: usage /tools activate <tool> [tool...]"}
			}
			names := make([]any, 0, len(fields)-2)
			for _, name := range fields[2:] {
				if strings.TrimSpace(name) != "" {
					names = append(names, strings.TrimSpace(name))
				}
			}
			out, err := c.engine.ExecuteToolsMeta(map[string]any{"activate": names})
			if err != nil {
				return []string{fmt.Sprintf("error: %v", err)}
			}
			return prefixMultiline("tools", out)
		case "reset":
			out, err := c.engine.ExecuteToolsMeta(map[string]any{"reset_active": true})
			if err != nil {
				return []string{fmt.Sprintf("error: %v", err)}
			}
			return prefixMultiline("tools", out)
		}
	}
	query := ""
	if len(fields) > 1 {
		query = strings.Join(fields[1:], " ")
	}
	args := map[string]any{"include_inactive": true}
	if strings.TrimSpace(query) != "" {
		args["query"] = strings.TrimSpace(query)
	}
	out, err := c.engine.ExecuteToolsMeta(args)
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	return prefixMultiline("tools", out)
}

func (c *chatTUI) skillLines(query string) []string {
	args := map[string]any{}
	if strings.TrimSpace(query) != "" {
		args["query"] = strings.TrimSpace(query)
	}
	out, err := turn.ExecuteSkillsMeta(c.cfg.WorkspaceRoot, args)
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	return prefixMultiline("skills", out)
}

func prefixMultiline(prefix, text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{prefix + ": (empty)"}
	}
	parts := strings.Split(text, "\n")
	out := make([]string, 0, len(parts))
	for i, line := range parts {
		if i == 0 {
			out = append(out, prefix+": "+line)
		} else {
			out = append(out, "  "+line)
		}
	}
	return out
}

func (c *chatTUI) pluginLines() []string {
	lines := []string{"plugins: extensions:"}
	extensions := c.engine.ExtensionInfos()
	if len(extensions) == 0 {
		lines = append(lines, "- none loaded")
	} else {
		for _, ext := range extensions {
			line := fmt.Sprintf("- %s %s %s", ext.Status, ext.Engine, ext.Path)
			if ext.Error != "" {
				line += ": " + truncate(ext.Error, 120)
			}
			lines = append(lines, line)
		}
	}
	hooks := c.engine.HookInfos()
	lines = append(lines, fmt.Sprintf("plugins: hooks: %d", len(hooks)))
	for _, hook := range hooks {
		lines = append(lines, fmt.Sprintf("- %s from %s #%d", hook.Name, hook.Source, hook.ID))
	}
	return lines
}

func (c *chatTUI) Render(app *gotui.App) *gotui.Element {
	w, h := app.Size()
	padding := 1
	if w < 80 || h < 20 {
		padding = 0
	}
	contentWidth := w - (padding * 2)
	if contentWidth < 20 {
		contentWidth = 20
	}

	root := gotui.New(
		gotui.WithDirection(gotui.Column),
		gotui.WithWidthPercent(100),
		gotui.WithHeightPercent(100),
		gotui.WithPadding(padding),
		gotui.WithGap(0),
	)

	statusLines := c.wrapLines(fmt.Sprintf("Status: %s", c.status), contentWidth)
	root.AddChild(c.renderLineBlock(statusLines, gotui.NewStyle().Bold()))

	ctxLines := c.contextSummaryLines(contentWidth)
	root.AddChild(c.renderLineBlock(ctxLines, gotui.NewStyle().Dim()))

	sep := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithText(c.horizontalRule(contentWidth)),
		gotui.WithTextStyle(gotui.NewStyle().Dim()),
	)
	root.AddChild(sep)

	c.ensureInput()
	c.input.width = contentWidth
	inputHeight := c.input.Render(app).HeightForWidth(contentWidth)
	if inputHeight < 1 {
		inputHeight = 1
	}
	footerLines := c.wrapLines(c.footerText(), contentWidth)
	reservedHeight := (padding * 2) + len(statusLines) + len(ctxLines) + 3 + inputHeight + len(footerLines)
	transcriptHeight := h - reservedHeight
	if transcriptHeight < 4 {
		transcriptHeight = 4
	}
	transcript := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithHeight(transcriptHeight),
		gotui.WithScrollable(gotui.ScrollVertical),
		gotui.WithScrollOffset(0, c.transcriptScroll),
		gotui.WithScrollbarStyle(gotui.NewStyle().Dim()),
	)
	c.transcriptRef.Set(transcript)
	c.transcriptRegion = transcript
	for _, line := range c.visibleTranscript() {
		transcript.AddChild(gotui.New(
			gotui.WithWidthPercent(100),
			gotui.WithText(line),
		))
	}
	root.AddChild(transcript)

	inputTopSep := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithText(c.horizontalRule(contentWidth)),
		gotui.WithTextStyle(gotui.NewStyle().Dim()),
	)
	root.AddChild(inputTopSep)

	inputEl := app.MountPersistent(c, 0, func() gotui.Component { return c.input })
	c.inputRegion = inputEl
	root.AddChild(inputEl)

	inputBottomSep := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithText(c.horizontalRule(contentWidth)),
		gotui.WithTextStyle(gotui.NewStyle().Dim()),
	)
	root.AddChild(inputBottomSep)

	root.AddChild(c.renderLineBlock(footerLines, gotui.NewStyle().Dim()))

	return root
}

func (c *chatTUI) footerText() string {
	return "Hints: /help · /tools active|activate|reset · /skills · /model · /thinking · /compact · /scrollback · /settings · /approvals · /cancel · /agents · /tree · /plugins · /fork · /switch · /send · Esc blur · Tab focus · F2/F3 history · PgUp/PgDn scroll · Ctrl-D quit"
}

func (c *chatTUI) currentPadding() int {
	if c.app != nil {
		w, h := c.app.Size()
		if w < 80 || h < 20 {
			return 0
		}
	}
	return 1
}

func (c *chatTUI) currentContentWidth() int {
	if c.app == nil {
		return 78
	}
	w, h := c.app.Size()
	padding := 1
	if w < 80 || h < 20 {
		padding = 0
	}
	contentWidth := w - (padding * 2)
	if contentWidth < 20 {
		contentWidth = 20
	}
	return contentWidth
}

func (c *chatTUI) horizontalRule(width int) string {
	if width < 8 {
		width = 8
	}
	return strings.Repeat("─", width)
}

func (c *chatTUI) wrapLines(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	out := []string{}
	for _, raw := range strings.Split(text, "\n") {
		runes := []rune(raw)
		if len(runes) == 0 {
			out = append(out, "")
			continue
		}
		for len(runes) > width {
			out = append(out, string(runes[:width]))
			runes = runes[width:]
		}
		out = append(out, string(runes))
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}

func (c *chatTUI) renderLineBlock(lines []string, style gotui.Style) *gotui.Element {
	block := gotui.New(
		gotui.WithDirection(gotui.Column),
		gotui.WithWidthPercent(100),
		gotui.WithHeight(len(lines)),
	)
	for _, line := range lines {
		block.AddChild(gotui.New(
			gotui.WithWidthPercent(100),
			gotui.WithText(line),
			gotui.WithTextStyle(style),
		))
	}
	return block
}

func (c *chatTUI) transcriptRenderWidth() int {
	width := c.currentContentWidth()
	if width <= 0 {
		return 80
	}
	return width
}

func (c *chatTUI) loadTranscript() []string {
	if c.store == nil {
		return append([]string(nil), c.transcript...)
	}
	msgs, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, c.renderMessageLines(m, c.transcriptRenderWidth())...)
	}
	return c.pruneTranscript(out)
}

func (c *chatTUI) renderMessageLines(m store.Message, width int) []string {
	kind, _ := m.Payload["kind"].(string)
	if kind == "compaction" || m.Role == "tool_result" || kind == "tool_result" {
		return []string{c.renderMessageLine(m)}
	}
	prefix := "you: "
	switch m.Role {
	case "assistant":
		prefix = c.cfg.AssistantName + ": "
	case "system":
		prefix = "sys: "
	}
	if looksLikeMarkdown(m.Content) {
		return renderMarkdownTranscript(prefix, m.Content, width)
	}
	return []string{c.renderMessageLine(m)}
}

func (c *chatTUI) renderMessageLine(m store.Message) string {
	kind, _ := m.Payload["kind"].(string)
	switch {
	case kind == "compaction":
		tokens := toInt(m.Payload["tokens_before"], 0)
		return fmt.Sprintf("compact: %s (tokens_before=%d)", truncate(m.Content, 160), tokens)
	case m.Role == "tool_result" || kind == "tool_result":
		toolName, _ := m.Payload["tool_name"].(string)
		if toolName == "" {
			toolName = "tool"
		}
		isErr, _ := m.Payload["is_error"].(bool)
		status := "ok"
		if isErr {
			status = "error"
		}
		return fmt.Sprintf("tool[%s/%s]: %s", toolName, status, truncate(m.Content, 160))
	}
	prefix := "you"
	switch m.Role {
	case "assistant":
		prefix = c.cfg.AssistantName
	case "system":
		prefix = "sys"
	}
	return fmt.Sprintf("%s: %s", prefix, truncate(m.Content, 200))
}

func toInt(value any, fallback int) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return fallback
	}
}

func (c *chatTUI) transcriptViewportHeight() int {
	if c.transcriptRef == nil || c.transcriptRef.El() == nil {
		return 4
	}
	_, h := c.transcriptRef.El().ViewportSize()
	if h <= 0 {
		return 4
	}
	return h
}

func (c *chatTUI) scrollTranscript(delta int) {
	lines := c.visibleTranscript()
	maxScroll := len(lines) - c.transcriptViewportHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}
	c.transcriptScroll += delta
	if c.transcriptScroll < 0 {
		c.transcriptScroll = 0
	}
	if c.transcriptScroll > maxScroll {
		c.transcriptScroll = maxScroll
	}
	c.stickToBottom = c.transcriptScroll >= maxScroll
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) pageTranscript(delta int) {
	step := c.transcriptViewportHeight() - 1
	if step < 1 {
		step = 1
	}
	c.scrollTranscript(delta * step)
}

func (c *chatTUI) scrollTranscriptToTop() {
	c.transcriptScroll = 0
	c.stickToBottom = false
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) scrollTranscriptToBottom() {
	lines := c.visibleTranscript()
	maxScroll := len(lines) - c.transcriptViewportHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}
	c.transcriptScroll = maxScroll
	c.stickToBottom = true
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) visibleTranscript() []string {
	lines := c.pruneTranscript(append([]string(nil), c.transcript...))
	if len(lines) == 0 {
		return []string{"(no messages yet)"}
	}
	return lines
}

type tuiContextSummary struct {
	sessionTitle string
	agentID      string
	parent       string
	model        string
	provider     string
	thinking     string
	status       string
	messageCount int
	turnCount    int
}

func (c *chatTUI) contextSummaryData() tuiContextSummary {
	data := tuiContextSummary{sessionTitle: c.sessionID, agentID: "agent", parent: "root", model: c.cfg.DefaultModel, provider: c.cfg.DefaultProvider, thinking: c.cfg.DefaultThinkingLevel, status: "idle"}
	session, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		return data
	}
	messages, _ := c.store.ListMessages(context.Background(), c.sessionID)
	turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
	data.messageCount = len(messages)
	data.turnCount = len(turns)
	data.sessionTitle = session.Title
	data.agentID = c.agentIDForSession(session)
	if session.ParentSessionID != "" {
		data.parent = session.ParentSessionID
	}
	state := session.State
	if v, ok := state["model"].(string); ok && v != "" {
		data.model = v
	}
	if v, ok := state["provider"].(string); ok && v != "" {
		data.provider = v
	}
	if v, ok := state["thinking_level"].(string); ok && v != "" {
		data.thinking = v
	}
	if v, ok := state["status"].(string); ok && v != "" {
		data.status = v
	}
	return data
}

func (c *chatTUI) contextSummaryLines(width int) []string {
	data := c.contextSummaryData()
	if width >= 110 {
		return c.wrapLines(fmt.Sprintf("Session: %s · Agent: @%s · Parent: %s · Model: %s · Provider: %s · Thinking: %s · Status: %s · Messages: %d · Turns: %d", data.sessionTitle, data.agentID, data.parent, data.model, data.provider, data.thinking, data.status, data.messageCount, data.turnCount), width)
	}
	if width >= 80 {
		return []string{
			fmt.Sprintf("Session: %s · Agent: @%s · Parent: %s", data.sessionTitle, data.agentID, data.parent),
			fmt.Sprintf("Model: %s · Provider: %s · Thinking: %s · Status: %s", data.model, data.provider, data.thinking, data.status),
			fmt.Sprintf("Messages: %d · Turns: %d", data.messageCount, data.turnCount),
		}
	}
	return []string{
		fmt.Sprintf("Session: %s", data.sessionTitle),
		fmt.Sprintf("Agent: @%s · Parent: %s", data.agentID, data.parent),
		fmt.Sprintf("Model: %s", data.model),
		fmt.Sprintf("Provider: %s · Thinking: %s", data.provider, data.thinking),
		fmt.Sprintf("Status: %s · Messages: %d · Turns: %d", data.status, data.messageCount, data.turnCount),
	}
}

func (c *chatTUI) contextSummary() string {
	return strings.Join(c.contextSummaryLines(9999), "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func Main() {
	dbPath := flag.String("db", ".gi-run/gi.db", "SQLite database path")
	workspace := flag.String("workspace", "/workspace", "Workspace root")
	model := flag.String("model", "", "Override default model")
	flag.Parse()
	if err := Run(*dbPath, *workspace, *model); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
