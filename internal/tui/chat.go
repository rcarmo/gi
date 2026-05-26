package tui

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/skills"
	"github.com/rcarmo/gi/internal/store"
	gitools "github.com/rcarmo/gi/internal/tools"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func initialSessionID(ctx context.Context, s *store.Store) (string, error) {
	if sessionID, err := s.ResolveMainSessionID(ctx, "agent", "gi", "default"); err == nil {
		return sessionID, nil
	}
	sessionIDs, err := s.ListSessionIDs(ctx)
	if err != nil {
		return "", fmt.Errorf("list session ids: %w", err)
	}
	if len(sessionIDs) == 0 {
		return "", nil
	}
	return sessionIDs[0], nil
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
	app               *gotui.App
	store             *store.Store
	engine            *turn.Engine
	sessionID         string
	cfg               config.RuntimeConfig
	history           []string
	histIdx           int
	running           bool
	status            string
	draft             string
	inputActive       bool
	eventCh           chan map[string]any
	topicEventCh      chan topics.Envelope
	subscribedCh      chan map[string]any
	topicUnsubscribe  func()
	input             *multilineInput
	queuedDrafts      []string
	inputRegion       *gotui.Element
	transcriptRegion  *gotui.Element
	transcriptRef     *gotui.Ref
	transcript        []string
	transcriptScroll  int
	stickToBottom     bool
	draftLineIndex    int
	outputWidth       int
	osc52Writer       io.Writer
	clipboardLookPath func(string) (string, error)
	clipboardRun      func(context.Context, string, []string, string) error
}

