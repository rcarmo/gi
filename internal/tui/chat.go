package tui

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	"github.com/rcarmo/gi/internal/inference"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/skills"
	"github.com/rcarmo/gi/internal/store"
	gitools "github.com/rcarmo/gi/internal/tools"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

const defaultTUIAgentID = "agent"

func initialSessionID(ctx context.Context, s *store.Store) (string, error) {
	if sessionID, err := s.ResolveMainSessionID(ctx, defaultTUIAgentID, "gi", "default"); err == nil {
		return sessionID, nil
	}
	return "", nil
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
		alloc := gisession.AllocateDefaultSession(defaultTUIAgentID, "gi", "default", id)
		sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(context.Background(), store.ResolveOrCreateSessionFromAllocationInput{ID: id, Title: "@" + defaultTUIAgentID, State: map[string]any{"status": "idle", "queue_count": 0, "model": cfg.DefaultModel, "provider": cfg.DefaultProvider, "thinking_level": cfg.DefaultThinkingLevel}, Allocation: alloc})
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
	cleanup := chat.Init()
	defer cleanup()
	app.SetRootComponent(chat)
	return app.Run()
}

const transcriptBlockMarkerPrefix = "⟦gi:block:"

type transcriptBlockHitTarget struct {
	Key string
	Ref *gotui.Ref
}

type transcriptBlockMeta struct {
	Key       string `json:"key"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Status    string `json:"status,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
	EndedAt   string `json:"ended_at,omitempty"`
	Detail    string `json:"detail,omitempty"`
	Footer    string `json:"footer,omitempty"`
}

type transcriptBlockSpan struct {
	HeaderIndex int
	BodyCount   int
}

type transcriptRenderableBlock struct {
	Key          string
	Kind         string
	Header       string
	Subheader    string
	Body         []string
	Expanded     bool
	Expandable   bool
	PreviewLimit int
	PreviewTail  bool
	Footer       string
	Status       string
	Selected     bool
	Border       gotui.BorderStyle
	BorderStyle  gotui.Style
	HeaderStyle  gotui.Style
	BodyStyle    gotui.Style
	HintStyle    gotui.Style
	SelectedHint string
}

// bashPreviewLines mirrors PiSwift's bash output preview window: when collapsed,
// only the trailing N lines of bash output are shown with a skipped-line hint.
const bashPreviewLines = 10

type chatTUI struct {
	app                     *gotui.App
	store                   *store.Store
	engine                  *turn.Engine
	sessionID               string
	cfg                     config.RuntimeConfig
	history                 []string
	histIdx                 int
	historySearchQuery      string
	historySearchIdx        int
	running                 bool
	status                  string
	draft                   string
	inputActive             bool
	eventCh                 chan map[string]any
	topicEventCh            chan topics.Envelope
	subscribedCh            chan map[string]any
	topicUnsubscribe        func()
	input                   *multilineInput
	queuedDrafts            []string
	inputRegion             *gotui.Element
	transcriptRegion        *gotui.Element
	transcriptRef           *gotui.Ref
	transcript              []string
	transcriptScroll        int
	stickToBottom           bool
	draftLineIndex          int
	draftLineCount          int
	outputWidth             int
	osc52Writer             io.Writer
	clipboardLookPath       func(string) (string, error)
	clipboardRun            func(context.Context, string, []string, string) error
	transcriptBlockRefs     []transcriptBlockHitTarget
	transcriptExpanded      map[string]bool
	transcriptBlockSpans    map[string]transcriptBlockSpan
	transcriptToolBlocks    map[string]string
	selectedTranscriptBlock string
	thinkingText            string
	thinkingBlockKey        string
	thinkingStartedAt       string
	thinkingIndicatorKey    string
	thinkingIndicatorStart  string
	lastInputTokens         int
	lastOutputTokens        int
	lastContextTokens       int
	lastCacheRead           int
	lastCacheWrite          int
	lastCostTotal           float64
	modelMenuOpen           bool
	modelMenuKind           string
	modelMenuValues         map[string]string
	modelMenuChoices        []string
	modelMenuAll            []string
	modelMenuQuery          string
	modelMenuSelected       int
	modelMenuScroll         int
	extensionStatuses       map[string]string
	extensionWidgets        map[string][]string
	extensionToolModes      map[string]string
}

func (c *chatTUI) ensureInput() {
	if c.input != nil {
		if c.input.onChange == nil {
			c.input.onChange = c.onInputChanged
		}
		return
	}
	c.input = newMultilineInput(80, "Send a message…", c.onSubmit, c.onInputChanged)
	c.input.onRestoreQueued = c.restoreQueuedDraft
	c.input.onComplete = c.completeInputPath
}

func (c *chatTUI) onInputChanged(string) {
	c.stickToBottom = true
	c.scrollTranscriptToBottom()
}

func encodeTranscriptBlockMarker(meta transcriptBlockMeta) string {
	payload, err := json.Marshal(meta)
	if err != nil {
		return transcriptBlockMarkerPrefix + "error⟧"
	}
	return transcriptBlockMarkerPrefix + base64.RawURLEncoding.EncodeToString(payload) + "⟧"
}

func parseTranscriptBlockMarker(line string) (transcriptBlockMeta, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, transcriptBlockMarkerPrefix) || !strings.HasSuffix(line, "⟧") {
		return transcriptBlockMeta{}, false
	}
	encoded := strings.TrimSuffix(strings.TrimPrefix(line, transcriptBlockMarkerPrefix), "⟧")
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return transcriptBlockMeta{}, false
	}
	var meta transcriptBlockMeta
	if err := json.Unmarshal(payload, &meta); err != nil || strings.TrimSpace(meta.Key) == "" {
		return transcriptBlockMeta{}, false
	}
	return meta, true
}

func (c *chatTUI) ensureTranscriptBlockState() {
	if c.transcriptExpanded == nil {
		c.transcriptExpanded = map[string]bool{}
	}
	if c.transcriptBlockSpans == nil {
		c.transcriptBlockSpans = map[string]transcriptBlockSpan{}
	}
	if c.transcriptToolBlocks == nil {
		c.transcriptToolBlocks = map[string]string{}
	}
}

func (c *chatTUI) reindexTranscriptBlocks() {
	c.ensureTranscriptBlockState()
	spans := map[string]transcriptBlockSpan{}
	for i := 0; i < len(c.transcript); i++ {
		meta, ok := parseTranscriptBlockMarker(c.transcript[i])
		if !ok {
			continue
		}
		bodyCount := 0
		for j := i + 1; j < len(c.transcript); j++ {
			if _, isBlock := parseTranscriptBlockMarker(c.transcript[j]); isBlock {
				break
			}
			if strings.HasPrefix(c.transcript[j], "│ ") {
				bodyCount++
				continue
			}
			break
		}
		spans[meta.Key] = transcriptBlockSpan{HeaderIndex: i, BodyCount: bodyCount}
	}
	c.transcriptBlockSpans = spans
	for toolKey, blockKey := range c.transcriptToolBlocks {
		if _, ok := spans[blockKey]; !ok {
			delete(c.transcriptToolBlocks, toolKey)
		}
	}
	if c.selectedTranscriptBlock != "" {
		if _, ok := spans[c.selectedTranscriptBlock]; !ok {
			c.selectedTranscriptBlock = ""
		}
	}
}

func (c *chatTUI) appendTranscriptBlock(meta transcriptBlockMeta, body []string) {
	c.ensureTranscriptBlockState()
	lines := []string{encodeTranscriptBlockMarker(meta)}
	for _, line := range body {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "" {
			trimmed = " "
		}
		lines = append(lines, "│ "+trimmed)
	}
	c.appendTranscript(lines...)
	c.reindexTranscriptBlocks()
	if meta.Key != "" {
		c.selectedTranscriptBlock = meta.Key
	}
}

func (c *chatTUI) replaceTranscriptBlock(meta transcriptBlockMeta, body []string) {
	c.ensureTranscriptBlockState()
	span, ok := c.transcriptBlockSpans[meta.Key]
	if !ok || span.HeaderIndex < 0 || span.HeaderIndex >= len(c.transcript) {
		c.appendTranscriptBlock(meta, body)
		return
	}
	prefix := append([]string(nil), c.transcript[:span.HeaderIndex]...)
	middle := []string{encodeTranscriptBlockMarker(meta)}
	for _, line := range body {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "" {
			trimmed = " "
		}
		middle = append(middle, "│ "+trimmed)
	}
	end := span.HeaderIndex + 1 + span.BodyCount
	if end > len(c.transcript) {
		end = len(c.transcript)
	}
	suffix := append([]string(nil), c.transcript[end:]...)
	c.transcript = append(prefix, append(middle, suffix...)...)
	c.applyTranscriptLimit()
	c.reindexTranscriptBlocks()
	if meta.Key != "" {
		c.selectedTranscriptBlock = meta.Key
	}
}

func (c *chatTUI) deleteTranscriptBlock(key string) bool {
	c.ensureTranscriptBlockState()
	span, ok := c.transcriptBlockSpans[key]
	if !ok || span.HeaderIndex < 0 || span.HeaderIndex >= len(c.transcript) {
		delete(c.transcriptBlockSpans, key)
		return false
	}
	if meta, ok := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex]); !ok || meta.Key != key {
		delete(c.transcriptBlockSpans, key)
		c.reindexTranscriptBlocks()
		return false
	}
	end := span.HeaderIndex + 1 + span.BodyCount
	if end > len(c.transcript) {
		end = len(c.transcript)
	}
	c.transcript = append(c.transcript[:span.HeaderIndex], c.transcript[end:]...)
	delete(c.transcriptBlockSpans, key)
	delete(c.transcriptExpanded, key)
	if c.selectedTranscriptBlock == key {
		c.selectedTranscriptBlock = ""
	}
	c.reindexTranscriptBlocks()
	return true
}

func (c *chatTUI) toggleTranscriptBlock(key string) bool {
	if strings.TrimSpace(key) == "" {
		return false
	}
	c.ensureTranscriptBlockState()
	c.selectedTranscriptBlock = key
	c.transcriptExpanded[key] = !c.transcriptExpanded[key]
	if c.app != nil {
		c.app.MarkDirty()
	}
	return true
}

func (c *chatTUI) transcriptBlockOrder() []string {
	order := make([]string, 0, len(c.transcriptBlockSpans))
	for _, line := range c.transcript {
		meta, ok := parseTranscriptBlockMarker(line)
		if ok {
			order = append(order, meta.Key)
		}
	}
	return order
}

func (c *chatTUI) selectTranscriptBlock(delta int) {
	c.reindexTranscriptBlocks()
	order := c.transcriptBlockOrder()
	if len(order) == 0 {
		return
	}
	idx := 0
	if c.selectedTranscriptBlock != "" {
		for i, key := range order {
			if key == c.selectedTranscriptBlock {
				idx = i
				break
			}
		}
	}
	idx = (idx + delta) % len(order)
	if idx < 0 {
		idx += len(order)
	}
	c.selectedTranscriptBlock = order[idx]
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) toggleSelectedTranscriptBlock() {
	c.reindexTranscriptBlocks()
	order := c.transcriptBlockOrder()
	if len(order) == 0 {
		return
	}
	key := c.selectedTranscriptBlock
	if key == "" {
		key = order[len(order)-1]
	}
	c.toggleTranscriptBlock(key)
}

func (c *chatTUI) Init() func() {
	c.eventCh = make(chan map[string]any, 64)
	c.bindSession(c.sessionID)
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.histIdx = -1
	c.historySearchIdx = -1
	c.history = c.loadCommandHistory()
	c.inputActive = true
	c.transcript = c.loadTranscript()
	c.reindexTranscriptBlocks()
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
	watchers = append(watchers, gotui.OnTimer(120*time.Millisecond, func() {
		if c.hasRunningTranscriptBlock() && c.app != nil {
			c.app.MarkDirty()
		}
	}))
	watchers = append(watchers, gotui.OnTimer(time.Second, func() {
		if c.app != nil {
			c.app.MarkDirty()
		}
	}))
	return watchers
}

