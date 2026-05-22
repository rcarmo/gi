package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func TestFocusInputActivatesInput(t *testing.T) {
	c := &chatTUI{}
	c.focusInput()
	if !c.inputActive {
		t.Fatal("expected inputActive to be true")
	}
}

func TestHandleMouseOnInputRegionFocusesInput(t *testing.T) {
	el := gotui.New(gotui.WithWidth(20), gotui.WithHeight(1))
	buf := gotui.NewBuffer(80, 25)
	el.Render(buf, 80, 25)
	c := &chatTUI{inputRegion: el}
	c.inputActive = false
	consumed := c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 2, Y: 0})
	if !consumed {
		t.Fatal("expected mouse click to be consumed")
	}
	if !c.inputActive {
		t.Fatal("expected inputActive to become true after click")
	}
}

func TestHandleMouseWheelScrollsTranscript(t *testing.T) {
	c := &chatTUI{
		transcript:       []string{"1", "2", "3", "4", "5", "6", "7", "8"},
		transcriptScroll: 3,
	}
	if !c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseWheelUp, Action: gotui.MousePress}) {
		t.Fatal("expected wheel event to be consumed")
	}
	if c.transcriptScroll >= 3 {
		t.Fatalf("expected transcript scroll to move up, got %d", c.transcriptScroll)
	}
}

func TestVisibleTranscriptDoesNotMutateDraftLineIndex(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{ScrollbackLimit: 3}, transcript: []string{"1", "2", "3", "4"}, draftLineIndex: 3}
	_ = c.visibleTranscript()
	if c.draftLineIndex != 3 {
		t.Fatalf("expected visibleTranscript to leave draftLineIndex unchanged, got %d", c.draftLineIndex)
	}
}

func TestBindSessionReusesTopicEventChannel(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_topic_a", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session a: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_topic_b", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session b: %v", err)
	}
	engine := turn.New(s)
	c := &chatTUI{store: s, engine: engine}
	c.bindSession("session_topic_a")
	first := c.topicEventCh
	if first == nil {
		t.Fatal("expected topicEventCh to be initialized")
	}
	c.bindSession("session_topic_b")
	if c.topicEventCh != first {
		t.Fatal("expected bindSession to reuse topicEventCh so watchers stay attached")
	}
}

func TestScrollTranscriptClampsToBounds(t *testing.T) {
	c := &chatTUI{transcript: []string{"1", "2", "3", "4", "5", "6", "7"}}
	c.scrollTranscript(-10)
	if c.transcriptScroll != 0 {
		t.Fatalf("expected top clamp at 0, got %d", c.transcriptScroll)
	}
	c.scrollTranscript(100)
	if c.transcriptScroll != 3 {
		t.Fatalf("expected bottom clamp at 3, got %d", c.transcriptScroll)
	}
}

func TestHandleForkCommandSwitchesSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	root, err := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: root.ID, cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, eventCh: make(chan map[string]any, 64)}
	c.handleCommand("/fork @agent1")
	if c.sessionID == root.ID {
		t.Fatalf("expected fork command to switch session")
	}
	sess, err := s.GetSession(ctx, c.sessionID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sess.Scope == nil || sess.Scope.AgentID != "agent1" {
		t.Fatalf("unexpected forked session: %#v", sess)
	}
}

func TestListAgentLinesReportsAgentIndexErrors(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_error_agents", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	c := &chatTUI{store: s}
	lines := c.listAgentLines()
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "error:") {
		t.Fatalf("expected listAgentLines to surface error line, got %#v", lines)
	}
}

func TestTreeLinesReportsAgentIndexErrors(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_error_tree", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	c := &chatTUI{store: s}
	lines := c.treeLines()
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "error:") {
		t.Fatalf("expected treeLines to surface error line, got %#v", lines)
	}
}