func (c *chatTUI) ensureInput() {
	if c.input != nil {
		return
	}
	c.input = newMultilineInput(80, "Send a message…", c.onSubmit, nil)
	c.input.onRestoreQueued = c.restoreQueuedDraft
	c.input.onComplete = c.completeInputPath
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
		c.renderToolEvent(payload["type"], payload["tool"], payload["error"], payload["reason"])
	case "runtime.hook":
		c.renderHookEvent(payload["type"], payload["hook"], payload["reason"], payload["tool"], payload["error"])
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
			if status == "queued" {
				c.resetRunningDraftState()
				c.status = "Queued"
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
		c.renderRoutingEvent(payload["type"], payload["target_agent_id"], payload["target_session"], payload["target_session_id"], payload["source_agent_id"])
	case "runtime.inbound_work":
		c.renderInboundWorkEvent(payload["type"], payload["source_kind"], payload["status"], payload["attempt_count"], payload["error"])
	case "runtime.dispatcher":
		c.renderDispatcherEvent(payload["type"], payload["worker_id"], payload["processed_count"], payload["error"])
	case "session.compaction":
		c.renderCompactionEvent(payload["messages_before"], payload["messages_after"], payload["tokens_before"])
	case "session.routing":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		c.renderRoutingEvent(payload["type"], payload["target_agent_id"], payload["target_session"], payload["target_session_id"], payload["source_agent_id"])
	case "session.steering":
		c.renderSteeringEvent(payload["type"])
	case "turn.subturn":
		c.renderSubturnEvent(payload["type"], payload["child_turn_id"], payload["status"])
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
	case "tool_finished", "tool_failed":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		c.renderToolEvent(evType, ev["tool"], ev["error"], ev["reason"])
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
		c.renderCompactionEvent(ev["messages_before"], ev["messages_after"], ev["tokens_before"])
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "routing_decision", "routing_incoming":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		c.renderRoutingEvent(evType, ev["target_agent_id"], ev["target_session"], ev["target_session_id"], ev["source_agent_id"])
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

func (c *chatTUI) renderSteeringEvent(eventTypeValue any) {
	typ, _ := eventTypeValue.(string)
	switch typ {
	case "steering_enqueued":
		c.status = "Queued follow-up"
	case "steering_injected":
		c.status = "Injected follow-up"
	case "steering_continued":
		c.status = "Continuing queued follow-up"
	}
}

func (c *chatTUI) renderSubturnEvent(eventTypeValue, childTurnValue, statusValue any) {
	typ, _ := eventTypeValue.(string)
	childTurn, _ := childTurnValue.(string)
	status, _ := statusValue.(string)
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

func (c *chatTUI) renderHookEvent(eventTypeValue, hookNameValue, reasonValue, toolNameValue, errValue any) {
	typ, _ := eventTypeValue.(string)
	hookName, _ := hookNameValue.(string)
	reason, _ := reasonValue.(string)
	toolName, _ := toolNameValue.(string)
	errText, _ := errValue.(string)
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
}

func (c *chatTUI) renderInboundWorkEvent(eventTypeValue, sourceKindValue, statusValue, attemptCountValue, errValue any) {
	typ, _ := eventTypeValue.(string)
	sourceKind, _ := sourceKindValue.(string)
	status, _ := statusValue.(string)
	attemptCount := intFromAny(attemptCountValue)
	errText, _ := errValue.(string)
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
}

func (c *chatTUI) renderDispatcherEvent(eventTypeValue, workerIDValue, processedCountValue, errValue any) {
	typ, _ := eventTypeValue.(string)
	workerID, _ := workerIDValue.(string)
	processedCount := intFromAny(processedCountValue)
	errText, _ := errValue.(string)
	line := ""
	switch typ {
	case "dispatcher_lease_acquired":
		line = "sys: inbound dispatcher lease acquired"
	case "dispatcher_lease_released":
		line = "sys: inbound dispatcher lease released"
	case "dispatcher_drain_completed":
		line = "sys: inbound dispatcher drain completed"
		if processedCount > 0 {
			line += fmt.Sprintf(" (%d processed)", processedCount)
		}
	case "dispatcher_error":
		line = "sys: inbound dispatcher error"
	}
	if line != "" {
		if workerID != "" {
			line += fmt.Sprintf(" [%s]", workerID)
		}
		if errText != "" {
			line += ": " + truncate(errText, 120)
		}
		c.appendTranscript(line)
	}
}

func (c *chatTUI) renderCompactionEvent(messagesBeforeValue, messagesAfterValue, tokensBeforeValue any) {
	before := intFromAny(messagesBeforeValue)
	after := intFromAny(messagesAfterValue)
	tokens := intFromAny(tokensBeforeValue)
	c.appendTranscript(fmt.Sprintf("sys[compact]: messages %d→%d · tokens_before=%d", before, after, tokens))
	c.status = "Compacted context"
}

func (c *chatTUI) renderToolEvent(eventTypeValue, toolNameValue, errValue, reasonValue any) {
	typ, _ := eventTypeValue.(string)
	toolName, _ := toolNameValue.(string)
	errText, _ := errValue.(string)
	reason, _ := reasonValue.(string)
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
		line := fmt.Sprintf("sys: tool failed: %s", toolName)
		if errText != "" {
			line = fmt.Sprintf("%s: %s", line, truncate(errText, 120))
		}
		c.appendTranscript(line)
		c.status = fmt.Sprintf("Tool failed: %s", toolName)
	case "tool_skipped":
		line := fmt.Sprintf("sys: tool skipped: %s", toolName)
		if reason != "" {
			line = fmt.Sprintf("%s: %s", line, truncate(reason, 120))
		}
		c.appendTranscript(line)
	}
}

func (c *chatTUI) renderRoutingEvent(eventType, targetAgentValue, targetSessionValue, targetSessionIDValue, sourceAgentValue any) {
	typ, _ := eventType.(string)
	targetAgent, _ := targetAgentValue.(string)
	targetSession, _ := targetSessionValue.(string)
	if targetSession == "" {
		targetSession, _ = targetSessionIDValue.(string)
	}
	sourceAgent, _ := sourceAgentValue.(string)
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
		gotui.OnPreemptStop(gotui.KeyCtrlL, func(ke gotui.KeyEvent) { c.cycleModel(1) }),
		gotui.OnPreemptStop(gotui.Rune('l').Alt(), func(ke gotui.KeyEvent) { c.cycleModel(-1) }),
		gotui.OnPreemptStop(gotui.KeyCtrlT, func(ke gotui.KeyEvent) { c.cycleThinking(1) }),
		gotui.OnPreemptStop(gotui.Rune('t').Alt(), func(ke gotui.KeyEvent) { c.cycleThinking(-1) }),
		gotui.OnPreemptStop(gotui.KeyUp, func(ke gotui.KeyEvent) {
			c.recallHistory(-1)
		}),
		gotui.OnPreemptStop(gotui.KeyDown, func(ke gotui.KeyEvent) {
			c.recallHistory(1)
		}),
	}
}

func (c *chatTUI) cycleModel(delta int) {
	if len(c.cfg.EnabledModels) == 0 {
		c.appendTranscript("sys: no enabled models configured; use /scoped-models add <model>")
		return
	}
	idx := -1
	for i, model := range c.cfg.EnabledModels {
		if model == c.cfg.DefaultModel {
			idx = i
			break
		}
	}
	if idx < 0 {
		idx = 0
	} else {
		idx = (idx + delta) % len(c.cfg.EnabledModels)
		if idx < 0 {
			idx += len(c.cfg.EnabledModels)
		}
	}
	lines := c.modelCommand([]string{"/model", c.cfg.EnabledModels[idx]})
	c.appendTranscript(lines...)
}

func (c *chatTUI) cycleThinking(delta int) {
	levels := []string{"low", "medium", "high"}
	current := strings.ToLower(strings.TrimSpace(c.cfg.DefaultThinkingLevel))
	idx := 0
	for i, level := range levels {
		if level == current {
			idx = i
			break
		}
	}
	idx = (idx + delta) % len(levels)
	if idx < 0 {
		idx += len(levels)
	}
	c.appendTranscript(c.thinkingCommand([]string{"/thinking", levels[idx]})...)
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

func (c *chatTUI) restoreQueuedDraft() {
	if len(c.queuedDrafts) == 0 || c.input == nil {
		return
	}
	idx := len(c.queuedDrafts) - 1
	draft := c.queuedDrafts[idx]
	c.queuedDrafts = c.queuedDrafts[:idx]
	c.input.SetText(draft)
	c.draft = draft
	c.status = "Restored queued draft"
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) onSubmit(text string) {
	c.submitWithMetadata(text, nil)
}

func (c *chatTUI) submitWithMetadata(text string, metadata map[string]any) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	c.history = append(c.history, text)
	c.histIdx = -1
	c.input.SetText("")
	if strings.HasPrefix(text, "/skill:") {
		c.appendTranscript(c.skillCommandLines(text)...)
		return
	}
	if strings.HasPrefix(text, "/") {
		c.handleCommand(text)
		return
	}
	if strings.HasPrefix(text, "!!") {
		c.appendTranscript(c.localShellShortcutLines(strings.TrimSpace(strings.TrimPrefix(text, "!!")))...)
		return
	}
	if strings.HasPrefix(text, "!") {
		cmd := strings.TrimSpace(strings.TrimPrefix(text, "!"))
		if cmd == "" {
			c.appendTranscript("sys: usage: !command sends a shell request to the model; !!command runs it locally")
			return
		}
		text = fmt.Sprintf("Run this shell command and summarize the result: %s", cmd)
	}
	if c.running {
		c.queuedDrafts = append(c.queuedDrafts, text)
		c.appendTranscript(fmt.Sprintf("you [queued]: %s", text))
		c.status = "Queued follow-up"
		c.stickToBottom = true
		c.scrollTranscriptToBottom()
		if c.app != nil {
			c.app.MarkDirty()
		}
		go func() {
			_, err := c.engine.SubmitPromptRouted(context.Background(), turn.RunInput{SessionID: c.sessionID, Prompt: text, Intent: "prompt", Model: c.cfg.DefaultModel})
			if err != nil {
				if c.app != nil {
					c.app.QueueUpdate(func() {
						c.appendTranscript(fmt.Sprintf("error: queue follow-up: %v", err))
						c.app.MarkDirty()
					})
				} else {
					c.appendTranscript(fmt.Sprintf("error: queue follow-up: %v", err))
				}
			}
		}()
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
			Metadata:  metadata,
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
	case "/commands", "/palette":
		query := ""
		if len(fields) > 1 {
			query = strings.Join(fields[1:], " ")
		}
		c.appendTranscript(c.commandPaletteLines(query)...)
	case "/session":
		c.appendTranscript(c.sessionLines()...)
	case "/new":
		c.appendTranscript(c.newSessionLines()...)
	case "/name":
		c.appendTranscript(c.nameSessionLines(text, fields)...)
	case "/resume":
		c.appendTranscript(c.resumeLines(fields)...)
	case "/clone":
		c.appendTranscript(c.cloneSessionLines(fields)...)
	case "/copy":
		c.appendTranscript(c.copyLastAssistantLines(fields[1:]...)...)
	case "/attach":
		c.appendTranscript(c.attachCommand(text, fields)...)
	case "/reload":
		c.appendTranscript(c.reloadLines()...)
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
	case "/scoped-models":
		c.transcript = append(c.transcript, c.scopedModelsCommand(fields)...)
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
		targetSessionID, err := c.engine.ResolveOrCreatePeerSessionID(context.Background(), c.sessionID, target)
		if err != nil {
			c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
			break
		}
		c.switchSession(targetSessionID)
		c.transcript = append(c.transcript, fmt.Sprintf("sys: switched to @%s (%s)", target, targetSessionID))
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
		if lines, handled := c.extensionCommandLines(text, fields); handled {
			c.appendTranscript(lines...)
		} else {
			c.appendTranscript("sys: commands: /help, /commands [query], /session, /new, /name <name>, /resume [index|session_id], /clone [@agentN], /copy [--osc52|--native|--auto|--fallback], /attach <path> [prompt], /reload, /tools [query|active|activate|reset], /skills [query], /skill:name [args], /model [name], /scoped-models [add|remove|set], /thinking [level], /compact, /scrollback [n], /settings, /approvals, /cancel, /agents, /tree, /plugins, /fork [@agentN], /switch @agent|session_id, /send @agent message, /where, !cmd, !!cmd")
		}
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

func (c *chatTUI) extensionCommandLines(text string, fields []string) ([]string, bool) {
	if c.engine == nil || len(fields) == 0 {
		return nil, false
	}
	name := strings.TrimPrefix(fields[0], "/")
	rawArgs := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	res, handled, err := c.engine.InvokeExtensionCommand(context.Background(), name, rawArgs, c.sessionID, "")
	if !handled {
		return nil, false
	}
	if err != nil {
		return []string{fmt.Sprintf("error: extension command /%s: %v", name, err)}, true
	}
	switch strings.ToLower(strings.TrimSpace(res.Type)) {
	case "noop":
		if strings.TrimSpace(res.Status) != "" {
			c.status = res.Status
		}
		return []string{}, true
	case "error":
		if res.Error != "" {
			return []string{fmt.Sprintf("error: extension command /%s: %s", name, res.Error)}, true
		}
		return append([]string{fmt.Sprintf("error: extension command /%s", name)}, res.Lines...), true
	case "submit":
		prompt := strings.TrimSpace(res.Prompt)
		if prompt == "" {
			return []string{fmt.Sprintf("error: extension command /%s returned empty submit prompt", name)}, true
		}
		c.onSubmit(prompt)
		return []string{fmt.Sprintf("sys: extension command /%s submitted prompt", name)}, true
	default:
		if len(res.Lines) == 0 && strings.TrimSpace(res.Status) != "" {
			return []string{"sys: " + res.Status}, true
		}
		return res.Lines, true
	}
}

func (c *chatTUI) commandPaletteLines(query string) []string {
	commands := []struct{ name, hint string }{
		{"/help", "show grouped help"},
		{"/commands [query]", "filter command palette textually"},
		{"/session", "show current session details"},
		{"/new", "create and switch to a new main session"},
		{"/name <name>", "rename current session"},
		{"/resume [index|session_id]", "list or switch recent sessions"},
		{"/clone [@agentN]", "clone active branch/session"},
		{"/copy [--osc52|--native|--auto|--fallback]", "copy last assistant message with opt-in target"},
		{"/attach <path> [prompt]", "attach local media and optionally submit a prompt"},
		{"/reload", "refresh config and discovery safely"},
		{"/tools [query|active|activate|reset]", "inspect or change active tools"},
		{"/skills [query]", "list discovered skills"},
		{"/skill:name [args]", "load a discovered SKILL.md"},
		{"/model [name|index]", "list or select model"},
		{"/scoped-models [list|add|remove|set]", "manage enabled models"},
		{"/thinking [level]", "show or set thinking level"},
		{"/compact", "request context compaction"},
		{"/scrollback [n]", "show or set transcript scrollback limit"},
		{"/settings", "show grouped runtime settings"},
		{"/cancel", "cancel latest active/queued turn"},
		{"/agents", "list configured agents"},
		{"/tree", "show session tree"},
		{"/plugins", "show loaded extensions"},
		{"/fork [@agentN]", "create peer/fork session"},
		{"/switch @agent|session_id", "switch active session"},
		{"/send @agent message", "send peer message"},
		{"/where", "show context summary"},
		{"!cmd", "ask model to run/summarize shell command"},
		{"!!cmd", "run local shell command"},
	}
	q := strings.ToLower(strings.TrimSpace(query))
	lines := []string{"commands: palette"}
	for _, cmd := range commands {
		text := cmd.name + " " + cmd.hint
		if q != "" && !strings.Contains(strings.ToLower(text), q) {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %-36s %s", cmd.name, cmd.hint))
	}
	if c.engine != nil {
		for _, cmd := range c.engine.ExtensionCommandInfos() {
			name := "/" + cmd.Name
			hint := strings.TrimSpace(cmd.Description)
			if hint == "" {
				hint = "extension command"
			}
			text := name + " " + hint + " " + cmd.Usage
			if q != "" && !strings.Contains(strings.ToLower(text), q) {
				continue
			}
			lines = append(lines, fmt.Sprintf("- %-36s %s", name, hint))
		}
	}
	if len(lines) == 1 {
		return []string{fmt.Sprintf("commands: no matches for %q", query)}
	}
	return lines
}

func (c *chatTUI) helpLines() []string {
	return []string{
		"help",
		"enter send · shift-enter newline · esc blur · ctrl-d exit",
		"/commands  all commands",
		"/model     choose model · ctrl-l cycles",
		"/session   details for this chat",
		"/where     compact context",
		"/attach    add media",
		"!cmd       ask model about shell · !!cmd run locally",
	}
}

func (c *chatTUI) completeInputPath(text string, cursor int) (string, int, bool) {
	runes := []rune(text)
	if cursor < 0 || cursor > len(runes) {
		cursor = len(runes)
	}
	start := cursor
	for start > 0 && !isWordSpace(runes[start-1]) {
		start--
	}
	prefix := string(runes[start:cursor])
	if prefix == "" {
		return text, cursor, false
	}
	fileRef := strings.HasPrefix(prefix, "@")
	searchPrefix := prefix
	if fileRef {
		searchPrefix = strings.TrimPrefix(prefix, "@")
		if searchPrefix == "" {
			searchPrefix = "*"
		}
	}
	root := strings.TrimSpace(c.cfg.WorkspaceRoot)
	if root == "" {
		root = "."
	}
	pattern := searchPrefix + "*"
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(root, pattern)
	}
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return text, cursor, false
	}
	sort.Strings(matches)
	match := matches[0]
	if info, err := os.Stat(match); err == nil && info.IsDir() {
		match += string(os.PathSeparator)
	}
	insert := match
	if !filepath.IsAbs(searchPrefix) {
		if rel, err := filepath.Rel(root, match); err == nil && !strings.HasPrefix(rel, "..") {
			insert = rel
			if strings.HasSuffix(match, string(os.PathSeparator)) && !strings.HasSuffix(insert, string(os.PathSeparator)) {
				insert += string(os.PathSeparator)
			}
		}
	}
	if fileRef {
		insert = "@" + insert
	}
	out := string(runes[:start]) + insert + string(runes[cursor:])
	return out, start + utf8.RuneCountInString(insert), true
}

func (c *chatTUI) localShellShortcutLines(command string) []string {
	command = strings.TrimSpace(command)
	if command == "" {
		return []string{"sys: usage: !!command runs a local shell command; !command sends a shell request to the model"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	if root := strings.TrimSpace(c.cfg.WorkspaceRoot); root != "" {
		cmd.Dir = root
	}
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if text == "" {
		text = "(no output)"
	}
	text = truncate(text, 2000)
	lines := []string{fmt.Sprintf("local$ %s", command)}
	if err != nil {
		lines = append(lines, fmt.Sprintf("error: %v", err))
	}
	for _, line := range strings.Split(text, "\n") {
		lines = append(lines, "│ "+line)
	}
	return lines
}

func (c *chatTUI) reloadLines() []string {
	workspace := c.cfg.WorkspaceRoot
	if strings.TrimSpace(workspace) == "" {
		workspace = "/workspace"
	}
	reloaded := config.Load(workspace)
	c.cfg = reloaded
	if c.input != nil {
		c.input.placeholder = "Send a message…"
	}
	lines := []string{
		"reload: config refreshed",
		fmt.Sprintf("- workspace=%s", c.cfg.WorkspaceRoot),
		fmt.Sprintf("- provider=%s model=%s thinking=%s", c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel),
		fmt.Sprintf("- discovery: skills=%d tools=%d", len(c.cfg.Discovery.Skills), len(c.cfg.Discovery.Tools)),
	}
	discoveredExtensions := gitools.DiscoverExtensionScripts(c.cfg.WorkspaceRoot)
	mountedExtensions := 0
	if c.engine != nil {
		mountedExtensions = len(c.engine.ExtensionInfos())
	}
	lines = append(lines, fmt.Sprintf("- extensions: discovered=%d mounted=%d", len(discoveredExtensions), mountedExtensions))
	if len(discoveredExtensions) != mountedExtensions {
		lines = append(lines, "- note: extension set changed; safe handler unload/reload is deferred, restart to remount extensions")
	} else {
		lines = append(lines, "- note: extension handlers remain mounted; /reload refreshes config and skill/tool discovery safely")
	}
	return lines
}

func (c *chatTUI) copyLastAssistantLines(args ...string) []string {
	messages, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		return []string{fmt.Sprintf("error: copy last assistant message: %v", err)}
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "assistant" || strings.TrimSpace(messages[i].Content) == "" {
			continue
		}
		content := strings.TrimSpace(messages[i].Content)
		mode, persist, usage := c.copyModeFromArgs(args)
		if usage != "" {
			return []string{usage}
		}
		if persist {
			if err := config.PersistClipboardMode(c.cfg.WorkspaceRoot, mode); err != nil {
				return []string{fmt.Sprintf("warn: failed to persist clipboard mode: %v", err)}
			}
			c.cfg.TUIClipboardMode = mode
		}
		if mode == "osc52" {
			if err := c.writeOSC52(content); err != nil {
				return c.copyFallbackLines(content, fmt.Sprintf("OSC 52 failed: %v", err))
			}
			return []string{fmt.Sprintf("copy: sent %d chars using OSC 52", len(content))}
		}
		if mode == "native" || mode == "auto" {
			if err := c.copyNative(content); err == nil {
				return []string{fmt.Sprintf("copy: sent %d chars using native clipboard helper", len(content))}
			} else if mode == "native" {
				return c.copyFallbackLines(content, fmt.Sprintf("native clipboard failed: %v", err))
			}
		}
		return c.copyFallbackLines(content, "clipboard unavailable")
	}
	return []string{"copy: no assistant message found"}
}

func (c *chatTUI) copyModeFromArgs(args []string) (mode string, persist bool, usage string) {
	mode = strings.TrimSpace(c.cfg.TUIClipboardMode)
	if mode == "" {
		mode = "off"
	}
	for i := 0; i < len(args); i++ {
		switch strings.ToLower(strings.TrimSpace(args[i])) {
		case "--osc52":
			mode = "osc52"
		case "--native":
			mode = "native"
		case "--auto":
			mode = "auto"
		case "--fallback", "--off":
			mode = "off"
		case "--mode":
			if i+1 >= len(args) {
				return "", false, "copy: usage /copy [--osc52|--native|--auto|--fallback|--mode <off|osc52|native|auto>] [--persist]"
			}
			i++
			mode = strings.ToLower(strings.TrimSpace(args[i]))
		case "--persist":
			persist = true
		default:
			return "", false, "copy: usage /copy [--osc52|--native|--auto|--fallback|--mode <off|osc52|native|auto>] [--persist]"
		}
	}
	switch mode {
	case "osc52", "native", "auto":
		return mode, persist, ""
	default:
		return "off", persist, ""
	}
}

func (c *chatTUI) copyFallbackLines(content, reason string) []string {
	lines := []string{fmt.Sprintf("copy: %s; last assistant message follows (%d chars)", reason, len(content))}
	lines = append(lines, prefixMultiline("copy", content)...)
	return lines
}

const osc52PayloadLimit = 64 * 1024

func osc52Sequence(content string) (string, error) {
	if len(content) > osc52PayloadLimit {
		return "", fmt.Errorf("payload too large (%d > %d bytes)", len(content), osc52PayloadLimit)
	}
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(content)) + "\x07", nil
}

func (c *chatTUI) writeOSC52(content string) error {
	seq, err := osc52Sequence(content)
	if err != nil {
		return err
	}
	w := c.osc52Writer
	if w == nil {
		w = os.Stdout
	}
	_, err = io.WriteString(w, seq)
	return err
}

func (c *chatTUI) copyNative(content string) error {
	helper, ok := selectNativeClipboardHelper(runtime.GOOS, os.Getenv("WAYLAND_DISPLAY") != "", os.Getenv("DISPLAY") != "", c.lookupClipboardHelper)
	if !ok {
		return fmt.Errorf("no native clipboard helper found")
	}
	runner := c.clipboardRun
	if runner == nil {
		runner = runClipboardHelper
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return runner(ctx, helper.name, helper.args, content)
}

type clipboardHelper struct {
	name string
	args []string
}

func (c *chatTUI) lookupClipboardHelper(name string) (string, error) {
	if c.clipboardLookPath != nil {
		return c.clipboardLookPath(name)
	}
	return exec.LookPath(name)
}

func selectNativeClipboardHelper(goos string, wayland, x11 bool, lookPath func(string) (string, error)) (clipboardHelper, bool) {
	candidates := []clipboardHelper{}
	switch goos {
	case "darwin":
		candidates = append(candidates, clipboardHelper{name: "pbcopy"})
	case "windows":
		candidates = append(candidates, clipboardHelper{name: "clip.exe"})
	default:
		if wayland {
			candidates = append(candidates, clipboardHelper{name: "wl-copy"})
		}
		if x11 {
			candidates = append(candidates, clipboardHelper{name: "xclip", args: []string{"-selection", "clipboard"}}, clipboardHelper{name: "xsel", args: []string{"--clipboard", "--input"}})
		}
		candidates = append(candidates, clipboardHelper{name: "clip.exe"})
	}
	for _, candidate := range candidates {
		if _, err := lookPath(candidate.name); err == nil {
			return candidate, true
		}
	}
	return clipboardHelper{}, false
}

func runClipboardHelper(ctx context.Context, name string, args []string, content string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	_, writeErr := io.WriteString(stdin, content)
	closeErr := stdin.Close()
	waitErr := cmd.Wait()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return waitErr
}

func (c *chatTUI) cloneSessionLines(fields []string) []string {
	target := ""
	if len(fields) > 1 {
		target = strings.TrimPrefix(strings.TrimSpace(fields[1]), "@")
	}
	if target == "" {
		target = c.nextForkAgentID()
	}
	newID := store.NowID("session")
	cloned, err := c.store.CloneSession(context.Background(), c.sessionID, newID, "@"+target, target)
	if err != nil {
		return []string{fmt.Sprintf("error: clone session: %v", err)}
	}
	c.switchSession(cloned.ID)
	return []string{fmt.Sprintf("sys: cloned to @%s (%s)", c.agentIDForSession(cloned), cloned.ID)}
}

func (c *chatTUI) newSessionLines() []string {
	id := store.NowID("session")
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", id)
	state := map[string]any{"status": "idle", "queue_count": 0, "model": c.cfg.DefaultModel, "provider": c.cfg.DefaultProvider, "thinking_level": c.cfg.DefaultThinkingLevel}
	sess, _, err := c.store.ResolveOrCreateMainSessionFromAllocation(context.Background(), store.ResolveOrCreateSessionFromAllocationInput{ID: id, Title: "@agent", State: state, Allocation: alloc})
	if err != nil {
		return []string{fmt.Sprintf("error: create session: %v", err)}
	}
	c.switchSession(sess.ID)
	return []string{fmt.Sprintf("sys: new session @%s (%s)", c.agentIDForSession(sess), sess.ID)}
}

func (c *chatTUI) nameSessionLines(text string, fields []string) []string {
	if len(fields) < 2 {
		return []string{"sys: usage /name <name>"}
	}
	name := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	if name == "" {
		return []string{"sys: usage /name <name>"}
	}
	if err := c.store.UpdateSessionTitle(context.Background(), c.sessionID, name); err != nil {
		return []string{fmt.Sprintf("error: rename session: %v", err)}
	}
	return []string{fmt.Sprintf("sys: session renamed to %s", name)}
}

func (c *chatTUI) resumeLines(fields []string) []string {
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		return []string{fmt.Sprintf("error: list sessions: %v", err)}
	}
	if len(fields) > 1 {
		arg := strings.TrimSpace(fields[1])
		var target *store.Session
		if idx, err := strconv.Atoi(arg); err == nil {
			if idx < 1 || idx > len(sessions) {
				return []string{fmt.Sprintf("error: resume index out of range: %d", idx)}
			}
			target = &sessions[idx-1]
		} else {
			resolved, err := c.resolveSessionRef(arg)
			if err != nil {
				return []string{fmt.Sprintf("error: %v", err)}
			}
			target = resolved
		}
		c.switchSession(target.ID)
		return []string{fmt.Sprintf("sys: resumed @%s (%s)", c.agentIDForSession(target), target.ID)}
	}
	if len(sessions) == 0 {
		return []string{"resume: no sessions"}
	}
	limit := len(sessions)
	if limit > 10 {
		limit = 10
	}
	lines := []string{"resume: recent sessions"}
	for i := 0; i < limit; i++ {
		sess := sessions[i]
		messages, _ := c.store.ListMessages(context.Background(), sess.ID)
		turns, _ := c.store.ListTurns(context.Background(), sess.ID)
		status, _ := sess.State["status"].(string)
		if status == "" {
			status = "idle"
		}
		if c.compactOutput() {
			lines = append(lines, fmt.Sprintf("%d. @%s %s [%s] · %s · m=%d t=%d", i+1, c.agentIDForSession(&sess), compactText(sess.Title, 18), compactID(sess.ID), status, len(messages), len(turns)))
		} else {
			lines = append(lines, fmt.Sprintf("%d. @%s %s (%s) · %s · messages=%d turns=%d", i+1, c.agentIDForSession(&sess), sess.Title, sess.ID, status, len(messages), len(turns)))
		}
	}
	lines = append(lines, "resume: use /resume <index|session_id>")
	return lines
}

func (c *chatTUI) sessionLines() []string {
	sess, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	messages, _ := c.store.ListMessages(context.Background(), c.sessionID)
	turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
	queuedTurns, _ := c.store.CountQueuedTurns(context.Background(), c.sessionID)
	steeringDepth, _ := c.store.SteeringQueueLength(context.Background(), c.sessionID)
	activeTurnID, claimToken, _ := c.store.GetSessionActiveTurn(context.Background(), c.sessionID)
	model := c.cfg.DefaultModel
	provider := c.cfg.DefaultProvider
	thinking := c.cfg.DefaultThinkingLevel
	status := "idle"
	queueCount := queuedTurns
	if v, ok := sess.State["model"].(string); ok && v != "" {
		model = v
	}
	if v, ok := sess.State["provider"].(string); ok && v != "" {
		provider = v
	}
	if v, ok := sess.State["thinking_level"].(string); ok && v != "" {
		thinking = v
	}
	if v, ok := sess.State["status"].(string); ok && v != "" {
		status = v
	}
	if v := intFromAny(sess.State["queue_count"]); v > queueCount {
		queueCount = v
	}
	parent := sess.ParentSessionID
	if parent == "" {
		parent = "root"
	}
	active := "none"
	if activeTurnID != "" {
		active = activeTurnID
		if claimToken != "" {
			active += " (claimed)"
		}
	}
	return []string{
		"session: current",
		fmt.Sprintf("- id=%s", sess.ID),
		fmt.Sprintf("- title=%s agent=@%s parent=%s", sess.Title, c.agentIDForSession(sess), parent),
		fmt.Sprintf("- messages=%d turns=%d queued_turns=%d steering=%d", len(messages), len(turns), queueCount, steeringDepth),
		fmt.Sprintf("- status=%s running=%v active_turn=%s", status, c.running, active),
		fmt.Sprintf("- model=%s provider=%s thinking=%s", model, provider, thinking),
	}
}

func (c *chatTUI) modelCommand(fields []string) []string {
	if len(fields) == 1 {
		return c.modelListLines()
	}
	model := strings.TrimSpace(strings.Join(fields[1:], " "))
	if model == "" {
		return []string{"sys: usage /model <model|index>"}
	}
	if idx, err := strconv.Atoi(model); err == nil && idx > 0 && idx <= len(c.cfg.EnabledModels) {
		model = c.cfg.EnabledModels[idx-1]
	}
	c.cfg.DefaultModel = model
	if strings.Contains(model, "/") {
		c.cfg.DefaultProvider = strings.SplitN(model, "/", 2)[0]
	}
	lines := []string{fmt.Sprintf("model: %s", model)}
	if err := c.store.TouchSessionState(context.Background(), c.sessionID, map[string]any{"model": model}); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist model in session state: %v", err))
	}
	if err := config.PersistModelSelection(c.cfg.WorkspaceRoot, c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel, c.cfg.EnabledModels); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist model selection: %v", err))
	}
	return lines
}