func (c *chatTUI) hasRunningTranscriptBlock() bool {
	for _, line := range c.transcript {
		meta, ok := parseTranscriptBlockMarker(line)
		if ok && meta.Status == "running" {
			return true
		}
	}
	return false
}

func (c *chatTUI) handleTopicEvent(env topics.Envelope) {
	payload := env.Payload
	switch env.Topic {
	case "turn.status":
		title, _ := payload["title"].(string)
		status, _ := payload["status"].(string)
		if status == "running" {
			c.markRunning()
			c.showThinkingIndicator(env.Timestamp)
		}
		_ = title
		if status == "idle" {
			c.resetRunningDraftState()
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
	case "turn.response":
		if topicPayloadType(payload, "new_post", "agent_response") {
			return
		}
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
		if topicPayloadType(payload, "agent_draft_delta") {
			return
		}
		delta, _ := payload["delta"].(string)
		if delta != "" {
			c.markRunning()
			c.clearThinkingIndicator()
			c.draft += delta
			c.updateDraftTranscriptLine()
		}
	case "turn.thought":
		if topicPayloadType(payload, "agent_thought_delta") {
			return
		}
		c.markRunning()
		c.clearThinkingIndicator()
		delta, _ := payload["delta"].(string)
		c.updateThinkingTranscript(delta, env.Timestamp)
	case "runtime.tool":
		c.renderToolEvent(payload, env.Timestamp)
	case "runtime.hook":
		c.renderHookEvent(payload, env.Timestamp)
	case "runtime.turn":
		typ, _ := payload["type"].(string)
		status, _ := payload["status"].(string)
		switch typ {
		case "turn_usage":
			c.updateUsageFromPayload(payload)
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
				c.clearThinkingIndicator()
				c.markRunning()
			}
		}
	case "runtime.session":
		typ, _ := payload["type"].(string)
		status, _ := payload["status"].(string)
		switch typ {
		case "session_running", "session_state":
			if status == "running" {
				c.markRunning()
				c.showThinkingIndicator(env.Timestamp)
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
		c.renderRoutingEvent(payload, env.Timestamp)
	case "runtime.inbound_work":
		c.renderInboundWorkEvent(payload, env.Timestamp)
	case "runtime.dispatcher":
		c.renderDispatcherEvent(payload, env.Timestamp)
	case "session.compaction":
		c.renderCompactionEvent(payload, env.Timestamp)
	case "session.routing":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		c.renderRoutingEvent(payload, env.Timestamp)
	case "session.steering":
		c.renderSteeringEvent(payload["type"])
	case "turn.subturn":
		c.renderSubturnEvent(payload, env.Timestamp)
	case "extension.status":
		key, _ := payload["key"].(string)
		text, _ := payload["text"].(string)
		c.setExtensionStatus(key, text)
	case "extension.widget":
		key, _ := payload["key"].(string)
		c.setExtensionWidget(key, widgetPayloadLines(payload))
	case "extension.tool_render":
		tool, _ := payload["tool"].(string)
		mode, _ := payload["mode"].(string)
		c.setExtensionToolRender(tool, mode)
	}
	if c.stickToBottom {
		c.scrollTranscriptToBottom()
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func topicPayloadType(payload map[string]any, types ...string) bool {
	got, _ := payload["type"].(string)
	got = strings.TrimSpace(got)
	for _, typ := range types {
		if got == typ {
			return true
		}
	}
	return false
}

func (c *chatTUI) updateUsageFromPayload(payload map[string]any) {
	usage, _ := payload["usage"].(map[string]any)
	if usage == nil {
		return
	}
	c.lastInputTokens = intFromAny(usage["input"])
	c.lastOutputTokens = intFromAny(usage["output"])
	c.lastContextTokens = intFromAny(usage["total"])
	if c.lastContextTokens == 0 {
		c.lastContextTokens = intFromAny(usage["totalTokens"])
	}
	c.lastCacheRead = intFromAny(usage["cache_read"])
	c.lastCacheWrite = intFromAny(usage["cache_write"])
	if cost := floatFromAny(usage["cost_total"]); cost > 0 {
		c.lastCostTotal = cost
	}
}

func (c *chatTUI) useTopicNativeRuntimeStatus() bool {
	return c.topicUnsubscribe != nil
}

func (c *chatTUI) handleEvent(ev map[string]any) {
	evType, _ := ev["type"].(string)
	switch evType {
	case "agent_draft_delta":
		delta, _ := ev["delta"].(string)
		c.markRunning()
		c.clearThinkingIndicator()
		c.draft += delta
		c.updateDraftTranscriptLine()
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "new_post":
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
		c.markRunning()
		c.clearThinkingIndicator()
		delta, _ := ev["delta"].(string)
		c.updateThinkingTranscript(delta, time.Time{})
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "tool_started", "tool_finished", "tool_failed", "tool_skipped":
		if c.useTopicNativeRuntimeStatus() {
			return
		}
		payload := cloneAnyMap(ev)
		payload["type"] = evType
		c.renderToolEvent(payload, time.Time{})
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
		payload := cloneAnyMap(ev)
		payload["type"] = evType
		c.renderCompactionEvent(payload, time.Time{})
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
		payload := cloneAnyMap(ev)
		payload["type"] = evType
		c.renderRoutingEvent(payload, time.Time{})
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "agent_status":
		title := ""
		if v, ok := ev["title"].(string); ok {
			title = v
		}
		if title != "" {
			c.markRunning()
			c.showThinkingIndicator(time.Time{})
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
			c.clearDraftTranscriptLine()
			c.running = false
			c.finishThinkingTranscript(time.Now().UTC())
			c.appendTranscript("error: " + truncate(msg, 160))
			c.status = "Error"
			if c.app != nil {
				c.app.MarkDirty()
			}
		}
	}
}

func (c *chatTUI) updateDraftTranscriptLine() {
	lines := c.renderStreamingDraftLines()
	if len(lines) == 0 {
		return
	}
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		end := c.draftLineIndex + c.draftLineCount
		if c.draftLineCount <= 0 || end > len(c.transcript) {
			end = c.draftLineIndex + 1
		}
		prefix := append([]string(nil), c.transcript[:c.draftLineIndex]...)
		suffix := append([]string(nil), c.transcript[end:]...)
		c.transcript = append(prefix, append(lines, suffix...)...)
		c.draftLineCount = len(lines)
		c.applyTranscriptLimit()
		return
	}
	c.appendTranscript(lines...)
	c.draftLineIndex = len(c.transcript) - len(lines)
	c.draftLineCount = len(lines)
}

func (c *chatTUI) renderStreamingDraftLines() []string {
	text := c.draft
	if strings.TrimSpace(text) == "" {
		return []string{fmt.Sprintf("%s: %s", c.cfg.AssistantName, text)}
	}
	if looksLikeMarkdown(text) {
		return renderMarkdownTranscript(c.cfg.AssistantName+": ", text, c.transcriptRenderWidth())
	}
	return []string{fmt.Sprintf("%s: %s", c.cfg.AssistantName, text)}
}

func (c *chatTUI) finalizeDraftTranscript(text string) {
	lines := renderMarkdownTranscript(c.cfg.AssistantName+": ", text, c.transcriptRenderWidth())
	if !looksLikeMarkdown(text) {
		lines = []string{fmt.Sprintf("%s: %s", c.cfg.AssistantName, text)}
	}
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		end := c.draftLineIndex + c.draftLineCount
		if c.draftLineCount <= 0 || end > len(c.transcript) {
			end = c.draftLineIndex + 1
		}
		prefix := append([]string(nil), c.transcript[:c.draftLineIndex]...)
		suffix := append([]string(nil), c.transcript[end:]...)
		c.transcript = append(prefix, append(lines, suffix...)...)
		c.applyTranscriptLimit()
		c.draftLineIndex = -1
		c.draftLineCount = 0
		return
	}
	c.appendTranscript(lines...)
}

func (c *chatTUI) clearDraftTranscriptLine() {
	if c.draftLineIndex >= 0 && c.draftLineIndex < len(c.transcript) {
		end := c.draftLineIndex + c.draftLineCount
		if c.draftLineCount <= 0 || end > len(c.transcript) {
			end = c.draftLineIndex + 1
		}
		c.transcript = append(c.transcript[:c.draftLineIndex], c.transcript[end:]...)
	}
	c.draft = ""
	c.draftLineIndex = -1
	c.draftLineCount = 0
}

func (c *chatTUI) promoteDraftToThinking(ts time.Time) {
	text := strings.TrimSpace(c.draft)
	if text == "" {
		return
	}
	c.clearDraftTranscriptLine()
	c.updateThinkingTranscript(text, ts)
}

func (c *chatTUI) updateThinkingTranscript(delta string, ts time.Time) {
	if strings.TrimSpace(delta) == "" {
		c.showThinkingIndicator(ts)
		return
	}
	c.clearThinkingIndicator()
	c.thinkingText += delta
	if strings.TrimSpace(c.thinkingBlockKey) == "" {
		c.thinkingBlockKey = fmt.Sprintf("thought:%d", time.Now().UnixNano())
	}
	if strings.TrimSpace(c.thinkingStartedAt) == "" {
		c.thinkingStartedAt = normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)
	}
	body := renderMarkdownTranscript("", c.thinkingText, c.transcriptRenderWidth())
	meta := transcriptBlockMeta{Key: c.thinkingBlockKey, Kind: "thought", Title: "", Status: "running", StartedAt: c.thinkingStartedAt}
	c.replaceTranscriptBlock(meta, body)
}

func (c *chatTUI) finishThinkingTranscript(ts time.Time) {
	c.clearThinkingIndicator()
	if strings.TrimSpace(c.thinkingBlockKey) != "" {
		span, ok := c.transcriptBlockSpans[c.thinkingBlockKey]
		if ok && span.HeaderIndex >= 0 && span.HeaderIndex < len(c.transcript) {
			if meta, ok := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex]); ok {
				meta.Status = "done"
				meta.EndedAt = normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)
				body := c.readTranscriptBlockBody(c.thinkingBlockKey)
				c.replaceTranscriptBlock(meta, body)
			}
		}
	}
	c.thinkingText = ""
	c.thinkingBlockKey = ""
	c.thinkingStartedAt = ""
}

func (c *chatTUI) readTranscriptBlockBody(key string) []string {
	span, ok := c.transcriptBlockSpans[key]
	if !ok || span.HeaderIndex < 0 || span.HeaderIndex >= len(c.transcript) {
		return nil
	}
	body := make([]string, 0, span.BodyCount)
	for i := 0; i < span.BodyCount; i++ {
		idx := span.HeaderIndex + 1 + i
		if idx >= len(c.transcript) {
			break
		}
		body = append(body, strings.TrimPrefix(c.transcript[idx], "│ "))
	}
	return body
}

func (c *chatTUI) latestToolResultBody(turnID, toolCallID, toolName string) []string {
	if c.store == nil || strings.TrimSpace(c.sessionID) == "" {
		return nil
	}
	messages, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		return nil
	}
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != "tool_result" {
			continue
		}
		payload := msg.Payload
		if strings.TrimSpace(toolCallID) != "" {
			if got, _ := payload["tool_call_id"].(string); strings.TrimSpace(got) != strings.TrimSpace(toolCallID) {
				continue
			}
		}
		if strings.TrimSpace(turnID) != "" {
			if got, _ := payload["turn_id"].(string); strings.TrimSpace(got) != strings.TrimSpace(turnID) {
				continue
			}
		}
		if strings.TrimSpace(toolName) != "" {
			if got, _ := payload["tool_name"].(string); strings.TrimSpace(got) != strings.TrimSpace(toolName) {
				continue
			}
		}
		trimmed := strings.TrimSpace(msg.Content)
		if trimmed == "" {
			return []string{"(empty)"}
		}
		parts := strings.Split(trimmed, "\n")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			out = append(out, strings.TrimRight(part, "\r"))
		}
		return out
	}
	return nil
}

func normalizeThinkingStatus(title string) string {
	if strings.Contains(strings.ToLower(title), "thinking") {
		return "Thinking..."
	}
	return title
}

func (c *chatTUI) markRunning() {
	c.running = true
}

