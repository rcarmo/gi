package tui

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

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

	sessions, _ := s.ListSessions(context.Background())
	var sessionID string
	if len(sessions) > 0 {
		sessionID = sessions[0].ID
	} else {
		id := store.NowID("session")
		alloc := gisession.AllocateDefaultSession("agent", "gi", "default", id)
		sess, err := s.CreateSessionWithMetadata(context.Background(), id, "", "@agent", map[string]any{"status": "idle", "queue_count": 0, "model": cfg.DefaultModel, "provider": cfg.DefaultProvider, "thinking_level": cfg.DefaultThinkingLevel}, &alloc.Scope, alloc.SessionAliases)
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
	subscribedCh     chan map[string]any
	input            *gotui.Input
	inputRegion      *gotui.Element
	transcriptRegion *gotui.Element
	transcriptRef    *gotui.Ref
	transcript       []string
	transcriptScroll int
	stickToBottom    bool
}

func (c *chatTUI) ensureInput() {
	if c.input != nil {
		return
	}
	c.input = gotui.NewInput(
		gotui.WithInputPlaceholder("Send a message…"),
		gotui.WithInputAutoFocus(true),
		gotui.WithInputWidth(80),
		gotui.WithInputBorder(gotui.BorderRounded),
		gotui.WithInputOnSubmit(c.onSubmit),
	)
}

func (c *chatTUI) Init() func() {
	c.eventCh = make(chan map[string]any, 64)
	c.bindSession(c.sessionID)
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.histIdx = -1
	c.history = []string{}
	c.inputActive = true
	c.transcript = c.loadTranscript()
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
	}
}

func (c *chatTUI) bindSession(sessionID string) {
	if c.subscribedCh != nil {
		c.engine.Unsubscribe(c.sessionID, c.subscribedCh)
		c.subscribedCh = nil
	}
	c.sessionID = sessionID
	c.subscribedCh = c.engine.Subscribe(sessionID)
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
	return []gotui.Watcher{
		gotui.NewChannelWatcher(c.eventCh, c.handleEvent),
	}
}