func TestTreeLinesShowsParentChildSessions(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	root, _ := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	child, _ := s.CloneSession(ctx, root.ID, "session_child", "@agent1", "agent1")
	c := &chatTUI{store: s, sessionID: child.ID}
	lines := strings.Join(c.treeLines(), "\n")
	for _, want := range []string{"tree: sessions:", "@gi session_root", "* @agent1 session_child"} {
		if !strings.Contains(lines, want) {
			t.Fatalf("tree missing %q:\n%s", want, lines)
		}
	}
}

func TestResolveSessionRefPropagatesLookupErrors(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	_ = s.Close()
	c := &chatTUI{store: s, sessionID: "session_root"}
	if _, err := c.resolveSessionRef("@agent1"); err == nil {
		t.Fatalf("expected resolve session ref to propagate store lookup errors")
	}
}

func TestNextForkAgentIDFallsBackWhenAgentIndexLookupFails(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_fork_error", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	c := &chatTUI{store: s, sessionID: "session_fork_error"}
	if got := c.nextForkAgentID(); got != "agent1" {
		t.Fatalf("expected deterministic fallback agent1 on index lookup error, got %q", got)
	}
}

func TestResolveSessionRefPrefersDirectSessionIDBeforeAgentLookup(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "agent1", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create direct-id session: %v", err)
	}
	root, err := s.CreateSession(ctx, "session_root_ref_precedence", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create root session: %v", err)
	}
	if _, err := s.CloneSession(ctx, root.ID, "session_child_ref_precedence", "@agent1", "agent1"); err != nil {
		t.Fatalf("create child agent1 session: %v", err)
	}
	c := &chatTUI{store: s, sessionID: root.ID}
	sess, err := c.resolveSessionRef("@agent1")
	if err != nil {
		t.Fatalf("resolve session ref precedence: %v", err)
	}
	if sess.ID != "agent1" {
		t.Fatalf("expected direct-id precedence session agent1, got %#v", sess)
	}
}

func TestResolveSessionRefByAgentIDCaseInsensitive(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	root, _ := s.CreateSession(ctx, "session_root_ci", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	_, _ = s.CloneSession(ctx, root.ID, "session_child_ci", "@agent1", "agent1")
	c := &chatTUI{store: s, sessionID: root.ID}
	sess, err := c.resolveSessionRef("@AGENT1")
	if err != nil {
		t.Fatalf("resolve session ref case-insensitive: %v", err)
	}
	if sess.ID != "session_child_ci" {
		t.Fatalf("unexpected case-insensitive session resolution: %#v", sess)
	}
}

func TestResolveSessionRefByAgentID(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	root, _ := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	_, _ = s.CloneSession(ctx, root.ID, "session_child", "@agent1", "agent1")
	c := &chatTUI{store: s, sessionID: root.ID}
	sess, err := c.resolveSessionRef("@agent1")
	if err != nil {
		t.Fatalf("resolve session ref: %v", err)
	}
	if sess.Scope == nil || sess.Scope.AgentID != "agent1" {
		t.Fatalf("unexpected session: %#v", sess)
	}
}

func TestAgentIDForSessionDefaults(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_agent_default", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s}
	if got := c.agentIDForSession(&store.Session{ID: "session_agent_default"}); got != "gi" {
		t.Fatalf("expected canonical default agent id gi, got %q", got)
	}
	if got := c.agentIDForSession(nil); got != defaultForkAgentID {
		t.Fatalf("expected nil session fallback %q, got %q", defaultForkAgentID, got)
	}
}

func TestIndexedAgentID(t *testing.T) {
	if got := indexedAgentID("session-a", map[string]string{"session-a": " agent-a "}); got != "agent-a" {
		t.Fatalf("expected trimmed indexed agent id, got %q", got)
	}
}

func TestFirstForkAgentID(t *testing.T) {
	if got := firstForkAgentID(" agent "); got != "agent1" {
		t.Fatalf("expected first fork agent id agent1, got %q", got)
	}
}

func TestNormalizeForkAgentBase(t *testing.T) {
	if got := normalizeForkAgentBase(" agent42 "); got != "agent" {
		t.Fatalf("expected normalized fork base agent, got %q", got)
	}
	if got := normalizeForkAgentBase("   "); got != defaultForkAgentID {
		t.Fatalf("expected empty fork base to default to %q, got %q", defaultForkAgentID, got)
	}
}

func TestResolveSessionRefPrefersCanonicalIdentityOverScopeSnapshot(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	rootAlloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_root_identity")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_root_identity", "", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}, &rootAlloc.Scope, rootAlloc.SessionAliases); err != nil {
		t.Fatalf("create root session: %v", err)
	}
	childAlloc := gisession.AllocateDefaultSession("agent1", "gi", "default", "session_child_identity")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_child_identity", "session_root_identity", "@agent1", map[string]any{"model": "bootstrap", "status": "idle"}, &childAlloc.Scope, childAlloc.SessionAliases); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update sessions set scope_json = ? where id = ?`, `{"version":1,"agent_id":"wrong77","channel":"gi","account":"default","dimensions":["chat"],"values":{"chat":"direct:session_child_identity"}}`, "session_child_identity"); err != nil {
		t.Fatalf("mutate child scope snapshot: %v", err)
	}
	c := &chatTUI{store: s, sessionID: "session_root_identity"}
	sess, err := c.resolveSessionRef("@agent1")
	if err != nil {
		t.Fatalf("resolve session ref: %v", err)
	}
	if sess.ID != "session_child_identity" {
		t.Fatalf("expected canonical identity session, got %#v", sess)
	}
	if got := c.nextForkAgentID(); got != "agent2" {
		t.Fatalf("expected canonical next fork agent id agent2, got %q", got)
	}
}

func TestInitialSessionIDPrefersMainSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	allocA := gisession.AllocateDefaultSession("agent", "gi", "default", "session_main_a")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_main_a", "", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}, &allocA.Scope, allocA.SessionAliases); err != nil {
		t.Fatalf("create session a: %v", err)
	}
	allocB := gisession.AllocateDefaultSession("agent", "gi", "default", "session_main_b")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_main_b", "", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}, &allocB.Scope, allocB.SessionAliases); err != nil {
		t.Fatalf("create session b: %v", err)
	}
	if err := s.SetMainSession(ctx, "session_main_b"); err != nil {
		t.Fatalf("set main session: %v", err)
	}
	sessionID, err := initialSessionID(ctx, s)
	if err != nil {
		t.Fatalf("initial session id: %v", err)
	}
	if sessionID != "session_main_b" {
		t.Fatalf("expected main session id, got %q", sessionID)
	}
}

func TestHandleEventStreamsDraftIntoTranscript(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": "hello"})
	if c.status != "⏳ hello…" || len(c.transcript) != 1 || c.transcript[0] != "Neo: hello" {
		t.Fatalf("unexpected first draft state: status=%q transcript=%#v", c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": " world"})
	if len(c.transcript) != 1 || c.transcript[0] != "Neo: hello world" {
		t.Fatalf("unexpected merged draft transcript: %#v", c.transcript)
	}
	c.handleEvent(map[string]any{"type": "new_post", "data": map[string]any{"content": "hello world"}})
	if c.running || c.draft != "" || c.draftLineIndex != -1 || c.transcript[0] != "Neo: hello world" {
		t.Fatalf("unexpected final streamed state: running=%v draft=%q draftLineIndex=%d transcript=%#v", c.running, c.draft, c.draftLineIndex, c.transcript)
	}
}

func TestHandleTopicEventTurnDraftRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.draft", Payload: map[string]any{"delta": "hello"}})
	if !c.running || c.status != "⏳ hello…" || len(c.transcript) != 1 || c.transcript[0] != "Neo: hello" {
		t.Fatalf("turn.draft rendering = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
}

func TestHandleTopicEventTurnThoughtRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.thought", Payload: map[string]any{"delta": "pondering"}})
	if !c.running || c.status != "Thinking…" {
		t.Fatalf("turn.thought rendering = running=%v status=%q", c.running, c.status)
	}
}

func TestHandleTopicEventTurnResponseSystemMessageRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, running: true, draft: "hello", transcript: []string{"Neo: hello"}, draftLineIndex: 0}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.response", Payload: map[string]any{"sender": "system", "data": map[string]any{"type": "system_message", "content": "Turn cancelled"}}})
	if c.running || c.draft != "" || c.draftLineIndex != -1 {
		t.Fatalf("expected system turn.response to clear running/draft state, got running=%v draft=%q draftLineIndex=%d", c.running, c.draft, c.draftLineIndex)
	}
	if got := c.transcript[len(c.transcript)-1]; got != "sys: Turn cancelled" {
		t.Fatalf("unexpected system turn.response transcript: %#v", c.transcript)
	}
}