func (c *chatTUI) showThinkingIndicator(ts time.Time) {
	if strings.TrimSpace(c.draft) != "" || strings.TrimSpace(c.thinkingText) != "" || strings.TrimSpace(c.thinkingBlockKey) != "" {
		return
	}
	c.ensureTranscriptBlockState()
	if strings.TrimSpace(c.thinkingIndicatorKey) == "" {
		c.thinkingIndicatorKey = fmt.Sprintf("thinking:%d", time.Now().UnixNano())
		c.thinkingIndicatorStart = normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)
	}
	meta := transcriptBlockMeta{Key: c.thinkingIndicatorKey, Kind: "thinking_indicator", Title: "Thinking...", Status: "running", StartedAt: c.thinkingIndicatorStart}
	if _, ok := c.transcriptBlockSpans[c.thinkingIndicatorKey]; ok {
		c.replaceTranscriptBlock(meta, nil)
	} else {
		c.appendTranscriptBlock(meta, nil)
	}
}

func (c *chatTUI) clearThinkingIndicator() {
	key := strings.TrimSpace(c.thinkingIndicatorKey)
	if key == "" {
		return
	}
	c.deleteTranscriptBlock(key)
	c.thinkingIndicatorKey = ""
	c.thinkingIndicatorStart = ""
}

func (c *chatTUI) renderSteeringEvent(eventTypeValue any) {}

func normalizeBlockTimestamp(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}

func formatBlockClock(ts string) string {
	if parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(ts)); err == nil {
		return parsed.Local().Format("15:04:05")
	}
	return ""
}

func formatBlockElapsed(startedAt, endedAt string) string {
	start, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(startedAt))
	if err != nil {
		return ""
	}
	end := time.Now().UTC()
	if strings.TrimSpace(endedAt) != "" {
		if parsed, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(endedAt)); parseErr == nil {
			end = parsed
		}
	}
	d := end.Sub(start)
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d / time.Minute)
	seconds := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dm%02ds", minutes, seconds)
}

func (c *chatTUI) renderSubturnEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	childTurn, _ := payload["child_turn_id"].(string)
	status, _ := payload["status"].(string)
	meta := transcriptBlockMeta{Key: fmt.Sprintf("subturn:%d:%s", time.Now().UnixNano(), childTurn), Kind: "subturn", Status: strings.TrimSpace(status), StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}
	switch typ {
	case "subturn_created":
		meta.Title = "Sub-turn started"
	case "subturn_status":
		meta.Title = "Sub-turn update"
	case "subturn_result_ready", "subturn_result_delivered", "subturn_orphaned":
		meta.Title = "Sub-turn result"
	default:
		return
	}
	body := []string{}
	if childTurn != "" {
		body = append(body, "turn="+childTurn)
	}
	if status != "" {
		body = append(body, "status="+status)
	}
	c.appendTranscriptBlock(meta, body)
}

func (c *chatTUI) renderHookEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	hookName, _ := payload["hook"].(string)
	reason, _ := payload["reason"].(string)
	toolName, _ := payload["tool"].(string)
	errText, _ := payload["error"].(string)
	durationMS := intFromAny(payload["duration_ms"])
	title := "Hook event"
	status := "info"
	switch typ {
	case "hook_deny", "hook_abort":
		title, status = "Hook denied", "error"
	case "hook_modify":
		title, status = "Hook modified", "ok"
	case "hook_respond":
		title, status = "Hook responded directly", "ok"
	case "hook_invocation":
		if errText != "" {
			title, status = "Hook invocation error", "error"
		} else {
			title, status = "Hook invoked", "running"
		}
	}
	body := []string{}
	if hookName != "" {
		body = append(body, "hook="+hookName)
	}
	if toolName != "" {
		body = append(body, "tool="+toolName)
	}
	if durationMS > 0 {
		body = append(body, fmt.Sprintf("duration=%dms", durationMS))
	}
	if reason != "" {
		body = append(body, "reason="+truncate(reason, 160))
	}
	if errText != "" {
		body = append(body, "error="+truncate(errText, 160))
		if durationMS > 0 && reason == "" {
			body = append(body, fmt.Sprintf("timeout=%dms", durationMS))
		}
	}
	c.appendTranscriptBlock(transcriptBlockMeta{Key: fmt.Sprintf("hook:%d:%s", time.Now().UnixNano(), hookName), Kind: "hook", Title: title, Status: status, StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}, body)
	statusLine := strings.ToLower(title)
	if hookName != "" {
		statusLine += " via " + hookName
	}
	if toolName != "" {
		statusLine += " for " + toolName
	}
	if reason != "" {
		statusLine += ": " + truncate(reason, 120)
	} else if errText != "" {
		statusLine += ": " + truncate(errText, 120)
	}
	c.status = statusLine
}

func (c *chatTUI) renderInboundWorkEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	sourceKind, _ := payload["source_kind"].(string)
	status, _ := payload["status"].(string)
	attemptCount := intFromAny(payload["attempt_count"])
	errText, _ := payload["error"].(string)
	title := "Inbound work"
	blockStatus := "info"
	switch typ {
	case "inbound_work_enqueued":
		title = "Inbound work queued"
	case "inbound_work_retry_scheduled":
		title = "Inbound work retry scheduled"
	case "inbound_work_failed":
		title, blockStatus = "Inbound work failed", "error"
	case "inbound_work_completed":
		title, blockStatus = "Inbound work completed", "ok"
	case "inbound_work_requeued":
		title = "Inbound work requeued"
	case "inbound_work_discarded":
		title, blockStatus = "Inbound work discarded", "error"
	}
	body := []string{}
	if sourceKind != "" {
		body = append(body, "source="+sourceKind)
	}
	if status != "" {
		body = append(body, "status="+status)
	}
	if attemptCount > 0 {
		body = append(body, fmt.Sprintf("attempt=%d", attemptCount))
	}
	if errText != "" {
		body = append(body, "error="+truncate(errText, 160))
	}
	c.appendTranscriptBlock(transcriptBlockMeta{Key: fmt.Sprintf("inbound:%d:%s", time.Now().UnixNano(), typ), Kind: "dispatcher", Title: title, Status: blockStatus, StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}, body)
	c.status = strings.ToLower(title)
	if sourceKind != "" {
		c.status += fmt.Sprintf(" (%s)", sourceKind)
	}
	if attemptCount > 0 && typ == "inbound_work_retry_scheduled" {
		c.status += fmt.Sprintf(" attempt %d", attemptCount)
	}
	if status != "" {
		c.status += fmt.Sprintf(" [%s]", status)
	}
	if errText != "" {
		c.status += ": " + truncate(errText, 120)
	}
}

func (c *chatTUI) renderDispatcherEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	workerID, _ := payload["worker_id"].(string)
	processedCount := intFromAny(payload["processed_count"])
	errText, _ := payload["error"].(string)
	title := "Dispatcher event"
	status := "info"
	switch typ {
	case "dispatcher_lease_acquired":
		title = "Inbound dispatcher lease acquired"
	case "dispatcher_lease_released":
		title = "Inbound dispatcher lease released"
	case "dispatcher_drain_completed":
		title, status = "Inbound dispatcher drain completed", "ok"
	case "dispatcher_error":
		title, status = "Inbound dispatcher error", "error"
	}
	body := []string{}
	if workerID != "" {
		body = append(body, "worker="+workerID)
	}
	if processedCount > 0 {
		body = append(body, fmt.Sprintf("processed=%d", processedCount))
	}
	if errText != "" {
		body = append(body, "error="+truncate(errText, 160))
	}
	c.appendTranscriptBlock(transcriptBlockMeta{Key: fmt.Sprintf("dispatcher:%d:%s", time.Now().UnixNano(), typ), Kind: "dispatcher", Title: title, Status: status, StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}, body)
	c.status = strings.ToLower(title)
	if processedCount > 0 && typ == "dispatcher_drain_completed" {
		c.status += fmt.Sprintf(" (%d processed)", processedCount)
	}
	if workerID != "" {
		c.status += fmt.Sprintf(" [%s]", workerID)
	}
	if errText != "" {
		c.status += ": " + truncate(errText, 120)
	}
}

func (c *chatTUI) renderCompactionEvent(payload map[string]any, ts time.Time) {
	before := intFromAny(payload["messages_before"])
	after := intFromAny(payload["messages_after"])
	tokens := intFromAny(payload["tokens_before"])
	c.appendTranscriptBlock(transcriptBlockMeta{Key: fmt.Sprintf("compact:%d", time.Now().UnixNano()), Kind: "compact", Title: "Context compacted", Status: "ok", StartedAt: normalizeBlockTimestamp(ts).Format(time.RFC3339Nano)}, []string{fmt.Sprintf("messages=%d→%d", before, after), fmt.Sprintf("tokens_before=%d", tokens)})
	c.status = "Compacted context"
}

func (c *chatTUI) toolRuntimeBlockKey(payload map[string]any, toolName string) string {
	toolCallID, _ := payload["tool_call_id"].(string)
	turnID, _ := payload["turn_id"].(string)
	iteration := intFromAny(payload["iteration"])
	if strings.TrimSpace(toolCallID) != "" {
		return "toolcall:" + toolCallID
	}
	return fmt.Sprintf("tool:%s:%s:%d", strings.TrimSpace(turnID), strings.TrimSpace(toolName), iteration)
}

func (c *chatTUI) toolInvocationBody(toolName string, payload map[string]any) []string {
	if payload == nil {
		return nil
	}
	if command := toolInvocationText(toolName, payload["arguments"]); command != "" {
		return []string{command}
	}
	return nil
}

func toolInvocationText(toolName string, args any) string {
	m, ok := args.(map[string]any)
	if !ok || len(m) == 0 {
		if raw, ok := args.(map[string]interface{}); ok && len(raw) > 0 {
			m = map[string]any{}
			for k, v := range raw {
				m[k] = v
			}
		} else {
			return ""
		}
	}
	for _, key := range []string{"command", "cmd", "script", "path", "query"} {
		if v, ok := m[key]; ok {
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
				return truncate(strings.Join(strings.Fields(s), " "), 500)
			}
		}
	}
	payload, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return truncate(string(payload), 500)
}

func (c *chatTUI) toolResultBody(payload map[string]any, turnID, toolCallID, toolName string) []string {
	if output, _ := payload["output"].(string); strings.TrimSpace(output) != "" {
		return toolOutputBodyLines(output)
	}
	return c.latestToolResultBody(turnID, toolCallID, toolName)
}

func toolOutputBodyLines(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{"(empty)"}
	}
	parts := strings.Split(text, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, strings.TrimRight(part, "\r"))
	}
	return out
}