func (c *chatTUI) modelListLines() []string {
	current := strings.TrimSpace(c.cfg.DefaultModel)
	if current == "" {
		current = "(none selected)"
	}
	provider := strings.TrimSpace(c.cfg.DefaultProvider)
	if provider == "" {
		provider = "(inferred from model when possible)"
	}
	thinking := strings.TrimSpace(c.cfg.DefaultThinkingLevel)
	if thinking == "" {
		thinking = "low"
	}
	modelWidth := 52
	if c.compactOutput() {
		modelWidth = 34
	}
	lines := []string{fmt.Sprintf("model %s · %s", compactMaybe(current, c.compactOutput(), modelWidth), thinking)}
	if provider != "" && provider != "(inferred from model when possible)" {
		lines[0] += fmt.Sprintf(" · %s", compactMaybe(provider, c.compactOutput(), 18))
	}
	if len(c.cfg.EnabledModels) == 0 {
		return append(lines, "no enabled models · /model <provider/model>")
	}
	for i, model := range c.cfg.EnabledModels {
		marker := " "
		if model == c.cfg.DefaultModel {
			marker = "›"
		}
		lines = append(lines, fmt.Sprintf("%s %d  %s", marker, i+1, compactMaybe(model, c.compactOutput(), modelWidth)))
	}
	lines = append(lines, "/model <n> to switch · ctrl-l cycles")
	return lines
}