func (c *chatTUI) handleEvent(ev map[string]any) {
	evType, _ := ev["type"].(string)
	switch evType {
	case "agent_draft_delta":
		delta, _ := ev["delta"].(string)
		c.draft += delta
		c.status = fmt.Sprintf("⏳ %s…", truncate(c.draft, 80))
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		c.app.MarkDirty()
	case "new_post":
		data, _ := ev["data"].(map[string]any)
		if data != nil {
			text, _ := data["content"].(string)
			if text != "" {
				c.transcript = append(c.transcript, fmt.Sprintf("%s: %s", c.cfg.AssistantName, text))
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				c.running = false
				c.draft = ""
				if c.stickToBottom {
					c.scrollTranscriptToBottom()
				}
				c.app.MarkDirty()
			}
		}
	case "agent_thought_delta":
		c.status = "Thinking…"
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "tool_finished":
		toolName, _ := ev["tool"].(string)
		if toolName != "" {
			c.status = fmt.Sprintf("Tool finished: %s", toolName)
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "tool_failed":
		toolName, _ := ev["tool"].(string)
		errText, _ := ev["error"].(string)
		line := fmt.Sprintf("sys: tool failed: %s", toolName)
		if errText != "" {
			line = fmt.Sprintf("%s: %s", line, truncate(errText, 120))
		}
		c.transcript = append(c.transcript, line)
		c.status = fmt.Sprintf("Tool failed: %s", toolName)
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "compaction":
		before := intFromEvent(ev, "messages_before")
		after := intFromEvent(ev, "messages_after")
		tokens := intFromEvent(ev, "tokens_before")
		c.transcript = append(c.transcript, fmt.Sprintf("sys: compacted context: messages %d→%d, tokens_before=%d", before, after, tokens))
		c.status = "Compacted context"
		if c.stickToBottom {
			c.scrollTranscriptToBottom()
		}
		if c.app != nil {
			c.app.MarkDirty()
		}
	case "agent_status":
		title, _ := ev["title"].(string)
		if title != "" {
			c.status = title
		} else {
			c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
		}
		c.app.MarkDirty()
	case "error":
		msg, _ := ev["error"].(string)
		if msg == "" {
			msg, _ = ev["message"].(string)
		}
		if msg != "" {
			c.transcript = append(c.transcript, "error: "+truncate(msg, 160))
			c.status = "Error"
			if c.app != nil {
				c.app.MarkDirty()
			}
		}
	}
}

func intFromEvent(ev map[string]any, key string) int {
	switch v := ev[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
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
	c.running = true
	c.status = "sending…"
	c.draft = ""
	c.transcript = append(c.transcript, fmt.Sprintf("you: %s", text))
	c.stickToBottom = true
	c.scrollTranscriptToBottom()
	c.app.MarkDirty()

	go func() {
		result, err := c.engine.SubmitPromptRouted(context.Background(), turn.RunInput{
			SessionID: c.sessionID,
			Prompt:    text,
			Intent:    "prompt",
			Model:     c.cfg.DefaultModel,
		})
		if err != nil {
			c.app.QueueUpdate(func() {
				c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
				c.running = false
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				c.app.MarkDirty()
			})
			return
		}
		if result != nil && result.Routed && result.SessionID != c.sessionID {
			c.app.QueueUpdate(func() {
				c.transcript = append(c.transcript, fmt.Sprintf("sys: routed to @%s (%s)", result.TargetAgentID, result.SessionID))
				c.running = false
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				c.app.MarkDirty()
			})
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
		c.transcript = append(c.transcript, c.compactLines()...)
	case "/cancel":
		c.transcript = append(c.transcript, c.cancelCommand())
	case "/agents":
		c.transcript = append(c.transcript, c.listAgentLines()...)
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
				c.app.QueueUpdate(func() {
					c.transcript = append(c.transcript, fmt.Sprintf("error: %v", err))
					c.running = false
					c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
					c.app.MarkDirty()
				})
				return
			}
			c.app.QueueUpdate(func() {
				c.transcript = append(c.transcript, fmt.Sprintf("sys: delivered to @%s (%s)", target, result.SessionID))
				c.running = false
				c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
				c.app.MarkDirty()
			})
		}()
		return
	case "/where":
		c.transcript = append(c.transcript, c.contextSummary())
	default:
		c.transcript = append(c.transcript, "sys: commands: /help, /tools [query|active|activate|reset], /skills [query], /model [name], /thinking [level], /compact, /cancel, /agents, /fork [@agentN], /switch @agent|session_id, /send @agent message, /where")
	}
	c.running = false
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.stickToBottom = true
	c.scrollTranscriptToBottom()
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) helpLines() []string {
	return []string{
		"sys: gi TUI help",
		"sys: keys: Enter send · Esc blur input · Ctrl-C/Ctrl-D quit · Up/Down history · PgUp/PgDn scroll · Home/End transcript",
		"sys: commands: /help, /tools [query|active|activate|reset], /skills [query], /model [name], /thinking [level], /compact, /cancel, /agents, /where",
		"sys: sessions: /fork [@agentN], /switch @agent|session_id, /send @agent message",
	}
}

func (c *chatTUI) modelCommand(fields []string) []string {
	if len(fields) == 1 {
		return []string{fmt.Sprintf("sys: model: %s", c.cfg.DefaultModel)}
	}
	model := strings.TrimSpace(strings.Join(fields[1:], " "))
	if model == "" {
		return []string{"sys: usage /model <model>"}
	}
	c.cfg.DefaultModel = model
	_ = c.store.TouchSessionState(context.Background(), c.sessionID, map[string]any{"model": model})
	return []string{fmt.Sprintf("sys: model set to %s", model)}
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
	_ = c.store.TouchSessionState(context.Background(), c.sessionID, map[string]any{"thinking_level": level})
	return []string{fmt.Sprintf("sys: thinking set to %s", level)}
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

func (c *chatTUI) switchSession(sessionID string) {
	c.bindSession(sessionID)
	c.transcript = c.loadTranscript()
	c.draft = ""
	c.running = false
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.scrollTranscriptToBottom()
}

func (c *chatTUI) listAgentLines() []string {
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	if len(sessions) == 0 {
		return []string{"sys: no sessions"}
	}
	lines := []string{"sys: agents:"}
	for _, sess := range sessions {
		marker := " "
		if sess.ID == c.sessionID {
			marker = "*"
		}
		parent := ""
		if sess.ParentSessionID != "" {
			parent = fmt.Sprintf(" parent=%s", sess.ParentSessionID)
		}
		lines = append(lines, fmt.Sprintf("%s @%s %s%s", marker, c.agentIDForSession(&sess), sess.ID, parent))
	}
	return lines
}

func (c *chatTUI) resolveSessionRef(ref string) (*store.Session, error) {
	ref = strings.TrimSpace(strings.TrimPrefix(ref, "@"))
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		return nil, err
	}
	for i := range sessions {
		if sessions[i].ID == ref || c.agentIDForSession(&sessions[i]) == ref {
			return &sessions[i], nil
		}
	}
	return nil, fmt.Errorf("unknown session or agent: %s", ref)
}

func (c *chatTUI) nextForkAgentID() string {
	current, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		return "agent1"
	}
	base := strings.TrimRight(c.agentIDForSession(current), "0123456789")
	if base == "" {
		base = c.agentIDForSession(current)
	}
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		return base + "1"
	}
	used := map[string]bool{}
	for _, sess := range sessions {
		used[c.agentIDForSession(&sess)] = true
	}
	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s%d", base, i)
		if !used[candidate] {
			return candidate
		}
	}
	return base + "999"
}