func (c *chatTUI) renderToolEvent(payload map[string]any, ts time.Time) {
	c.ensureTranscriptBlockState()
	typ, _ := payload["type"].(string)
	toolName, _ := payload["tool"].(string)
	errText, _ := payload["error"].(string)
	reason, _ := payload["reason"].(string)
	turnID, _ := payload["turn_id"].(string)
	toolCallID, _ := payload["tool_call_id"].(string)
	if strings.TrimSpace(toolName) == "" {
		toolName = "tool"
	}
	startedAt := normalizeBlockTimestamp(ts)
	toolKey := c.toolRuntimeBlockKey(payload, toolName)
	blockKey := c.transcriptToolBlocks[toolKey]
	meta := transcriptBlockMeta{Key: blockKey, Kind: "tool", Title: toolName}
	body := c.toolInvocationBody(toolName, payload)
	switch typ {
	case "tool_started":
		c.promoteDraftToThinking(startedAt)
		c.finishThinkingTranscript(startedAt)
		c.markRunning()
		meta.Status = "running"
		meta.StartedAt = startedAt.Format(time.RFC3339Nano)
		if meta.Key == "" {
			meta.Key = fmt.Sprintf("tool:%d:%s", time.Now().UnixNano(), toolName)
			c.transcriptToolBlocks[toolKey] = meta.Key
			c.appendTranscriptBlock(meta, body)
		} else {
			oldMeta, ok := parseTranscriptBlockMarker(c.transcript[c.transcriptBlockSpans[meta.Key].HeaderIndex])
			if ok && oldMeta.StartedAt != "" {
				meta.StartedAt = oldMeta.StartedAt
			}
			c.replaceTranscriptBlock(meta, body)
		}
	case "tool_finished", "tool_failed", "tool_skipped":
		if meta.Key == "" {
			meta.Key = fmt.Sprintf("tool:%d:%s", time.Now().UnixNano(), toolName)
		}
		existingBody := c.readTranscriptBlockBody(meta.Key)
		if span, ok := c.transcriptBlockSpans[meta.Key]; ok && span.HeaderIndex < len(c.transcript) {
			if oldMeta, ok := parseTranscriptBlockMarker(c.transcript[span.HeaderIndex]); ok && oldMeta.StartedAt != "" {
				meta.StartedAt = oldMeta.StartedAt
			}
		}
		if meta.StartedAt == "" {
			meta.StartedAt = startedAt.Format(time.RFC3339Nano)
		}
		meta.EndedAt = startedAt.Format(time.RFC3339Nano)
		if len(existingBody) > 0 {
			body = append([]string(nil), existingBody...)
		}
		switch typ {
		case "tool_finished":
			meta.Status = "ok"
		case "tool_failed":
			meta.Status = "error"
			if errText != "" {
				body = append(body, "error="+truncate(errText, 160))
			}
		case "tool_skipped":
			meta.Status = "skipped"
			if reason != "" {
				body = append(body, "reason="+truncate(reason, 160))
			}
		}
		if resultBody := c.toolResultBody(payload, turnID, toolCallID, toolName); len(resultBody) > 0 {
			body = append(body, resultBody...)
		}
		if meta.Key == "" {
			c.appendTranscriptBlock(meta, body)
		} else {
			c.replaceTranscriptBlock(meta, body)
		}
		c.transcriptToolBlocks[toolKey] = meta.Key
	}
}

func (c *chatTUI) renderRoutingEvent(payload map[string]any, ts time.Time) {
	typ, _ := payload["type"].(string)
	targetAgent, _ := payload["target_agent_id"].(string)
	sourceAgent, _ := payload["source_agent_id"].(string)
	switch typ {
	case "routing_decision":
		if targetAgent != "" {
			c.status = fmt.Sprintf("routed to @%s", targetAgent)
		}
	case "routing_incoming":
		if sourceAgent != "" {
			c.status = fmt.Sprintf("incoming route from @%s", sourceAgent)
		}
	}
}

func (c *chatTUI) resetRunningDraftState() {
	c.clearThinkingIndicator()
	c.finishThinkingTranscript(time.Now().UTC())
	// Preserve any streamed assistant draft when terminal/runtime status events
	// arrive before the final response event. This keeps partial/final output in
	// the timeline instead of erasing it on turn_completed/session_idle.
	c.running = false
	c.draft = ""
}

func intFromEvent(ev map[string]any, key string) int {
	return intFromAny(ev[key])
}

func cloneAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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

func floatFromAny(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func (c *chatTUI) KeyMap() gotui.KeyMap {
	if c.modelMenuOpen {
		return gotui.KeyMap{
			gotui.OnStop(gotui.KeyCtrlC, func(ke gotui.KeyEvent) { c.app.Stop() }),
			gotui.OnPreemptStop(gotui.KeyEscape, func(ke gotui.KeyEvent) { c.closeModelMenu() }),
			gotui.OnPreemptStop(gotui.KeyUp, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(-1) }),
			gotui.OnPreemptStop(gotui.KeyDown, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(1) }),
			gotui.OnPreemptStop(gotui.KeyPageUp, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(-5) }),
			gotui.OnPreemptStop(gotui.KeyPageDown, func(ke gotui.KeyEvent) { c.moveModelMenuSelection(5) }),
			gotui.OnPreemptStop(gotui.KeyHome, func(ke gotui.KeyEvent) { c.setModelMenuSelection(0) }),
			gotui.OnPreemptStop(gotui.KeyEnd, func(ke gotui.KeyEvent) { c.setModelMenuSelection(len(c.modelMenuChoices) - 1) }),
			gotui.OnPreemptStop(gotui.KeyEnter, func(ke gotui.KeyEvent) { c.acceptModelMenuSelection() }),
			gotui.OnPreemptStop(gotui.KeyBackspace, func(ke gotui.KeyEvent) { c.modelMenuBackspace() }),
			gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) { c.modelMenuTypeRune(ke.Rune) }),
		}
	}
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
		gotui.OnPreemptStop(gotui.KeyF6, func(ke gotui.KeyEvent) { c.selectTranscriptBlock(-1) }),
		gotui.OnPreemptStop(gotui.KeyF7, func(ke gotui.KeyEvent) { c.selectTranscriptBlock(1) }),
		gotui.OnPreemptStop(gotui.KeyF8, func(ke gotui.KeyEvent) { c.toggleSelectedTranscriptBlock() }),
		gotui.OnPreemptStop(gotui.KeyF2, func(ke gotui.KeyEvent) { c.recallHistory(-1) }),
		gotui.OnPreemptStop(gotui.KeyF3, func(ke gotui.KeyEvent) { c.recallHistory(1) }),
		gotui.OnPreemptStop(gotui.Rune('p').Ctrl(), func(ke gotui.KeyEvent) { c.recallHistory(-1) }),
		gotui.OnPreemptStop(gotui.Rune('n').Ctrl(), func(ke gotui.KeyEvent) { c.recallHistory(1) }),
		gotui.OnPreemptStop(gotui.KeyCtrlL, func(ke gotui.KeyEvent) { c.cycleModel(1) }),
		gotui.OnPreemptStop(gotui.Rune('l').Alt(), func(ke gotui.KeyEvent) { c.cycleModel(-1) }),
		gotui.OnPreemptStop(gotui.KeyCtrlT, func(ke gotui.KeyEvent) { c.cycleThinking(1) }),
		gotui.OnPreemptStop(gotui.KeyCtrlR, func(ke gotui.KeyEvent) { c.searchHistoryBackward() }),
		gotui.OnPreemptStop(gotui.Rune('t').Alt(), func(ke gotui.KeyEvent) { c.cycleThinking(-1) }),
		gotui.OnPreemptStop(gotui.KeyUp, func(ke gotui.KeyEvent) {
			c.recallHistory(-1)
		}),
		gotui.OnPreemptStop(gotui.KeyDown, func(ke gotui.KeyEvent) {
			c.recallHistory(1)
		}),
	}
}

func (c *chatTUI) openModelMenu() {
	choices := c.availableModelChoices()
	if len(choices) == 0 {
		c.appendTranscript("sys: no available models; use /model <provider/model>")
		return
	}
	selected := 0
	current := canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel)
	for i, model := range choices {
		if canonicalModelRef(c.cfg.DefaultProvider, model) == current {
			selected = i
			break
		}
	}
	c.modelMenuOpen = true
	c.modelMenuKind = "model"
	c.modelMenuValues = nil
	c.modelMenuAll = choices
	c.modelMenuQuery = ""
	c.modelMenuChoices = choices
	c.modelMenuSelected = selected
	c.modelMenuScroll = 0
	c.ensureModelMenuSelectionVisible()
	c.inputActive = false
	if c.app != nil {
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
}

func (c *chatTUI) openSessionMenu() {
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		c.appendTranscript(fmt.Sprintf("error: list sessions: %v", err))
		return
	}
	if len(sessions) == 0 {
		c.appendTranscript("sys: no sessions to resume")
		return
	}
	labels := make([]string, 0, len(sessions))
	values := map[string]string{}
	selected := 0
	for i := range sessions {
		sess := sessions[i]
		status, _ := sess.State["status"].(string)
		if status == "" {
			status = "idle"
		}
		label := fmt.Sprintf("@%s %s (%s) · %s", c.agentIDForSession(&sess), strings.TrimSpace(sess.Title), compactID(sess.ID), status)
		labels = append(labels, label)
		values[label] = sess.ID
		if sess.ID == c.sessionID {
			selected = i
		}
	}
	c.modelMenuOpen = true
	c.modelMenuKind = "session"
	c.modelMenuValues = values
	c.modelMenuAll = labels
	c.modelMenuQuery = ""
	c.modelMenuChoices = labels
	c.modelMenuSelected = selected
	c.modelMenuScroll = 0
	c.ensureModelMenuSelectionVisible()
	c.inputActive = false
	if c.app != nil {
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
}

func fuzzyMatch(query, candidate string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	candidate = strings.ToLower(candidate)
	// Each whitespace-separated token must appear as a substring; this keeps
	// filtering intuitive ("gpt" matches only gpt models) while still allowing
	// multi-term queries like "openai mini".
	for _, token := range strings.Fields(query) {
		if !strings.Contains(candidate, token) {
			return false
		}
	}
	return true
}

func filterModelMenuChoices(all []string, query string) []string {
	if strings.TrimSpace(query) == "" {
		return append([]string(nil), all...)
	}
	out := make([]string, 0, len(all))
	for _, model := range all {
		if fuzzyMatch(query, model) {
			out = append(out, model)
		}
	}
	return out
}

func (c *chatTUI) applyModelMenuFilter() {
	c.modelMenuChoices = filterModelMenuChoices(c.modelMenuAll, c.modelMenuQuery)
	c.modelMenuSelected = 0
	c.modelMenuScroll = 0
	c.ensureModelMenuSelectionVisible()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) modelMenuTypeRune(r rune) {
	if r == 0 {
		return
	}
	c.modelMenuQuery += string(r)
	c.applyModelMenuFilter()
}

func (c *chatTUI) modelMenuBackspace() {
	if c.modelMenuQuery == "" {
		return
	}
	q := []rune(c.modelMenuQuery)
	c.modelMenuQuery = string(q[:len(q)-1])
	c.applyModelMenuFilter()
}

func (c *chatTUI) closeModelMenu() {
	c.modelMenuOpen = false
	c.modelMenuKind = ""
	c.modelMenuValues = nil
	c.modelMenuChoices = nil
	c.modelMenuAll = nil
	c.modelMenuQuery = ""
	c.modelMenuSelected = 0
	c.modelMenuScroll = 0
	c.focusInput()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) moveModelMenuSelection(delta int) {
	c.setModelMenuSelection(c.modelMenuSelected + delta)
}