func TestHandleTopicEventTurnStatusAndResponseRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1, running: true, draft: "hello"}
	c.running = false
	c.handleTopicEvent(topics.Envelope{Topic: "turn.status", Payload: map[string]any{"title": "Thinking…", "status": "running"}})
	if !c.running || c.status != "Thinking…" {
		t.Fatalf("turn.status rendering = running=%v status=%q", c.running, c.status)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "turn.status", Payload: map[string]any{"status": "idle"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != -1 || len(c.transcript) != 0 {
		t.Fatalf("turn.status idle cleanup = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.transcript = []string{"Neo: hello"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "turn.response", Payload: map[string]any{"data": map[string]any{"content": "hello world"}}})
	if c.running || c.draft != "" || c.draftLineIndex != -1 || c.transcript[0] != "Neo: hello world" {
		t.Fatalf("turn.response rendering = running=%v draft=%q draftLineIndex=%d transcript=%#v", c.running, c.draft, c.draftLineIndex, c.transcript)
	}
}

func TestHandleTopicEventSteeringAndSubturnRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "session.steering", Payload: map[string]any{"type": "steering_enqueued"}})
	if c.status != "Queued follow-up" {
		t.Fatalf("steering status = %q", c.status)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.subturn", Payload: map[string]any{"type": "subturn_created", "child_turn_id": "turn_child"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: sub-turn started: turn_child" {
		t.Fatalf("subturn created transcript = %q", got)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.subturn", Payload: map[string]any{"type": "subturn_status", "child_turn_id": "turn_child", "status": "completed"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: sub-turn turn_child: completed" {
		t.Fatalf("subturn status transcript = %q", got)
	}
}

func TestHandleTopicEventCompactionAndRoutingRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "session.compaction", Payload: map[string]any{"messages_before": 10, "messages_after": 4, "tokens_before": 1234}})
	if c.status != "Compacted context" || c.transcript[len(c.transcript)-1] != "sys: compacted context: messages 10→4, tokens_before=1234" {
		t.Fatalf("compaction topic status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.routing", Payload: map[string]any{"type": "routing_decision", "target_agent_id": "agent1", "target_session_id": "session_child"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: routed to @agent1 (session_child)" {
		t.Fatalf("routing decision transcript = %q", got)
	}
}

func TestHandleTopicEventInboundWorkRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Payload: map[string]any{"type": "inbound_work_enqueued", "source_kind": "ipc", "status": "queued"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: inbound work queued (ipc) [queued]" {
		t.Fatalf("inbound work enqueue transcript = %q", got)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Payload: map[string]any{"type": "inbound_work_retry_scheduled", "source_kind": "ipc", "status": "retry", "attempt_count": 2}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: inbound work retry scheduled (ipc) attempt 2 [retry]" {
		t.Fatalf("inbound work retry transcript = %q", got)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Payload: map[string]any{"type": "inbound_work_requeued", "source_kind": "ipc", "status": "queued"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: inbound work requeued (ipc) [queued]" {
		t.Fatalf("inbound work requeued transcript = %q", got)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Payload: map[string]any{"type": "inbound_work_discarded", "source_kind": "ipc", "status": "discarded"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: inbound work discarded (ipc) [discarded]" {
		t.Fatalf("inbound work discarded transcript = %q", got)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Payload: map[string]any{"type": "inbound_work_failed", "source_kind": "ipc", "status": "failed", "error": "decode failed"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: inbound work failed (ipc) [failed]: decode failed" {
		t.Fatalf("inbound work failed transcript = %q", got)
	}
}

func TestHandleTopicEventTurnAndSessionRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.running = false
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.turn", Payload: map[string]any{"type": "turn_state", "status": "running", "phase": "waiting_on_tools", "tool": "read"}})
	if !c.running || c.status != "Running: read" {
		t.Fatalf("turn state status = running=%v status=%q", c.running, c.status)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.turn", Payload: map[string]any{"type": "turn_completed", "status": "completed"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != -1 || len(c.transcript) != 0 {
		t.Fatalf("turn completed cleanup = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.turn", Payload: map[string]any{"type": "turn_terminal", "status": "failed"}})
	if c.status != "Turn failed" || c.running || c.draft != "" || c.draftLineIndex != -1 || len(c.transcript) != 0 {
		t.Fatalf("turn terminal cleanup = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.running = false
	c.status = ""
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.session", Payload: map[string]any{"type": "session_running", "status": "running"}})
	if !c.running || c.status != "Neo · bootstrap" {
		t.Fatalf("session running state = running=%v status=%q", c.running, c.status)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.session", Payload: map[string]any{"type": "session_idle", "status": "idle"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != -1 || len(c.transcript) != 0 {
		t.Fatalf("session idle cleanup = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.session", Payload: map[string]any{"type": "session_state", "status": "queued"}})
	if c.status != "Queued" || c.running || c.draft != "" || c.draftLineIndex != -1 || len(c.transcript) != 0 {
		t.Fatalf("session state queued rendering = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.session", Payload: map[string]any{"type": "session_state", "status": "idle"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != -1 || len(c.transcript) != 0 {
		t.Fatalf("session state idle cleanup = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
}

func TestHandleTopicEventHookInvocationErrorRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.hook", Payload: map[string]any{"type": "hook_invocation", "hook": "tool_call", "tool": "grep", "error": "timed out after 1500ms"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: hook invocation error via tool_call for grep: timed out after 1500ms" {
		t.Fatalf("hook invocation error transcript = %q", got)
	}
}

func TestHandleTopicEventHookDecisionRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.hook", Payload: map[string]any{"type": "hook_modify", "hook": "tool_call", "tool": "grep"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: hook modified via tool_call for grep" {
		t.Fatalf("hook modify transcript = %q", got)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.hook", Payload: map[string]any{"type": "hook_respond", "hook": "tool_call", "tool": "grep"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: hook responded directly via tool_call for grep" {
		t.Fatalf("hook respond transcript = %q", got)
	}
}

func TestHandleTopicEventStatusRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.running = false
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.tool", Payload: map[string]any{"type": "tool_started", "tool": "read"}})
	if !c.running || c.status != "Running: read" {
		t.Fatalf("tool started status = running=%v status=%q", c.running, c.status)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.tool", Payload: map[string]any{"type": "tool_skipped", "tool": "shell", "reason": "queued user steering message"}})
	if len(c.transcript) == 0 || c.transcript[len(c.transcript)-1] != "sys: tool skipped: shell: queued user steering message" {
		t.Fatalf("tool skipped transcript = %#v", c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.hook", Payload: map[string]any{"type": "hook_deny", "hook": "approve_tool", "tool": "shell", "reason": "tool not approved"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: hook hook_deny via approve_tool for shell: tool not approved" {
		t.Fatalf("hook deny transcript = %q", got)
	}
}

func TestHandleEventErrorClearsRunningState(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, running: true, draft: "hello", draftLineIndex: 0, transcript: []string{"Neo: hello"}}
	c.handleEvent(map[string]any{"type": "error", "error": "boom"})
	if c.running || c.draft != "" || c.draftLineIndex != -1 {
		t.Fatalf("expected legacy error path to clear running/draft state, got running=%v draft=%q draftLineIndex=%d", c.running, c.draft, c.draftLineIndex)
	}
	if got := c.transcript[len(c.transcript)-1]; got != "error: boom" {
		t.Fatalf("unexpected error transcript: %#v", c.transcript)
	}
}

func TestHandleEventAgentStatusWithoutAppDoesNotPanic(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}}
	c.handleEvent(map[string]any{"type": "agent_status", "title": "Thinking…"})
	if !c.running || c.status != "Thinking…" {
		t.Fatalf("agent_status without app = running=%v status=%q", c.running, c.status)
	}
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleEvent(map[string]any{"type": "agent_status"})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != -1 || len(c.transcript) != 0 {
		t.Fatalf("agent_status idle cleanup = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
}

func TestUseTopicNativeRuntimeStatusRequiresLiveSubscription(t *testing.T) {
	c := &chatTUI{topicEventCh: make(chan topics.Envelope, 1)}
	if c.useTopicNativeRuntimeStatus() {
		t.Fatal("expected topic-native runtime status to remain disabled without a live topic subscription")
	}
	c.topicUnsubscribe = func() {}
	if !c.useTopicNativeRuntimeStatus() {
		t.Fatal("expected topic-native runtime status to activate with a live topic subscription")
	}
}

func TestHandleEventStatusRenderingSkipsDuplicateLegacyRuntimeEventsWhenTopicNativeActive(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1, topicEventCh: make(chan topics.Envelope, 1), topicUnsubscribe: func() {}}
	c.handleEvent(map[string]any{"type": "tool_failed", "tool": "shell", "error": "boom"})
	if c.status != "" || len(c.transcript) != 0 {
		t.Fatalf("expected topic-native path to suppress duplicate legacy tool event, got status=%q transcript=%#v", c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "compaction", "messages_before": 10, "messages_after": 4, "tokens_before": 1234})
	if c.status != "" || len(c.transcript) != 0 {
		t.Fatalf("expected topic-native path to suppress duplicate legacy compaction event, got status=%q transcript=%#v", c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "agent_status", "title": "Thinking…"})
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": "hello"})
	c.handleEvent(map[string]any{"type": "agent_thought_delta", "delta": "pondering"})
	c.handleEvent(map[string]any{"type": "new_post", "data": map[string]any{"content": "hello world"}})
	c.handleEvent(map[string]any{"type": "routing_decision", "target_agent_id": "agent1", "target_session": "session_child"})
	if c.status != "" || len(c.transcript) != 0 || c.draft != "" {
		t.Fatalf("expected topic-native path to suppress duplicate legacy turn status/response events, got status=%q draft=%q transcript=%#v", c.status, c.draft, c.transcript)
	}
}

func TestHandleEventStatusRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.running = false
	c.handleEvent(map[string]any{"type": "agent_thought_delta"})
	if !c.running || c.status != "Thinking…" {
		t.Fatalf("thinking status = running=%v status=%q", c.running, c.status)
	}
	c.running = false
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": "hello"})
	if !c.running || c.status != "⏳ hello…" || len(c.transcript) == 0 || c.transcript[len(c.transcript)-1] != "Neo: hello" {
		t.Fatalf("draft status = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "tool_finished", "tool": "read"})
	if c.status != "Tool finished: read" {
		t.Fatalf("tool finished status = %q", c.status)
	}
	c.handleEvent(map[string]any{"type": "tool_failed", "tool": "shell", "error": "boom"})
	if c.status != "Tool failed: shell" || len(c.transcript) == 0 || c.transcript[len(c.transcript)-1] != "sys: tool failed: shell: boom" {
		t.Fatalf("tool failed status/transcript = %q %#v", c.status, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleEvent(map[string]any{"type": "new_post", "data": map[string]any{"content": "hello world"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != -1 || len(c.transcript) != 1 || c.transcript[0] != "Neo: hello world" {
		t.Fatalf("new_post cleanup = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "compaction", "messages_before": 10, "messages_after": 4, "tokens_before": 1234})
	if c.status != "Compacted context" || c.transcript[len(c.transcript)-1] != "sys: compacted context: messages 10→4, tokens_before=1234" {
		t.Fatalf("compaction status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "routing_decision", "target_agent_id": "agent1", "target_session": "session_child"})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: routed to @agent1 (session_child)" {
		t.Fatalf("legacy routing decision transcript = %q", got)
	}
}

func TestOnSubmitRequiresModelSelectionBeforeFirstPrompt(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_1", cfg: config.RuntimeConfig{AssistantName: "Neo"}, draftLineIndex: -1}
	c.input = newMultilineInput(80, "", c.onSubmit, nil)
	c.onSubmit("hello")
	if c.running {
		t.Fatal("expected prompt submission to be blocked without a model")
	}
	joined := strings.Join(c.transcript, "\n")
	if !strings.Contains(joined, "no model selected") {
		t.Fatalf("expected model-selection guidance, got:\n%s", joined)
	}
}

func TestInitUsesConfiguredGiDefaultModelWithoutFirstUsePrompt(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	sess, err := s.CreateSession(context.Background(), "session_1", "@agent", map[string]any{"status": "idle"})
	if err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: sess.ID, cfg: config.RuntimeConfig{AssistantName: "Gi", DefaultModel: "opencode-zen/minimax-m2.5-free"}}
	cleanup := c.Init()
	defer cleanup()
	if c.status != "Gi · opencode-zen/minimax-m2.5-free" {
		t.Fatalf("unexpected status: %q", c.status)
	}
	joined := strings.Join(c.transcript, "\n")
	if strings.Contains(joined, "no model selected") {
		t.Fatalf("unexpected first-use prompt with default model: %s", joined)
	}
}

func TestModelCommandPersistsSelection(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, sessionID: "session_1", cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultThinkingLevel: "medium", EnabledModels: []string{"qwen3:latest"}}}
	lines := c.modelCommand([]string{"/model", "ollama/gemma4:latest"})
	if len(lines) == 0 || !strings.Contains(lines[0], "model set to ollama/gemma4:latest") {
		t.Fatalf("unexpected model command output: %#v", lines)
	}
	cfg := config.Load(root)
	if cfg.DefaultProvider != "ollama" || cfg.DefaultModel != "ollama/gemma4:latest" {
		t.Fatalf("unexpected persisted model config: %#v", cfg)
	}
}

func TestAppendTranscriptPrunesToScrollbackLimit(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{ScrollbackLimit: 3}}
	c.appendTranscript("1", "2", "3", "4")
	if got := strings.Join(c.transcript, ","); got != "2,3,4" {
		t.Fatalf("unexpected pruned transcript: %s", got)
	}
}

func TestScrollbackCommandPersistsLimit(t *testing.T) {
	root := t.TempDir()
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: root, ScrollbackLimit: 1000}}
	lines := c.scrollbackCommand([]string{"/scrollback", "250"})
	if len(lines) == 0 || !strings.Contains(lines[0], "250") {
		t.Fatalf("unexpected scrollback output: %#v", lines)
	}
	cfg := config.Load(root)
	if cfg.ScrollbackLimit != 250 {
		t.Fatalf("unexpected persisted scrollback limit: %#v", cfg)
	}
}

func TestSettingsLinesExposeRuntimeState(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	engine := turn.New(s)
	c := &chatTUI{store: s, engine: engine, cfg: config.RuntimeConfig{WorkspaceRoot: "/tmp/ws", DefaultProvider: "test", DefaultModel: "m", DefaultThinkingLevel: "low", MaxIterations: 64}}
	lines := strings.Join(c.settingsLines(), "\n")
	for _, want := range []string{"settings: runtime:", "provider=test model=m thinking=low", "scrollback_limit=1000", "workspace=/tmp/ws", "tools active="} {
		if !strings.Contains(lines, want) {
			t.Fatalf("settings missing %q:\n%s", want, lines)
		}
	}
}

func TestRenderMessageLinesFormatsMarkdown(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	lines := c.renderMessageLines(store.Message{Role: "assistant", Content: "# Title\n\n- first\n- second", Payload: map[string]any{"kind": "chat"}}, 80)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Neo: TITLE", "=====", "• first", "• second"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("markdown render missing %q:\n%s", want, joined)
		}
	}
}

func TestRenderMessageLinesFormatsTableResponsively(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	content := "| Name | Role | Value |\n| --- | --- | --- |\n| Alice | admin | 42 |\n| Bob | user | 7 |"
	lines := c.renderMessageLines(store.Message{Role: "assistant", Content: content, Payload: map[string]any{"kind": "chat"}}, 20)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Name: Alice", "Role: admin", "Value: 42", "Name: Bob", "Role: user", "Value: 7"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("responsive table render missing %q:\n%s", want, joined)
		}
	}
}

func TestRenderMessageLineFoldsToolAndCompaction(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	toolLine := c.renderMessageLine(store.Message{Role: "tool_result", Content: "long tool output", Payload: map[string]any{"kind": "tool_result", "tool_name": "shell", "is_error": false}})
	if toolLine != "tool[shell/ok]: long tool output" {
		t.Fatalf("tool line = %q", toolLine)
	}
	compactionLine := c.renderMessageLine(store.Message{Role: "assistant", Content: "summary", Payload: map[string]any{"kind": "compaction", "tokens_before": 1200}})
	if compactionLine != "compact: summary (tokens_before=1200)" {
		t.Fatalf("compaction line = %q", compactionLine)
	}
}

func TestMultilineInputExpandsForWrappedText(t *testing.T) {
	inp := newMultilineInput(10, "", nil, nil)
	inp.SetText("1234567890123456789012345")
	el := inp.Render(nil)
	if got := el.HeightForWidth(10); got < 3 {
		t.Fatalf("expected expanded input height, got %d", got)
	}
}

func TestContextSummaryLinesWrapForNarrowWidth(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	sess, err := s.CreateSession(context.Background(), "session_1", "@agent", map[string]any{"status": "idle", "model": "opencode-zen/minimax-m2.5-free", "provider": "opencode-zen", "thinking_level": "low"})
	if err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{store: s, sessionID: sess.ID, cfg: config.RuntimeConfig{DefaultModel: "opencode-zen/minimax-m2.5-free", DefaultProvider: "opencode-zen", DefaultThinkingLevel: "low"}}
	lines := c.contextSummaryLines(60)
	if len(lines) < 3 {
		t.Fatalf("expected wrapped summary lines, got %#v", lines)
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Session:", "Agent:", "Model:", "Provider:", "Messages:"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("responsive summary missing %q:\n%s", want, joined)
		}
	}
}

func TestFooterTextContainsStableHints(t *testing.T) {
	c := &chatTUI{}
	footer := c.footerText()
	for _, want := range []string{"Hints:", "/help", "/tools active|activate|reset", "Tab focus", "F2/F3 history", "Ctrl-D quit"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("footer missing %q: %s", want, footer)
		}
	}
}

func TestPluginLinesShowsExtensionsAndHooks(t *testing.T) {
	rootDir := t.TempDir()
	extDir := filepath.Join(rootDir, ".gi", "extensions")
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extDir, "demo.joke"), []byte("nil"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.NewWithRuntimeConfig(s, config.RuntimeConfig{WorkspaceRoot: rootDir, DefaultModel: "bootstrap"}, "")
	engine.RegisterHook("tool_call", "test-hook", func(context.Context, turn.HookRequest) (turn.HookResponse, error) { return turn.HookResponse{}, nil })
	c := &chatTUI{store: s, engine: engine}
	lines := strings.Join(c.pluginLines(), "\n")
	for _, want := range []string{"plugins: extensions:", "loaded joker .gi/extensions/demo.joke", "plugins: hooks:", "tool_call from test-hook"} {
		if !strings.Contains(lines, want) {
			t.Fatalf("plugins missing %q:\n%s", want, lines)
		}
	}
}