func (c *chatTUI) scopedModelsCommand(fields []string) []string {
	if len(fields) == 1 || strings.EqualFold(fields[1], "list") {
		lines := []string{"scoped-models: enabled models:"}
		if len(c.cfg.EnabledModels) == 0 {
			return append(lines, "- none configured", "- usage: /scoped-models add <model> | remove <model|index> | set <model> [model...]")
		}
		for i, model := range c.cfg.EnabledModels {
			marker := " "
			if model == c.cfg.DefaultModel {
				marker = "*"
			}
			lines = append(lines, fmt.Sprintf("%s%d. %s", marker, i+1, model))
		}
		lines = append(lines, "- usage: /scoped-models add <model> | remove <model|index> | set <model> [model...]")
		return lines
	}
	action := strings.ToLower(strings.TrimSpace(fields[1]))
	args := fields[2:]
	models := append([]string(nil), c.cfg.EnabledModels...)
	switch action {
	case "add":
		if len(args) == 0 {
			return []string{"sys: usage /scoped-models add <model> [model...]"}
		}
		for _, model := range args {
			model = strings.TrimSpace(model)
			if model != "" && !containsString(models, model) {
				models = append(models, model)
			}
		}
	case "remove", "rm":
		if len(args) != 1 {
			return []string{"sys: usage /scoped-models remove <model|index>"}
		}
		needle := strings.TrimSpace(args[0])
		removed := false
		if idx, err := strconv.Atoi(needle); err == nil && idx > 0 && idx <= len(models) {
			models = append(models[:idx-1], models[idx:]...)
			removed = true
		} else {
			filtered := models[:0]
			for _, model := range models {
				if model == needle {
					removed = true
					continue
				}
				filtered = append(filtered, model)
			}
			models = filtered
		}
		if !removed {
			return []string{fmt.Sprintf("sys: scoped model not found: %s", needle)}
		}
	case "set":
		if len(args) == 0 {
			return []string{"sys: usage /scoped-models set <model> [model...]"}
		}
		models = nil
		for _, model := range args {
			model = strings.TrimSpace(model)
			if model != "" && !containsString(models, model) {
				models = append(models, model)
			}
		}
	default:
		return []string{"sys: usage /scoped-models [list|add|remove|set]"}
	}
	c.cfg.EnabledModels = models
	if c.cfg.DefaultModel != "" && !containsString(models, c.cfg.DefaultModel) && len(models) > 0 {
		c.cfg.DefaultModel = models[0]
	}
	lines := []string{fmt.Sprintf("sys: scoped models updated (%d enabled)", len(models))}
	if err := config.PersistModelSelection(c.cfg.WorkspaceRoot, c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel, c.cfg.EnabledModels); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist scoped models: %v", err))
	}
	return append(lines, c.modelListLines()...)
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
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
	activeTools := "(none)"
	if c.engine != nil {
		if names := c.engine.ActiveTools(); len(names) > 0 {
			activeTools = strings.Join(names, ", ")
		}
	}
	enabledModels := "(none)"
	if len(c.cfg.EnabledModels) > 0 {
		enabledModels = strings.Join(c.cfg.EnabledModels, ", ")
	}
	return []string{
		"settings: runtime",
		fmt.Sprintf("- workspace: %s", compactMaybe(c.cfg.WorkspaceRoot, c.compactOutput(), 48)),
		fmt.Sprintf("- max_iterations: %d", c.cfg.MaxIterations),
		fmt.Sprintf("- active_tools: %s", activeTools),
		"settings: model",
		fmt.Sprintf("- provider: %s", c.cfg.DefaultProvider),
		fmt.Sprintf("- model: %s", c.cfg.DefaultModel),
		fmt.Sprintf("- thinking: %s", c.cfg.DefaultThinkingLevel),
		fmt.Sprintf("- enabled_models: %s", enabledModels),
		"settings: editor",
		fmt.Sprintf("- scrollback_limit: %d", c.currentScrollbackLimit()),
		fmt.Sprintf("- clipboard_mode: %s", c.cfg.TUIClipboardMode),
		"- shortcuts: Ctrl+L/Alt+L model cycle, Ctrl+T/Alt+T thinking cycle, Tab path completion, @path completion",
		"settings: session",
		fmt.Sprintf("- session_id: %s", c.sessionID),
		fmt.Sprintf("- running: %v", c.running),
		"settings: discovery",
		fmt.Sprintf("- tools_discovery: %d", len(c.cfg.Discovery.Tools)),
		fmt.Sprintf("- skills_discovery: %d", len(c.cfg.Discovery.Skills)),
		"settings: compaction",
		fmt.Sprintf("- enabled: %v", c.cfg.Compaction.Enabled),
		fmt.Sprintf("- threshold_tokens: %d keep_recent_tokens: %d reserve_tokens: %d", c.cfg.Compaction.ThresholdTokens, c.cfg.Compaction.KeepRecentTokens, c.cfg.Compaction.ReserveTokens),
		"settings: peering",
		fmt.Sprintf("- enabled: %v", c.cfg.Peering.Enabled),
		fmt.Sprintf("- hostname: %s auth_key_env: %s auth_key_keychain: %s", c.cfg.Peering.Hostname, c.cfg.Peering.AuthKeyEnv, c.cfg.Peering.AuthKeyKeychain),
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

func (c *chatTUI) skillCommandLines(text string) []string {
	body := strings.TrimSpace(strings.TrimPrefix(text, "/skill:"))
	if body == "" {
		return []string{"sys: usage /skill:<name> [args]"}
	}
	name := body
	args := ""
	if fields := strings.Fields(body); len(fields) > 0 {
		name = fields[0]
		args = strings.TrimSpace(strings.TrimPrefix(body, name))
	}
	out, err := skills.ExecuteMeta(c.cfg.WorkspaceRoot, map[string]any{"name": name})
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	lines := []string{fmt.Sprintf("skill:%s loaded", name)}
	if args != "" {
		lines = append(lines, fmt.Sprintf("- args: %s", args))
	}
	lines = append(lines, prefixMultiline("skill", out)...)
	return lines
}

func (c *chatTUI) skillLines(query string) []string {
	args := map[string]any{}
	if strings.TrimSpace(query) != "" {
		args["query"] = strings.TrimSpace(query)
	}
	out, err := skills.ExecuteMeta(c.cfg.WorkspaceRoot, args)
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
	commands := c.engine.ExtensionCommandInfos()
	conflicts := c.engine.ExtensionCommandConflicts()
	lines = append(lines, fmt.Sprintf("plugins: commands: %d", len(commands)))
	for _, cmd := range commands {
		usage := strings.TrimSpace(cmd.Usage)
		if usage == "" {
			usage = "/" + cmd.Name
		}
		lines = append(lines, fmt.Sprintf("- /%s from %s (%s) · %s", cmd.Name, gitools.FirstNonEmpty(cmd.Source, "extension"), gitools.FirstNonEmpty(cmd.Engine, "unknown"), usage))
	}
	for _, cmd := range conflicts {
		lines = append(lines, fmt.Sprintf("- conflict /%s from %s (%s)", cmd.Name, gitools.FirstNonEmpty(cmd.Source, "extension"), gitools.FirstNonEmpty(cmd.Engine, "unknown")))
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

	c.ensureInput()
	c.input.width = contentWidth
	inputHeight := c.input.Render(app).HeightForWidth(contentWidth)
	if inputHeight < 1 {
		inputHeight = 1
	}
	footerLines := c.wrapLines(c.footerTextForWidth(contentWidth), contentWidth)
	reservedHeight := (padding * 2) + 1 + inputHeight + len(footerLines)
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

	if len(footerLines) > 0 {
		root.AddChild(c.renderLineBlock(footerLines, gotui.NewStyle().Dim()))
	}

	return root
}

func (c *chatTUI) statusLine() string {
	state := "idle"
	icon := "●"
	status := strings.TrimSpace(c.status)
	lower := strings.ToLower(status)
	switch {
	case c.running || strings.Contains(lower, "thinking") || strings.Contains(lower, "running"):
		state, icon = "running", "▶"
	case strings.Contains(lower, "queued"):
		state, icon = "queued", "◷"
	case strings.Contains(lower, "tool"):
		state, icon = "tool", "◆"
	case strings.Contains(lower, "hook"):
		state, icon = "hook", "◇"
	case strings.Contains(lower, "failed") || strings.Contains(lower, "error"):
		state, icon = "error", "!"
	case strings.Contains(lower, "compact"):
		state, icon = "compact", "◌"
	}
	data := c.contextSummaryData()
	model := data.model
	if model == "" {
		model = c.cfg.DefaultModel
	}
	base := fmt.Sprintf("%s %s · %s · %s · m%d/t%d", icon, c.cfg.AssistantName, model, data.thinking, data.messageCount, data.turnCount)
	if data.queuedTurns > 0 || data.steeringDepth > 0 {
		base += fmt.Sprintf(" · q%d/s%d", data.queuedTurns, data.steeringDepth)
	}
	if state != "idle" || (status != "" && !strings.Contains(status, c.cfg.DefaultModel)) {
		base += fmt.Sprintf(" · %s", strings.TrimSpace(status))
	}
	return base
}

func (c *chatTUI) footerTextForWidth(width int) string {
	left := strings.TrimSpace(c.cfg.WorkspaceRoot)
	if left == "" {
		left = "."
	} else if base := filepath.Base(left); base != "." && base != string(filepath.Separator) {
		left = base
	}
	model := strings.TrimSpace(c.cfg.DefaultModel)
	if model == "" {
		model = "model unset"
	}
	thinking := strings.TrimSpace(c.cfg.DefaultThinkingLevel)
	if thinking != "" {
		model += " • " + thinking
	}
	if width <= 0 {
		return left + "    " + model
	}
	if len(left)+len(model)+4 >= width {
		leftWidth := width / 3
		if leftWidth < 12 {
			leftWidth = 12
		}
		modelWidth := width / 2
		if modelWidth < 12 {
			modelWidth = 12
		}
		return compactMaybe(left, true, leftWidth) + "  " + compactMaybe(model, true, modelWidth)
	}
	return left + strings.Repeat(" ", width-len(left)-len(model)) + model
}

func (c *chatTUI) footerText() string {
	return c.footerTextForWidth(c.currentContentWidth())
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

func (c *chatTUI) compactOutput() bool { return c.currentContentWidth() < 72 }

func compactID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 16 {
		return id
	}
	return id[:8] + "…" + id[len(id)-5:]
}

func compactText(text string, max int) string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if text == "" {
		return "(untitled)"
	}
	return truncate(text, max)
}

func compactMaybe(text string, compact bool, max int) string {
	if !compact {
		return text
	}
	return compactText(text, max)
}

func (c *chatTUI) currentContentWidth() int {
	if c.outputWidth > 0 {
		return c.outputWidth
	}
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
		return fmt.Sprintf("tool[%s/%s]: %s", toolName, status, foldedContentSummary(m.Content, 160))
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

func foldedContentSummary(content string, maxLen int) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "(empty)"
	}
	lines := strings.Split(trimmed, "\n")
	first := strings.TrimSpace(lines[0])
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			first = strings.TrimSpace(line)
			break
		}
	}
	first = strings.Join(strings.Fields(first), " ")
	if len(lines) <= 1 {
		return truncate(first, maxLen)
	}
	summary := fmt.Sprintf("%d lines · %s", len(lines), first)
	return truncate(summary, maxLen)
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
	sessionTitle  string
	agentID       string
	parent        string
	model         string
	provider      string
	thinking      string
	status        string
	messageCount  int
	turnCount     int
	queuedTurns   int
	steeringDepth int
}