func (c *chatTUI) setModelMenuSelection(idx int) {
	if len(c.modelMenuChoices) == 0 {
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(c.modelMenuChoices) {
		idx = len(c.modelMenuChoices) - 1
	}
	c.modelMenuSelected = idx
	c.ensureModelMenuSelectionVisible()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) modelMenuVisibleRows() int {
	rows := 7
	if c.app != nil {
		_, h := c.app.Size()
		if h < 24 {
			rows = 5
		}
	}
	if rows < 3 {
		rows = 3
	}
	return rows
}

func (c *chatTUI) ensureModelMenuSelectionVisible() {
	rows := c.modelMenuVisibleRows()
	if c.modelMenuSelected < c.modelMenuScroll {
		c.modelMenuScroll = c.modelMenuSelected
	}
	if c.modelMenuSelected >= c.modelMenuScroll+rows {
		c.modelMenuScroll = c.modelMenuSelected - rows + 1
	}
	if c.modelMenuScroll < 0 {
		c.modelMenuScroll = 0
	}
	maxScroll := len(c.modelMenuChoices) - rows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if c.modelMenuScroll > maxScroll {
		c.modelMenuScroll = maxScroll
	}
}

func (c *chatTUI) acceptModelMenuSelection() {
	if !c.modelMenuOpen || len(c.modelMenuChoices) == 0 || c.modelMenuSelected < 0 || c.modelMenuSelected >= len(c.modelMenuChoices) {
		return
	}
	label := c.modelMenuChoices[c.modelMenuSelected]
	kind := c.modelMenuKind
	value := label
	if c.modelMenuValues != nil {
		if v, ok := c.modelMenuValues[label]; ok {
			value = v
		}
	}
	c.modelMenuOpen = false
	c.modelMenuKind = ""
	c.modelMenuValues = nil
	c.modelMenuChoices = nil
	c.modelMenuAll = nil
	c.modelMenuQuery = ""
	c.modelMenuScroll = 0
	c.focusInput()
	switch kind {
	case "session":
		c.switchSession(value)
		c.appendTranscript(fmt.Sprintf("sys: resumed %s", value))
	default:
		c.appendTranscript(c.modelCommand([]string{"/model", value})...)
	}
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) modelMenuHeight() int {
	if !c.modelMenuOpen {
		return 0
	}
	rows := c.modelMenuVisibleRows()
	if len(c.modelMenuChoices) < rows {
		rows = len(c.modelMenuChoices)
	}
	// rows + title + search line
	return rows + 3
}

func (c *chatTUI) renderModelMenu(width int) *gotui.Element {
	rows := c.modelMenuVisibleRows()
	start := c.modelMenuScroll
	end := start + rows
	if end > len(c.modelMenuChoices) {
		end = len(c.modelMenuChoices)
	}
	menu := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithHeight(c.modelMenuHeight()),
		gotui.WithDirection(gotui.Column),
		gotui.WithBorder(gotui.BorderRounded),
		gotui.WithBorderStyle(gotui.NewStyle().Foreground(gotui.Blue)),
		gotui.WithPaddingTRBL(0, 1, 0, 1),
	)
	current := strings.TrimSpace(c.cfg.DefaultModel)
	noun := "model"
	if c.modelMenuKind == "session" {
		noun = "session"
	}
	title := "Select " + noun + " · ↑/↓ navigate · Enter select · Esc cancel"
	if width < 72 {
		title = "Select " + noun + " · ↑/↓ Enter Esc"
	}
	if c.modelMenuKind != "session" && current != "" {
		title += " · current " + compactMaybe(current, c.compactOutput(), 28)
	}
	menu.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(truncate(title, max(20, width-4))), gotui.WithTextStyle(gotui.NewStyle().Bold())))
	search := "search: " + c.modelMenuQuery + "▌"
	if strings.TrimSpace(c.modelMenuQuery) == "" {
		search = "search: (type to filter)"
	} else {
		search += fmt.Sprintf("  (%d match)", len(c.modelMenuChoices))
	}
	menu.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(truncate(search, max(20, width-4))), gotui.WithTextStyle(gotui.NewStyle().Dim())))
	if len(c.modelMenuChoices) == 0 {
		menu.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText("  no matching models"), gotui.WithTextStyle(gotui.NewStyle().Dim())))
		return menu
	}
	for i := start; i < end; i++ {
		model := c.modelMenuChoices[i]
		prefix := "  "
		style := gotui.NewStyle()
		if i == c.modelMenuSelected {
			prefix = "› "
			style = style.Reverse().Bold()
		} else if canonicalModelRef(c.cfg.DefaultProvider, model) == canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel) {
			prefix = "* "
			style = style.Foreground(gotui.Cyan)
		}
		label := fmt.Sprintf("%s%d. %s", prefix, i+1, model)
		menu.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(truncate(label, max(20, width-4))), gotui.WithTextStyle(style)))
	}
	return menu
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

func (c *chatTUI) loadCommandHistory() []string {
	if c.store == nil || strings.TrimSpace(c.sessionID) == "" {
		return nil
	}
	messages, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		return nil
	}
	history := make([]string, 0, len(messages))
	for _, msg := range messages {
		if msg.Role != "user" || strings.TrimSpace(msg.Content) == "" {
			continue
		}
		history = append(history, strings.TrimSpace(msg.Content))
	}
	limit := c.currentHistoryLimit()
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}
	return history
}

func (c *chatTUI) currentHistoryLimit() int {
	if c.cfg.TUIHistoryLimit > 0 {
		return c.cfg.TUIHistoryLimit
	}
	return 10000
}

func (c *chatTUI) applyHistoryLimit() {
	limit := c.currentHistoryLimit()
	if limit <= 0 || len(c.history) <= limit {
		return
	}
	c.history = append([]string(nil), c.history[len(c.history)-limit:]...)
}