func (c *chatTUI) agentIDForSession(sess *store.Session) string {
	if sess != nil && sess.Scope != nil && sess.Scope.AgentID != "" {
		return sess.Scope.AgentID
	}
	return "agent"
}

func (c *chatTUI) Render(app *gotui.App) *gotui.Element {
	root := gotui.New(
		gotui.WithDirection(gotui.Column),
		gotui.WithWidthPercent(100),
		gotui.WithHeightPercent(100),
		gotui.WithPadding(1),
		gotui.WithGap(1),
	)

	statusEl := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithText(fmt.Sprintf("Status: %s", c.status)),
		gotui.WithTextStyle(gotui.NewStyle().Bold()),
	)
	root.AddChild(statusEl)

	ctxEl := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithText(c.contextSummary()),
		gotui.WithTextStyle(gotui.NewStyle().Dim()),
	)
	root.AddChild(ctxEl)

	sep := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithText("─────────────────────────────────────────────────────────────────────────────────"),
		gotui.WithTextStyle(gotui.NewStyle().Dim()),
	)
	root.AddChild(sep)

	_, h := app.Size()
	transcriptHeight := h - 16
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

	inputLabel := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithText("Input:"),
		gotui.WithTextStyle(gotui.NewStyle().Dim()),
	)
	root.AddChild(inputLabel)

	c.ensureInput()
	inputEl := app.MountPersistent(c, 0, func() gotui.Component { return c.input })
	c.inputRegion = inputEl
	root.AddChild(inputEl)

	footer := gotui.New(
		gotui.WithWidthPercent(100),
		gotui.WithText(c.footerText()),
		gotui.WithTextStyle(gotui.NewStyle().Dim()),
	)
	root.AddChild(footer)

	return root
}

func (c *chatTUI) footerText() string {
	return "Hints: /help · /tools active|activate|reset · /skills · /model · /thinking · /compact · /cancel · /agents · /fork · /switch · /send · Esc blur · Tab focus · F2/F3 history · PgUp/PgDn scroll · Ctrl-D quit"
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
		prefix := "you"
		switch m.Role {
		case "assistant":
			prefix = c.cfg.AssistantName
		case "system":
			prefix = "sys"
		}
		out = append(out, fmt.Sprintf("%s: %s", prefix, truncate(m.Content, 200)))
	}
	return out
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
	lines := c.loadTranscript()
	if len(c.transcript) > len(lines) {
		lines = append([]string(nil), c.transcript...)
	}
	if c.draft != "" {
		lines = append(lines, fmt.Sprintf("draft: %s", c.draft))
	}
	if len(lines) == 0 {
		return []string{"(no messages yet)"}
	}
	return lines
}

func (c *chatTUI) contextSummary() string {
	session, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		return fmt.Sprintf("session=%s", c.sessionID)
	}
	messages, _ := c.store.ListMessages(context.Background(), c.sessionID)
	messageCount := len(messages)
	turns, _ := c.store.ListTurns(context.Background(), c.sessionID)
	state := session.State
	model := c.cfg.DefaultModel
	provider := c.cfg.DefaultProvider
	thinking := c.cfg.DefaultThinkingLevel
	status := "idle"
	if v, ok := state["model"].(string); ok && v != "" {
		model = v
	}
	if v, ok := state["provider"].(string); ok && v != "" {
		provider = v
	}
	if v, ok := state["thinking_level"].(string); ok && v != "" {
		thinking = v
	}
	if v, ok := state["status"].(string); ok && v != "" {
		status = v
	}
	agentID := c.agentIDForSession(session)
	parent := "root"
	if session.ParentSessionID != "" {
		parent = session.ParentSessionID
	}
	return fmt.Sprintf("Session: %s · Agent: @%s · Parent: %s · Model: %s · Provider: %s · Thinking: %s · Status: %s · Messages: %d · Turns: %d", session.Title, agentID, parent, model, provider, thinking, status, messageCount, len(turns))
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