func (c *chatTUI) contextSummaryData() tuiContextSummary {
	data := tuiContextSummary{sessionTitle: c.sessionID, agentID: "agent", parent: "root", model: c.cfg.DefaultModel, provider: c.cfg.DefaultProvider, thinking: c.cfg.DefaultThinkingLevel, status: "idle"}
	if c.store == nil || c.sessionID == "" {
		return data
	}
	session, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		return data
	}
	messages, _ := c.store.ListMessages(context.Background(), c.sessionID)
	turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
	data.messageCount = len(messages)
	data.turnCount = len(turns)
	data.queuedTurns, _ = c.store.CountQueuedTurns(context.Background(), c.sessionID)
	data.steeringDepth, _ = c.store.SteeringQueueLength(context.Background(), c.sessionID)
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
	line := fmt.Sprintf("@%s · %s · %s · m%d/t%d", data.agentID, data.model, data.thinking, data.messageCount, data.turnCount)
	if data.queuedTurns > 0 || data.steeringDepth > 0 {
		line += fmt.Sprintf(" · q%d/s%d", data.queuedTurns, data.steeringDepth)
	}
	if data.sessionTitle != "" && data.sessionTitle != "@"+data.agentID {
		line = fmt.Sprintf("%s · %s", data.sessionTitle, line)
	}
	return c.wrapLines(line, width)
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