func (c *chatTUI) searchHistoryBackward() {
	if c.input == nil || len(c.history) == 0 {
		return
	}
	query := strings.TrimSpace(c.input.Text())
	if query == "" {
		c.recallHistory(-1)
		return
	}
	start := len(c.history) - 1
	if c.historySearchQuery == query && c.historySearchIdx > 0 {
		start = c.historySearchIdx - 1
	}
	needle := strings.ToLower(query)
	for i := start; i >= 0; i-- {
		if strings.Contains(strings.ToLower(c.history[i]), needle) {
			c.historySearchQuery = query
			c.historySearchIdx = i
			c.focusInput()
			c.input.SetText(c.history[i])
			c.status = fmt.Sprintf("History match %d/%d", i+1, len(c.history))
			return
		}
	}
	c.status = fmt.Sprintf("No history match for %q", query)
	if c.app != nil {
		c.app.MarkDirty()
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
	if c.handleTranscriptScrollEvent(me) {
		return true
	}
	if me.Button != gotui.MouseLeft || me.Action != gotui.MousePress {
		return false
	}
	if c.handleTranscriptBlockClick(me) {
		return true
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

func (c *chatTUI) handleTranscriptBlockClick(me gotui.MouseEvent) bool {
	for _, target := range c.transcriptBlockRefs {
		if target.Ref == nil || target.Ref.El() == nil || !target.Ref.El().ContainsPoint(me.X, me.Y) {
			continue
		}
		return c.toggleTranscriptBlock(target.Key)
	}
	return false
}

func (c *chatTUI) handleTranscriptScrollEvent(me gotui.MouseEvent) bool {
	switch me.Button {
	case gotui.MouseWheelUp, gotui.MouseWheelDown:
	default:
		return false
	}
	if c.transcriptRegion != nil && c.transcriptRegion.ContainsPoint(me.X, me.Y) {
		if c.transcriptRegion.HandleEvent(me) {
			_, y := c.transcriptRegion.ScrollOffset()
			c.transcriptScroll = y
			lines := c.visibleTranscript()
			maxScroll := len(lines) - c.transcriptViewportHeight()
			if maxScroll < 0 {
				maxScroll = 0
			}
			c.stickToBottom = c.transcriptScroll >= maxScroll
			if c.app != nil {
				c.app.MarkDirty()
			}
			return true
		}
		return false
	}
	if c.transcriptRegion == nil {
		switch me.Button {
		case gotui.MouseWheelUp:
			c.scrollTranscript(-3)
			return true
		case gotui.MouseWheelDown:
			c.scrollTranscript(3)
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
	c.applyHistoryLimit()
	c.histIdx = -1
	c.historySearchIdx = -1
	c.historySearchQuery = ""
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
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.draft = ""
	c.draftLineIndex = -1
	c.draftLineCount = 0
	c.appendTranscript(fmt.Sprintf("you: %s", text))
	c.showThinkingIndicator(time.Now())
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
		if result != nil && strings.TrimSpace(result.SessionID) != "" && result.SessionID != c.sessionID {
			apply := func() {
				c.switchSession(result.SessionID)
				c.running = result.Status == "running" || result.Status == "queued"
				if c.running {
					c.showThinkingIndicator(time.Now())
				}
				if result.Routed {
					c.appendTranscript(fmt.Sprintf("sys: routed to @%s (%s)", result.TargetAgentID, result.SessionID))
				}
				c.stickToBottom = true
				c.scrollTranscriptToBottom()
			}
			if c.app != nil {
				c.app.QueueUpdate(func() {
					apply()
					c.app.MarkDirty()
				})
			} else {
				apply()
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
	case "/sessions":
		c.openSessionMenu()
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
		if len(fields) == 1 {
			c.openModelMenu()
		} else {
			c.transcript = append(c.transcript, c.modelCommand(fields)...)
		}
	case "/scoped-models":
		c.transcript = append(c.transcript, c.scopedModelsCommand(fields)...)
	case "/thinking":
		c.transcript = append(c.transcript, c.thinkingCommand(fields)...)
	case "/compact":
		c.appendTranscript(c.compactLines()...)
	case "/scrollback":
		c.appendTranscript(c.scrollbackCommand(fields)...)
	case "/history-limit":
		c.appendTranscript(c.historyLimitCommand(fields)...)
	case "/scrollbar":
		c.appendTranscript(c.scrollbarCommand(fields)...)
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
			c.appendTranscript("sys: commands: /help, /commands [query], /session, /new, /name <name>, /resume [index|session_id], /clone [@agentN], /copy [--osc52|--native|--auto|--fallback], /attach <path> [prompt], /reload, /tools [query|active|activate|reset], /skills [query], /skill:name [args], /model [name], /scoped-models [add|remove|set], /thinking [level], /compact, /scrollback [n], /history-limit [n], /settings, /approvals, /cancel, /agents, /tree, /plugins, /fork [@agentN], /switch @agent|session_id, /send @agent message, /where, !cmd, !!cmd")
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
		{"/sessions", "searchable session resume selector"},
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
		{"/history-limit [n]", "show or set TUI command history limit"},
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
		"enter send · shift-enter newline · esc blur · ctrl-d exit · f6/f7 select block · f8 expand",
		"/commands  all commands",
		"/model     choose model · type to filter · ctrl-l cycles",
		"/session   details for this chat",
		"/where     compact context",
		"/attach    add media",
		"ctrl-r     search command history (current input is query)",
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
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	if root := strings.TrimSpace(c.cfg.WorkspaceRoot); root != "" {
		cmd.Dir = root
	}
	out, err := cmd.CombinedOutput()
	status := "ok"
	if err != nil {
		status = "error"
	}
	return c.bashBlockLines(command, string(out), status, err, startedAt, time.Now())
}

// bashBlockLines builds a PiSwift-style bash execution transcript block: a
// header with the command, the full output retained in the block body (capped),
// and a footer pointing at a full-output file when the output is truncated.
func (c *chatTUI) bashBlockLines(command, output, status string, runErr error, startedAt, endedAt time.Time) []string {
	const maxBodyLines = 500
	text := strings.ReplaceAll(output, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimRight(text, "\n")
	var rawLines []string
	if strings.TrimSpace(text) == "" {
		rawLines = []string{"(no output)"}
	} else {
		rawLines = strings.Split(text, "\n")
	}
	footer := ""
	bodyLines := rawLines
	if len(rawLines) > maxBodyLines {
		if path, perr := c.writeBashFullOutput(command, output); perr == nil && path != "" {
			footer = "output truncated · full output: " + path
		} else {
			footer = fmt.Sprintf("output truncated · %d lines hidden", len(rawLines)-maxBodyLines)
		}
		bodyLines = rawLines[len(rawLines)-maxBodyLines:]
	}
	if runErr != nil {
		bodyLines = append(bodyLines, "error: "+truncate(runErr.Error(), 160))
	}
	meta := transcriptBlockMeta{
		Key:       fmt.Sprintf("bash:%d", time.Now().UnixNano()),
		Kind:      "bash",
		Title:     "$ " + command,
		Status:    status,
		StartedAt: startedAt.Format(time.RFC3339Nano),
		EndedAt:   endedAt.Format(time.RFC3339Nano),
		Footer:    footer,
	}
	lines := []string{encodeTranscriptBlockMarker(meta)}
	for _, line := range bodyLines {
		lines = append(lines, "│ "+line)
	}
	return lines
}

func (c *chatTUI) writeBashFullOutput(command, output string) (string, error) {
	dir := strings.TrimSpace(c.cfg.WorkspaceRoot)
	if dir == "" {
		dir = os.TempDir()
	} else {
		dir = filepath.Join(dir, ".gi-run", "bash-output")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("bash-%d.txt", time.Now().UnixNano()))
	header := "$ " + command + "\n\n"
	if err := os.WriteFile(path, []byte(header+output), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (c *chatTUI) reloadLines() []string {
	workspace := c.cfg.WorkspaceRoot
	if strings.TrimSpace(workspace) == "" {
		workspace = config.DefaultWorkspaceRoot()
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

func splitModelRef(defaultProvider, label string) (string, string) {
	label = strings.TrimSpace(label)
	if label == "" {
		return strings.TrimSpace(defaultProvider), ""
	}
	if slash := strings.Index(label, "/"); slash > 0 {
		return strings.TrimSpace(label[:slash]), strings.TrimSpace(label[slash+1:])
	}
	return strings.TrimSpace(defaultProvider), label
}

func canonicalModelRef(defaultProvider, label string) string {
	provider, id := splitModelRef(defaultProvider, label)
	if id == "" {
		return ""
	}
	if provider == "" {
		return id
	}
	return provider + "/" + id
}

func (c *chatTUI) availableModelChoices() []string {
	choices := make([]string, 0, len(c.cfg.EnabledModels)+8)
	seen := map[string]bool{}
	appendChoice := func(label string) {
		label = strings.TrimSpace(label)
		if label == "" {
			return
		}
		key := canonicalModelRef(c.cfg.DefaultProvider, label)
		if key == "" {
			key = label
		}
		if seen[key] {
			return
		}
		seen[key] = true
		choices = append(choices, label)
	}
	for _, model := range c.cfg.EnabledModels {
		appendChoice(model)
	}
	_, modelOptions := inference.ListRuntimeOptions(c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.EnabledModels)
	for _, option := range modelOptions {
		appendChoice(option.Label)
	}
	appendChoice(c.cfg.DefaultModel)
	return choices
}

func (c *chatTUI) modelCommand(fields []string) []string {
	if len(fields) == 1 {
		return c.modelListLines()
	}
	model := strings.TrimSpace(strings.Join(fields[1:], " "))
	if model == "" {
		return []string{"sys: usage /model <model|index>"}
	}
	if idx, err := strconv.Atoi(model); err == nil {
		choices := c.availableModelChoices()
		if idx > 0 && idx <= len(choices) {
			model = choices[idx-1]
		}
	}
	c.cfg.DefaultModel = model
	if strings.Contains(model, "/") {
		c.cfg.DefaultProvider = strings.SplitN(model, "/", 2)[0]
	}
	lines := []string{fmt.Sprintf("model: %s", model)}
	if c.store != nil && strings.TrimSpace(c.sessionID) != "" {
		if err := c.store.TouchSessionState(context.Background(), c.sessionID, map[string]any{"model": model}); err != nil {
			lines = append(lines, fmt.Sprintf("warn: failed to persist model in session state: %v", err))
		}
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
	choices := c.availableModelChoices()
	if len(choices) == 0 {
		return append(lines, "no available models · /model <provider/model>")
	}
	for i, model := range choices {
		marker := " "
		if canonicalModelRef(c.cfg.DefaultProvider, model) == canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel) {
			marker = "›"
		}
		lines = append(lines, fmt.Sprintf("%s %d  %s", marker, i+1, compactMaybe(model, c.compactOutput(), modelWidth)))
	}
	if len(c.cfg.EnabledModels) > 0 {
		lines = append(lines, fmt.Sprintf("enabled: %d · /scoped-models list to manage pinned models", len(c.cfg.EnabledModels)))
	}
	lines = append(lines, "/model <n> to switch · ctrl-l cycles enabled models")
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
		oldIndex := c.draftLineIndex
		c.draftLineIndex -= drop
		if c.draftLineIndex < 0 || drop >= oldIndex+c.draftLineCount {
			c.draftLineIndex = -1
			c.draftLineCount = 0
		}
	}
	c.reindexTranscriptBlocks()
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
		fmt.Sprintf("- history_limit: %d", c.currentHistoryLimit()),
		fmt.Sprintf("- scrollbar: %v", c.cfg.TUIScrollbar),
		"- shortcuts: Ctrl+L/Alt+L model cycle, Ctrl+T/Alt+T thinking cycle, Ctrl+R history search, Tab path completion, @path completion, F6/F7 transcript block select, F8 expand/collapse",
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

func (c *chatTUI) scrollbarCommand(fields []string) []string {
	if len(fields) == 1 {
		state := "off"
		if c.cfg.TUIScrollbar {
			state = "on"
		}
		return []string{fmt.Sprintf("sys: scrollbar: %s", state)}
	}
	value := strings.ToLower(strings.TrimSpace(fields[1]))
	enabled := false
	switch value {
	case "on", "true", "1", "yes", "enabled":
		enabled = true
	case "off", "false", "0", "no", "disabled":
		enabled = false
	default:
		return []string{"sys: usage /scrollbar <on|off>"}
	}
	c.cfg.TUIScrollbar = enabled
	state := "off"
	if enabled {
		state = "on"
	}
	lines := []string{fmt.Sprintf("sys: scrollbar set to %s", state)}
	if err := config.PersistTUIScrollbar(c.cfg.WorkspaceRoot, enabled); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist scrollbar setting: %v", err))
	}
	return lines
}

func (c *chatTUI) historyLimitCommand(fields []string) []string {
	if len(fields) == 1 {
		return []string{fmt.Sprintf("sys: history limit: %d", c.currentHistoryLimit())}
	}
	limit, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil || limit <= 0 {
		return []string{"sys: usage /history-limit <positive-limit>"}
	}
	c.cfg.TUIHistoryLimit = limit
	c.applyHistoryLimit()
	lines := []string{fmt.Sprintf("sys: history limit set to %d", limit)}
	if err := config.PersistTUIHistoryLimit(c.cfg.WorkspaceRoot, limit); err != nil {
		lines = append(lines, fmt.Sprintf("warn: failed to persist history limit: %v", err))
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
	footerLines := c.footerLines(contentWidth)
	widgetLines := c.extensionWidgetLines()
	menuHeight := c.modelMenuHeight()
	reservedHeight := (padding * 2) + len(footerLines) + len(widgetLines) + inputHeight + 2 + menuHeight
	transcriptHeight := h - reservedHeight
	if transcriptHeight < 4 {
		transcriptHeight = 4
	}
	transcriptOptions := []gotui.Option{
		gotui.WithWidthPercent(100),
		gotui.WithHeight(transcriptHeight),
		gotui.WithScrollable(gotui.ScrollVertical),
		gotui.WithScrollOffset(0, c.transcriptScroll),
		gotui.WithDirection(gotui.Column),
	}
	if c.cfg.TUIScrollbar {
		transcriptOptions = append(transcriptOptions, gotui.WithScrollbarStyle(gotui.NewStyle().Dim()))
	}
	transcript := gotui.New(transcriptOptions...)
	c.transcriptRef.Set(transcript)
	c.transcriptRegion = transcript
	c.transcriptBlockRefs = nil
	if c.transcriptExpanded == nil {
		c.transcriptExpanded = map[string]bool{}
	}
	blocks := c.buildTranscriptRenderableBlocks(c.visibleTranscript())
	if c.selectedTranscriptBlock == "" {
		for i := len(blocks) - 1; i >= 0; i-- {
			if blocks[i].Key != "" && (len(blocks[i].Body) > 0 || blocks[i].Subheader != "") {
				c.selectedTranscriptBlock = blocks[i].Key
				break
			}
		}
		blocks = c.buildTranscriptRenderableBlocks(c.visibleTranscript())
	}
	for _, block := range blocks {
		transcript.AddChild(c.renderTranscriptBlock(block))
	}
	root.AddChild(transcript)
	if c.modelMenuOpen {
		root.AddChild(c.renderModelMenu(contentWidth))
	}

	if len(widgetLines) > 0 {
		root.AddChild(c.renderLineBlock(widgetLines, gotui.NewStyle().Foreground(gotui.Blue)))
	}

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

func abbreviateHomePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	home = strings.TrimSpace(home)
	if home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	prefix := home + string(os.PathSeparator)
	if strings.HasPrefix(path, prefix) {
		return "~" + string(os.PathSeparator) + strings.TrimPrefix(path, prefix)
	}
	return path
}

func (c *chatTUI) footerPathLineForWidth(width int) string {
	workspace := strings.TrimSpace(c.cfg.WorkspaceRoot)
	if workspace == "" {
		workspace = "."
	}
	if branch := c.gitBranchName(workspace); branch != "" {
		workspace += " (" + branch + ")"
	}
	workspace = abbreviateHomePath(workspace)
	if width > 0 {
		return compactMaybe(workspace, true, width)
	}
	return workspace
}

func (c *chatTUI) footerModelText(data tuiContextSummary) string {
	model := strings.TrimSpace(data.model)
	if model == "" {
		model = strings.TrimSpace(c.cfg.DefaultModel)
	}
	if model == "" {
		model = "model unset"
	}
	if thinking := strings.TrimSpace(data.thinking); thinking != "" {
		model += " • " + thinking
	}
	return model
}

// footerLines builds a PiSwift-style multi-line footer: a path/branch line, a
// stats line (counts + token usage on the left, model/thinking/context on the
// right), and an optional transient notification line. The bottom band may grow
// to several lines but never adds top chrome.
func (c *chatTUI) footerLines(width int) []string {
	data := c.contextSummaryData()
	lines := []string{c.footerPathLineForWidth(width)}

	statsParts := []string{c.footerCountsText(data)}
	if data.inputTokens > 0 {
		statsParts = append(statsParts, "↑"+formatTokenCount(data.inputTokens))
	}
	if data.outputTokens > 0 {
		statsParts = append(statsParts, "↓"+formatTokenCount(data.outputTokens))
	}
	if data.cacheRead > 0 {
		statsParts = append(statsParts, "R"+formatTokenCount(data.cacheRead))
	}
	if data.cacheWrite > 0 {
		statsParts = append(statsParts, "W"+formatTokenCount(data.cacheWrite))
	}
	if data.costTotal > 0 {
		statsParts = append(statsParts, fmt.Sprintf("$%.3f", data.costTotal))
	}
	if data.contextTokens > 0 {
		statsParts = append(statsParts, formatContextUsage(data.contextTokens, data.contextWindow))
	}
	statsLeft := strings.Join(statsParts, " · ")
	model := c.footerModelText(data)
	lines = append(lines, joinFooterRow(statsLeft, model, width))

	if note := c.footerTransientNotice(data); note != "" {
		if width > 0 {
			note = compactMaybe(note, true, width)
		}
		lines = append(lines, note)
	}
	for _, status := range c.extensionStatusLines() {
		if width > 0 {
			status = compactMaybe(status, true, width)
		}
		lines = append(lines, status)
	}
	return lines
}

// setExtensionStatus is a backend-safe TUI extension slot: extensions can set a
// keyed status segment that renders as an extra dim footer line. It can never
// add top chrome; cleared keys (empty text) are removed.
func (c *chatTUI) setExtensionStatus(key, text string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	if c.extensionStatuses == nil {
		c.extensionStatuses = map[string]string{}
	}
	text = strings.TrimSpace(sanitizeStatusText(text))
	if text == "" {
		delete(c.extensionStatuses, key)
		return
	}
	c.extensionStatuses[key] = text
}

func (c *chatTUI) extensionStatusLines() []string {
	if len(c.extensionStatuses) == 0 {
		return nil
	}
	keys := make([]string, 0, len(c.extensionStatuses))
	for k := range c.extensionStatuses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, c.extensionStatuses[k])
	}
	return lines
}

func sanitizeStatusText(text string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}

// setExtensionWidget is a TUI extension slot that renders a keyed multi-line
// widget between the transcript and the editor. It stays inside the bottom band
// (above the editor, below the transcript) and can never add top chrome.
func (c *chatTUI) setExtensionWidget(key string, lines []string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	if c.extensionWidgets == nil {
		c.extensionWidgets = map[string][]string{}
	}
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned = append(cleaned, sanitizeStatusText(line))
	}
	if len(cleaned) == 0 {
		delete(c.extensionWidgets, key)
		return
	}
	c.extensionWidgets[key] = cleaned
}

func (c *chatTUI) extensionWidgetLines() []string {
	if len(c.extensionWidgets) == 0 {
		return nil
	}
	keys := make([]string, 0, len(c.extensionWidgets))
	for k := range c.extensionWidgets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var lines []string
	for _, k := range keys {
		lines = append(lines, c.extensionWidgets[k]...)
	}
	return lines
}

// setExtensionToolRender is a custom tool-renderer slot: extensions can choose
// how a named tool's block body renders without writing top chrome. Supported
// modes: "full" (default), "compact" (first body line), "hidden" (header only).
func (c *chatTUI) setExtensionToolRender(tool, mode string) {
	tool = strings.TrimSpace(tool)
	if tool == "" {
		return
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if c.extensionToolModes == nil {
		c.extensionToolModes = map[string]string{}
	}
	switch mode {
	case "", "full", "default":
		delete(c.extensionToolModes, tool)
	case "compact", "hidden":
		c.extensionToolModes[tool] = mode
	default:
		delete(c.extensionToolModes, tool)
	}
}

func (c *chatTUI) applyToolRenderMode(tool string, body []string) ([]string, bool) {
	mode := c.extensionToolModes[strings.TrimSpace(tool)]
	switch mode {
	case "hidden":
		return nil, false
	case "compact":
		if len(body) > 1 {
			return body[:1], false
		}
		return body, false
	}
	return body, len(body) > 2
}

func widgetPayloadLines(payload map[string]any) []string {
	switch v := payload["lines"].(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	if text, ok := payload["text"].(string); ok && text != "" {
		return strings.Split(text, "\n")
	}
	return nil
}

func (c *chatTUI) footerTransientNotice(data tuiContextSummary) string {
	status := strings.TrimSpace(c.status)
	if status == "" {
		return ""
	}
	if strings.Contains(status, c.cfg.DefaultModel) || strings.Contains(status, data.model) {
		return ""
	}
	return "» " + status
}

func joinFooterRow(left, right string, width int) string {
	if width <= 0 {
		return compactMaybe(left, true, 24) + "  " + compactMaybe(right, true, 36)
	}
	if len(left)+len(right)+2 >= width {
		rightWidth := 36
		if rightWidth > width/2 {
			rightWidth = width / 2
		}
		if rightWidth < 12 {
			rightWidth = 12
		}
		leftWidth := width - rightWidth - 2
		if leftWidth < 8 {
			leftWidth = 8
		}
		return compactMaybe(left, true, leftWidth) + "  " + compactMaybe(right, true, rightWidth)
	}
	return left + strings.Repeat(" ", width-len(left)-len(right)) + right
}

func (c *chatTUI) footerStatusLineForWidth(width int) string {
	data := c.contextSummaryData()
	left := c.footerNotificationText(data)
	model := strings.TrimSpace(data.model)
	if model == "" {
		model = strings.TrimSpace(c.cfg.DefaultModel)
	}
	if model == "" {
		model = "model unset"
	}
	thinking := strings.TrimSpace(data.thinking)
	if thinking != "" {
		model += " • " + thinking
	}
	if data.contextTokens > 0 {
		model += " • " + formatContextUsage(data.contextTokens, data.contextWindow)
	}
	return joinFooterRow(left, model, width)
}

func formatTokenCount(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%dK", n/1000)
	}
	return strconv.Itoa(n)
}

func formatContextUsage(tokens, window int) string {
	if tokens <= 0 {
		return "ctx 0"
	}
	if window > 0 {
		pct := int(float64(tokens) * 100 / float64(window))
		if pct < 0 {
			pct = 0
		}
		return fmt.Sprintf("ctx %s/%s %d%%", formatTokenCount(tokens), formatTokenCount(window), pct)
	}
	return fmt.Sprintf("ctx %s", formatTokenCount(tokens))
}

func (c *chatTUI) footerCountsText(data tuiContextSummary) string {
	left := fmt.Sprintf("m%d/t%d", data.messageCount, data.turnCount)
	if data.queuedTurns > 0 || data.steeringDepth > 0 {
		left += fmt.Sprintf(" q%d/s%d", data.queuedTurns, data.steeringDepth)
	}
	return left
}

func (c *chatTUI) footerNotificationText(data tuiContextSummary) string {
	counts := c.footerCountsText(data)
	status := strings.TrimSpace(c.status)
	if status != "" && !strings.Contains(status, c.cfg.DefaultModel) && !strings.Contains(status, data.model) {
		return counts + " · " + status
	}
	return counts
}

func (c *chatTUI) footerTextForWidth(width int) string {
	return c.footerStatusLineForWidth(width)
}

func (c *chatTUI) footerText() string {
	return c.footerTextForWidth(c.currentContentWidth())
}

func (c *chatTUI) gitBranchName(workspace string) string {
	if workspace == "" {
		return ""
	}
	headPath := filepath.Join(workspace, ".git", "HEAD")
	raw, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	head := strings.TrimSpace(string(raw))
	const prefix = "ref: refs/heads/"
	if strings.HasPrefix(head, prefix) {
		return strings.TrimPrefix(head, prefix)
	}
	if len(head) >= 7 {
		return head[:7]
	}
	return head
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
		block.AddChild(c.renderInlineStyledLine(line, style))
	}
	return block
}

type tuiInlineSegment struct {
	Text string
	Code bool
}

func parseTUIInlineSegments(line string) []tuiInlineSegment {
	if !strings.Contains(line, markdownInlineCodeStart) {
		return []tuiInlineSegment{{Text: stripMarkdownInlineStyleMarkers(line)}}
	}
	segments := []tuiInlineSegment{}
	for len(line) > 0 {
		start := strings.Index(line, markdownInlineCodeStart)
		if start < 0 {
			if line != "" {
				segments = append(segments, tuiInlineSegment{Text: stripMarkdownInlineStyleMarkers(line)})
			}
			break
		}
		if start > 0 {
			segments = append(segments, tuiInlineSegment{Text: stripMarkdownInlineStyleMarkers(line[:start])})
		}
		line = line[start+len(markdownInlineCodeStart):]
		end := strings.Index(line, markdownInlineCodeEnd)
		if end < 0 {
			segments = append(segments, tuiInlineSegment{Text: stripMarkdownInlineStyleMarkers(line)})
			break
		}
		segments = append(segments, tuiInlineSegment{Text: line[:end], Code: true})
		line = line[end+len(markdownInlineCodeEnd):]
	}
	if len(segments) == 0 {
		return []tuiInlineSegment{{Text: ""}}
	}
	return segments
}

func (c *chatTUI) renderInlineStyledLine(line string, style gotui.Style) *gotui.Element {
	segments := parseTUIInlineSegments(line)
	if len(segments) == 1 && !segments[0].Code {
		return gotui.New(
			gotui.WithWidthPercent(100),
			gotui.WithHeight(1),
			gotui.WithText(segments[0].Text),
			gotui.WithTextStyle(style),
		)
	}
	row := gotui.New(
		gotui.WithDirection(gotui.Row),
		gotui.WithWidthPercent(100),
		gotui.WithHeight(1),
	)
	for _, seg := range segments {
		if seg.Text == "" {
			continue
		}
		segStyle := style
		if seg.Code {
			segStyle = gotui.NewStyle().Foreground(gotui.BrightBlack).Dim()
		}
		row.AddChild(gotui.New(
			gotui.WithText(seg.Text),
			gotui.WithTextStyle(segStyle),
		))
	}
	return row
}

func (c *chatTUI) buildTranscriptRenderableBlocks(lines []string) []transcriptRenderableBlock {
	blocks := make([]transcriptRenderableBlock, 0, len(lines))
	seenErrorBlocks := map[string]bool{}
	c.ensureTranscriptBlockState()
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if meta, ok := parseTranscriptBlockMarker(line); ok {
			body := make([]string, 0, 4)
			j := i + 1
			for j < len(lines) && strings.HasPrefix(lines[j], "│ ") {
				body = append(body, strings.TrimPrefix(lines[j], "│ "))
				j++
			}
			expanded := c.transcriptExpanded[meta.Key]
			headStyle, bodyStyle, hintStyle, borderStyle, border := transcriptBlockPalette(meta.Kind, meta.Status, c.selectedTranscriptBlock == meta.Key)
			header := meta.Title
			subheader := ""
			if meta.Kind == "tool" {
				if elapsed := formatBlockElapsed(meta.StartedAt, meta.EndedAt); elapsed != "" {
					header += " · " + elapsed
				}
				if detail := strings.TrimSpace(meta.Detail); detail != "" {
					header += " · " + detail
				}
			} else if meta.Kind == "thinking" || meta.Kind == "thinking_indicator" {
				subheader = ""
			} else {
				timeBits := []string{}
				if clock := formatBlockClock(meta.StartedAt); clock != "" {
					timeBits = append(timeBits, clock)
				}
				if strings.TrimSpace(meta.EndedAt) != "" {
					if elapsed := formatBlockElapsed(meta.StartedAt, meta.EndedAt); elapsed != "" {
						timeBits = append(timeBits, elapsed)
					}
				}
				if meta.Detail != "" {
					timeBits = append(timeBits, meta.Detail)
				}
				subheader = strings.Join(timeBits, " · ")
			}
			expandable := len(body) > 2
			selectedHint := borderStyle
			previewLimit := 0
			previewTail := false
			if meta.Kind == "thinking" || meta.Kind == "thinking_indicator" {
				expandable = false
				selectedHint = ""
			}
			if meta.Kind == "tool" {
				body, expandable = c.applyToolRenderMode(meta.Title, body)
			}
			if meta.Kind == "bash" {
				previewLimit = bashPreviewLines
				previewTail = true
				expandable = len(body) > bashPreviewLines
			}
			blocks = append(blocks, transcriptRenderableBlock{Key: meta.Key, Kind: meta.Kind, Header: header, Subheader: subheader, Body: body, Expandable: expandable, Expanded: expanded, PreviewLimit: previewLimit, PreviewTail: previewTail, Footer: strings.TrimSpace(meta.Footer), Status: meta.Status, Selected: c.selectedTranscriptBlock == meta.Key, Border: gotui.BorderRounded, BorderStyle: border, HeaderStyle: headStyle, BodyStyle: bodyStyle, HintStyle: hintStyle, SelectedHint: selectedHint})
			i = j - 1
			continue
		}
		switch {
		case strings.HasPrefix(line, "local$ "):
			body := make([]string, 0, 4)
			j := i + 1
			for j < len(lines) {
				if strings.HasPrefix(lines[j], "│ ") {
					body = append(body, strings.TrimPrefix(lines[j], "│ "))
					j++
					continue
				}
				if strings.HasPrefix(lines[j], "error:") {
					body = append(body, lines[j])
					j++
					continue
				}
				break
			}
			key := fmt.Sprintf("local:%d:%s", i, line)
			headStyle, bodyStyle, hintStyle, borderStyle, border := transcriptBlockPalette("local", "info", c.selectedTranscriptBlock == key)
			blocks = append(blocks, transcriptRenderableBlock{Key: key, Kind: "local", Header: line, Body: body, Expandable: len(body) > 2, Expanded: c.transcriptExpanded[key], Selected: c.selectedTranscriptBlock == key, Border: gotui.BorderRounded, BorderStyle: border, HeaderStyle: headStyle, BodyStyle: bodyStyle, HintStyle: hintStyle, SelectedHint: borderStyle})
			i = j - 1
		case isTUIErrorLine(line):
			key := tuiErrorDedupKey(line)
			if seenErrorBlocks[key] {
				continue
			}
			seenErrorBlocks[key] = true
			body := []string{trimTUIErrorLine(line)}
			blockKey := "error:" + key
			headStyle, bodyStyle, hintStyle, selectedHint, borderStyle := transcriptBlockPalette("error", "error", c.selectedTranscriptBlock == blockKey)
			blocks = append(blocks, transcriptRenderableBlock{Key: blockKey, Kind: "error", Header: "Error", Body: body, Selected: c.selectedTranscriptBlock == blockKey, Border: gotui.BorderRounded, BorderStyle: borderStyle, HeaderStyle: headStyle, BodyStyle: bodyStyle, HintStyle: hintStyle, SelectedHint: selectedHint})
		default:
			kind := "plain"
			switch {
			case strings.HasPrefix(line, "sys:"):
				kind = "system"
			case strings.HasPrefix(line, "you"):
				kind = "user"
			case strings.HasPrefix(line, c.cfg.AssistantName+":"):
				kind = "assistant"
			}
			headStyle, bodyStyle, hintStyle, _, _ := transcriptBlockPalette(kind, "", false)
			blocks = append(blocks, transcriptRenderableBlock{Key: fmt.Sprintf("%s:%d:%s", kind, i, line), Kind: kind, Header: line, HeaderStyle: headStyle, BodyStyle: bodyStyle, HintStyle: hintStyle})
		}
	}
	return blocks
}

func shouldRenderTranscriptStatusLabel(kind, status string) bool {
	if status == "" {
		return false
	}
	if (kind == "tool" || kind == "bash") && status == "ok" {
		return false
	}
	return true
}

func isTUIErrorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "error:") {
		return true
	}
	if strings.HasPrefix(lower, "sys:") {
		bodyLower := strings.ToLower(strings.TrimSpace(trimmed[4:]))
		return strings.Contains(bodyLower, "inference error") || strings.Contains(bodyLower, "max retries exceeded")
	}
	return strings.Contains(lower, "max retries exceeded")
}

func trimTUIErrorLine(line string) string {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "sys:"):
		return strings.TrimSpace(trimmed[4:])
	case strings.HasPrefix(lower, "error:"):
		msg := strings.TrimSpace(trimmed[6:])
		if msg != "" {
			return msg
		}
	}
	return trimmed
}

func tuiErrorDedupKey(line string) string {
	msg := strings.ToLower(trimTUIErrorLine(line))
	msg = strings.Join(strings.Fields(msg), " ")
	if strings.Contains(msg, "max retries exceeded") {
		return "max retries exceeded"
	}
	if strings.HasPrefix(msg, "inference error:") {
		msg = strings.TrimSpace(strings.TrimPrefix(msg, "inference error:"))
	}
	if msg == "" {
		return "error"
	}
	return msg
}

func transcriptBlockPalette(kind, status string, selected bool) (gotui.Style, gotui.Style, gotui.Style, string, gotui.Style) {
	fg := gotui.White
	switch kind {
	case "tool":
		fg = gotui.Cyan
	case "thought":
		fg = gotui.Magenta
	case "hook", "route", "dispatcher", "subturn":
		fg = gotui.Blue
	case "local":
		fg = gotui.Magenta
	case "bash":
		fg = gotui.Cyan
	case "compact":
		fg = gotui.Yellow
	case "error":
		fg = gotui.Red
	case "user":
		fg = gotui.Green
	}
	switch status {
	case "error":
		fg = gotui.Red
	case "ok":
		fg = gotui.Green
	case "running":
		fg = gotui.Cyan
	case "skipped":
		fg = gotui.Yellow
	}
	head := gotui.NewStyle().Foreground(fg).Bold()
	body := gotui.NewStyle().Foreground(fg)
	hint := gotui.NewStyle().Dim()
	border := gotui.NewStyle().Foreground(fg)
	selectedHint := "F6/F7 select · F8 toggle · click to expand"
	if selected {
		border = border.Bold()
		selectedHint = "selected · F6/F7 move · F8 toggle · click to expand"
	}
	if kind == "system" || kind == "plain" || kind == "assistant" {
		head = gotui.NewStyle()
		body = gotui.NewStyle()
		hint = gotui.NewStyle().Dim()
		border = gotui.NewStyle().Foreground(gotui.BrightBlack)
		if kind == "assistant" {
			head = gotui.NewStyle().Bold()
		}
		if kind == "system" {
			head = gotui.NewStyle().Dim()
			body = gotui.NewStyle().Dim()
		}
	}
	return head, body, hint, selectedHint, border
}

func brailleSpinnerFrame(t time.Time) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if t.IsZero() {
		t = time.Now()
	}
	idx := int((t.UnixMilli() / 120) % int64(len(frames)))
	if idx < 0 {
		idx = 0
	}
	return frames[idx]
}

func (c *chatTUI) renderTranscriptBlock(block transcriptRenderableBlock) *gotui.Element {
	if block.Kind == "thought" {
		container := gotui.New(
			gotui.WithDirection(gotui.Column),
			gotui.WithWidthPercent(100),
			gotui.WithBorder(block.Border),
			gotui.WithBorderStyle(block.BorderStyle),
			gotui.WithPaddingTRBL(0, 1, 0, 1),
		)
		ref := gotui.NewRef()
		ref.Set(container)
		c.transcriptBlockRefs = append(c.transcriptBlockRefs, transcriptBlockHitTarget{Key: block.Key, Ref: ref})
		if len(block.Body) == 0 {
			container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(fmt.Sprintf("%s Thinking...", brailleSpinnerFrame(time.Now()))), gotui.WithTextStyle(block.BodyStyle)))
			return container
		}
		for _, line := range block.Body {
			container.AddChild(c.renderInlineStyledLine(line, block.BodyStyle))
		}
		return container
	}
	if len(block.Body) == 0 && block.Subheader == "" && block.Kind != "local" {
		headerText := block.Header
		if block.Status == "running" {
			headerText = fmt.Sprintf("%s %s", brailleSpinnerFrame(time.Now()), headerText)
		} else if shouldRenderTranscriptStatusLabel(block.Kind, strings.TrimSpace(block.Status)) {
			headerText = fmt.Sprintf("%s [%s]", headerText, block.Status)
		}
		return c.renderInlineStyledLine(headerText, block.HeaderStyle)
	}
	container := gotui.New(
		gotui.WithDirection(gotui.Column),
		gotui.WithWidthPercent(100),
		gotui.WithBorder(block.Border),
		gotui.WithBorderStyle(block.BorderStyle),
		gotui.WithPaddingTRBL(0, 1, 0, 1),
	)
	ref := gotui.NewRef()
	ref.Set(container)
	c.transcriptBlockRefs = append(c.transcriptBlockRefs, transcriptBlockHitTarget{Key: block.Key, Ref: ref})
	statusLabel := strings.TrimSpace(block.Status)
	headerText := block.Header
	if statusLabel == "running" {
		headerText = fmt.Sprintf("%s %s", brailleSpinnerFrame(time.Now()), headerText)
	} else if shouldRenderTranscriptStatusLabel(block.Kind, statusLabel) {
		headerText = fmt.Sprintf("%s [%s]", headerText, statusLabel)
	}
	container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(headerText), gotui.WithTextStyle(block.HeaderStyle)))
	if block.Subheader != "" {
		container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(block.Subheader), gotui.WithTextStyle(block.HintStyle)))
	}
	visibleBody := block.Body
	hiddenCount := 0
	previewLimit := 2
	if block.PreviewLimit > 0 {
		previewLimit = block.PreviewLimit
	}
	if block.Expandable && !block.Expanded {
		if len(visibleBody) > previewLimit {
			hiddenCount = len(visibleBody) - previewLimit
			if block.PreviewTail {
				visibleBody = visibleBody[len(visibleBody)-previewLimit:]
			} else {
				visibleBody = visibleBody[:previewLimit]
			}
		}
	}
	for _, line := range visibleBody {
		container.AddChild(c.renderInlineStyledLine(line, block.BodyStyle))
	}
	if block.Expandable && hiddenCount > 0 {
		hint := fmt.Sprintf("… %d more line(s) · F8 expand", hiddenCount)
		container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(hint), gotui.WithTextStyle(block.HintStyle)))
	} else if block.Expandable && block.Expanded && block.PreviewTail {
		container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText("F8 collapse"), gotui.WithTextStyle(block.HintStyle)))
	}
	if strings.TrimSpace(block.Footer) != "" {
		container.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithText(block.Footer), gotui.WithTextStyle(block.HintStyle)))
	}
	return container
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

func (c *chatTUI) renderToolCallSummaryLines(m store.Message, width int) []string {
	return nil
}

func (c *chatTUI) renderMessageLines(m store.Message, width int) []string {
	kind, _ := m.Payload["kind"].(string)
	if m.Role == "tool_result" || kind == "tool_result" {
		return c.renderToolResultLines(m)
	}
	if kind == "tool_calls" {
		return c.renderToolCallSummaryLines(m, width)
	}
	if kind == "compaction" {
		return c.renderCompactionMessageLines(m)
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

func (c *chatTUI) renderToolResultLines(m store.Message) []string {
	toolName, _ := m.Payload["tool_name"].(string)
	if toolName == "" {
		toolName = "tool"
	}
	isErr, _ := m.Payload["is_error"].(bool)
	status := "ok"
	if isErr {
		status = "error"
	}
	trimmed := strings.TrimSpace(m.Content)
	meta := transcriptBlockMeta{Key: "msg:" + m.ID, Kind: "tool", Title: toolName, Status: status, StartedAt: strings.TrimSpace(m.CreatedAt), EndedAt: strings.TrimSpace(m.CreatedAt)}
	if trimmed == "" {
		return []string{encodeTranscriptBlockMarker(meta), "│ (empty)"}
	}
	parts := strings.Split(trimmed, "\n")
	lines := []string{encodeTranscriptBlockMarker(meta)}
	if len(parts) == 1 {
		lines = append(lines, "│ "+truncate(strings.Join(strings.Fields(parts[0]), " "), 200))
		return lines
	}
	for _, part := range parts {
		lines = append(lines, "│ "+truncate(strings.TrimRight(part, "\r"), 200))
	}
	return lines
}

func (c *chatTUI) renderCompactionMessageLines(m store.Message) []string {
	tokens := toInt(m.Payload["tokens_before"], 0)
	return []string{encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "msg:" + m.ID, Kind: "compact", Title: "Context compacted", Status: "ok", StartedAt: strings.TrimSpace(m.CreatedAt), EndedAt: strings.TrimSpace(m.CreatedAt)}), fmt.Sprintf("│ summary=%s", truncate(m.Content, 160)), fmt.Sprintf("│ tokens_before=%d", tokens)}
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
	contextTokens int
	contextWindow int
	inputTokens   int
	outputTokens  int
	cacheRead     int
	cacheWrite    int
	costTotal     float64
}

func (c *chatTUI) latestUsageTokens(turns []store.Turn) (input, output, total int) {
	if c.store == nil || len(turns) == 0 {
		return 0, 0, 0
	}
	ctx := context.Background()
	for i := len(turns) - 1; i >= 0; i-- {
		events, err := c.store.ListTurnEvents(ctx, turns[i].ID)
		if err != nil {
			continue
		}
		for j := len(events) - 1; j >= 0; j-- {
			usage, _ := events[j].Payload["usage"].(map[string]any)
			if usage == nil {
				continue
			}
			input = intFromAny(usage["input"])
			output = intFromAny(usage["output"])
			total = intFromAny(usage["total"])
			if total == 0 {
				total = intFromAny(usage["totalTokens"])
			}
			return input, output, total
		}
	}
	return 0, 0, 0
}

func (c *chatTUI) applyModelContextWindow(data *tuiContextSummary) {
	if data == nil {
		return
	}
	if contextWindow := inference.ResolveModelContextWindow(data.provider, data.model); contextWindow > 0 {
		data.contextWindow = contextWindow
	}
}

func (c *chatTUI) contextSummaryData() tuiContextSummary {
	data := tuiContextSummary{sessionTitle: c.sessionID, agentID: "agent", parent: "root", model: c.cfg.DefaultModel, provider: c.cfg.DefaultProvider, thinking: c.cfg.DefaultThinkingLevel, status: "idle", contextWindow: c.cfg.Compaction.ContextWindow, contextTokens: c.lastContextTokens, inputTokens: c.lastInputTokens, outputTokens: c.lastOutputTokens, cacheRead: c.lastCacheRead, cacheWrite: c.lastCacheWrite, costTotal: c.lastCostTotal}
	c.applyModelContextWindow(&data)
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
	if data.contextTokens == 0 {
		data.inputTokens, data.outputTokens, data.contextTokens = c.latestUsageTokens(turns)
	}
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
	c.applyModelContextWindow(&data)
	return data
}

func (c *chatTUI) contextSummaryLines(width int) []string {
	data := c.contextSummaryData()
	line := fmt.Sprintf("@%s · %s · %s · m%d/t%d", data.agentID, data.model, data.thinking, data.messageCount, data.turnCount)
	if data.contextTokens > 0 {
		line += " · " + formatContextUsage(data.contextTokens, data.contextWindow)
	}
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
	dbPath := flag.String("db", config.DefaultTUIDBPath(), "SQLite database path")
	workspace := flag.String("workspace", config.DefaultWorkspaceRoot(), "Workspace root")
	model := flag.String("model", "", "Override default model")
	flag.Parse()
	if err := Run(*dbPath, *workspace, *model); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
