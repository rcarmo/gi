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
	for _, want := range []string{"tree: sessions:", "@gi session_root", "@agent · idle", "* @agent1 session_child", "@agent1 · idle", "messages=0 turns=0"} {
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

func TestResolveSessionRefByAgentIDDeterministicOrder(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	root, err := s.CreateSession(ctx, "session_root_order", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create root session: %v", err)
	}
	if _, err := s.CloneSession(ctx, root.ID, "session_child_order_b", "@agent1", "agent1"); err != nil {
		t.Fatalf("create child agent1 b: %v", err)
	}
	if _, err := s.CloneSession(ctx, root.ID, "session_child_order_a", "@agent1", "agent1"); err != nil {
		t.Fatalf("create child agent1 a: %v", err)
	}
	c := &chatTUI{store: s, sessionID: root.ID}
	sess, err := c.resolveSessionRef("@agent1")
	if err != nil {
		t.Fatalf("resolve session ref deterministic order: %v", err)
	}
	if sess.ID != "session_child_order_a" {
		t.Fatalf("expected deterministic first sorted session id, got %#v", sess)
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

func TestLoadSessionIdentityIndexEmptyStore(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	c := &chatTUI{store: s}
	sessionIDs, index, err := c.loadSessionIdentityIndex(context.Background())
	if err != nil {
		t.Fatalf("load session identity index empty store: %v", err)
	}
	if len(sessionIDs) != 0 || len(index) != 0 {
		t.Fatalf("expected empty session/index slices for empty store, got ids=%#v index=%#v", sessionIDs, index)
	}
}

func TestLoadSessionIdentityIndexNilContext(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_load_index_nil_ctx", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s}
	sessionIDs, index, err := c.loadSessionIdentityIndex(nil)
	if err != nil {
		t.Fatalf("load session identity index nil ctx: %v", err)
	}
	if len(sessionIDs) == 0 {
		t.Fatalf("expected at least one session id in nil-ctx load")
	}
	if index["session_load_index_nil_ctx"] != "gi" {
		t.Fatalf("expected canonical default agent id gi in nil-ctx load index, got %#v", index)
	}
}

func TestSessionAgentIDIndexNilContext(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_index_nil_ctx", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s}
	index, err := c.sessionAgentIDIndex(nil)
	if err != nil {
		t.Fatalf("session agent id index nil ctx: %v", err)
	}
	if index["session_index_nil_ctx"] != "gi" {
		t.Fatalf("expected canonical default agent id gi in nil-ctx index, got %#v", index)
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

func TestNormalizeIndexedSessionID(t *testing.T) {
	if got := normalizeIndexedSessionID("  session-a  "); got != "session-a" {
		t.Fatalf("expected normalized indexed session id session-a, got %q", got)
	}
}

func TestIndexedAgentIDLower(t *testing.T) {
	if got := indexedAgentIDLower("session-a", map[string]string{"session-a": " Agent-A "}); got != "agent-a" {
		t.Fatalf("expected lower indexed agent id agent-a, got %q", got)
	}
	if got := indexedAgentIDLower("  session-a  ", map[string]string{"session-a": " Agent-A "}); got != "agent-a" {
		t.Fatalf("expected lower indexed agent id with normalized key lookup, got %q", got)
	}
	if got := indexedAgentIDLower("session-a", nil); got != strings.ToLower(defaultForkAgentID) {
		t.Fatalf("expected nil index map to default lower indexed agent id to %q, got %q", strings.ToLower(defaultForkAgentID), got)
	}
}

func TestIndexedAgentIDOrDefault(t *testing.T) {
	if got := indexedAgentIDOrDefault("session-a", map[string]string{"session-a": " Agent-A "}); got != "Agent-A" && got != "agent-a" {
		t.Fatalf("expected indexed agent id to be non-empty normalized value, got %q", got)
	}
	if got := indexedAgentIDOrDefault("missing", map[string]string{}); got != defaultForkAgentID {
		t.Fatalf("expected missing indexed agent id to default to %q, got %q", defaultForkAgentID, got)
	}
	if got := indexedAgentIDOrDefault("   ", map[string]string{"": "agent-blank"}); got != defaultForkAgentID {
		t.Fatalf("expected blank session id to fail closed to %q, got %q", defaultForkAgentID, got)
	}
}

func TestNormalizeSessionRefLower(t *testing.T) {
	if got := strings.ToLower(normalizeSessionRef("  @Agent1  ")); got != "agent1" {
		t.Fatalf("expected normalized lower session ref agent1, got %q", got)
	}
}

func TestNormalizeSessionRef(t *testing.T) {
	if got := normalizeSessionRef("  @agent1  "); got != "agent1" {
		t.Fatalf("expected normalized session ref agent1, got %q", got)
	}
}

func TestForkAgentIDCandidate(t *testing.T) {
	if got := forkAgentIDCandidate("agent", 3); got != "agent3" {
		t.Fatalf("expected candidate agent3, got %q", got)
	}
}

func TestLastForkAgentID(t *testing.T) {
	if got := lastForkAgentID("agent"); got != forkAgentIDCandidate("agent", maxForkAgentIDSuffixExclusive-1) {
		t.Fatalf("expected last fork agent id to match max suffix candidate, got %q", got)
	}
}

func TestChooseNextForkAgentID(t *testing.T) {
	if candidate, ok := chooseNextForkAgentID("agent", map[string]bool{"agent1": true, "agent2": true}); !ok || candidate != "agent3" {
		t.Fatalf("expected next fork candidate agent3, got %q ok=%v", candidate, ok)
	}
	if candidate, ok := chooseNextForkAgentID("  ", nil); !ok || candidate != "agent1" {
		t.Fatalf("expected default-base candidate agent1 for nil used set, got %q ok=%v", candidate, ok)
	}
	if candidate, ok := chooseNextForkAgentID("agent", map[string]bool{}); !ok || candidate != "agent1" {
		t.Fatalf("expected first fork candidate agent1 for empty used set, got %q ok=%v", candidate, ok)
	}
	exhausted := map[string]bool{}
	for i := minForkAgentIDSuffix; i < maxForkAgentIDSuffixExclusive; i++ {
		exhausted[forkAgentIDCandidate("agent", i)] = true
	}
	if candidate, ok := chooseNextForkAgentID("agent", exhausted); ok || candidate != "" {
		t.Fatalf("expected no candidate on exhaustion, got %q ok=%v", candidate, ok)
	}
}

func TestFindSessionIDByNormalizedAgentRef(t *testing.T) {
	sessionID, ok := findSessionIDByNormalizedAgentRef([]string{"s2", "s1"}, map[string]string{"s1": "agent-a", "s2": "Agent-B"}, "agent-b")
	if !ok || sessionID != "s2" {
		t.Fatalf("expected deterministic normalized agent-ref match on s2, got id=%q ok=%v", sessionID, ok)
	}
	if sessionID, ok := findSessionIDByNormalizedAgentRef([]string{"s1"}, map[string]string{"s1": "agent-a"}, ""); ok || sessionID != "" {
		t.Fatalf("expected empty normalized ref to fail fast, got id=%q ok=%v", sessionID, ok)
	}
	if sessionID, ok := findSessionIDByNormalizedAgentRef(nil, map[string]string{"s1": "agent-a"}, "agent-a"); ok || sessionID != "" {
		t.Fatalf("expected empty session-id input to fail fast, got id=%q ok=%v", sessionID, ok)
	}
}

func TestFindSessionIDByAgentRef(t *testing.T) {
	sessionID, ok := findSessionIDByNormalizedAgentRef([]string{"s2", "s1"}, map[string]string{"s1": "agent-a", "s2": "Agent-B"}, strings.ToLower(normalizeSessionRef("agent-b")))
	if !ok || sessionID != "s2" {
		t.Fatalf("expected deterministic agent-ref match on s2, got id=%q ok=%v", sessionID, ok)
	}
	if sessionID, ok := findSessionIDByNormalizedAgentRef([]string{"s1"}, map[string]string{"s1": "agent-a"}, strings.ToLower(normalizeSessionRef("agent-z"))); ok || sessionID != "" {
		t.Fatalf("expected no match for unknown agent ref, got id=%q ok=%v", sessionID, ok)
	}
	if sessionID, ok := findSessionIDByNormalizedAgentRef([]string{"s-default"}, map[string]string{}, strings.ToLower(normalizeSessionRef(defaultForkAgentID))); !ok || sessionID != "s-default" {
		t.Fatalf("expected missing index entries to default-match %q, got id=%q ok=%v", defaultForkAgentID, sessionID, ok)
	}
	if sessionID, ok := findSessionIDByNormalizedAgentRef([]string{"s2"}, map[string]string{"s2": "agent-b"}, strings.ToLower(normalizeSessionRef(" @Agent-B "))); !ok || sessionID != "s2" {
		t.Fatalf("expected helper to normalize incoming ref before match, got id=%q ok=%v", sessionID, ok)
	}
	if sessionID, ok := findSessionIDByNormalizedAgentRef([]string{"s2"}, map[string]string{"s2": "agent-b"}, "  AGENT-B  "); !ok || sessionID != "s2" {
		t.Fatalf("expected normalized matcher to trim/lower direct input, got id=%q ok=%v", sessionID, ok)
	}
	if sessionID, ok := findSessionIDByNormalizedAgentRef([]string{"s2"}, map[string]string{"s2": "agent-b"}, strings.ToLower(normalizeSessionRef("   "))); ok || sessionID != "" {
		t.Fatalf("expected empty ref to fail fast with no match, got id=%q ok=%v", sessionID, ok)
	}
}

func TestBuildUsedForkAgentIDs(t *testing.T) {
	if used := buildUsedForkAgentIDs(nil, map[string]string{"s1": "agent-a"}); len(used) != 0 {
		t.Fatalf("expected nil session-id input to yield empty used set, got %#v", used)
	}
	used := buildUsedForkAgentIDs([]string{"s1", "s2", "   "}, map[string]string{"s1": "agent-a", "s2": " agent-b "})
	if !used["agent-a"] || !used["agent-b"] {
		t.Fatalf("expected used fork agent ids to include normalized values, got %#v", used)
	}
	if len(used) != 2 {
		t.Fatalf("expected blank session ids to be ignored, got %#v", used)
	}
	usedNilIndex := buildUsedForkAgentIDs([]string{"s-default"}, nil)
	if !usedNilIndex[defaultForkAgentID] || len(usedNilIndex) != 1 {
		t.Fatalf("expected nil agent index to default used set to %q, got %#v", defaultForkAgentID, usedNilIndex)
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
	if _, err := s.DB().ExecContext(ctx, `update sessions set state_json = state_json where id = ?`, "session_child_identity"); err != nil {
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
	if c.status != "Compacted context" || c.transcript[len(c.transcript)-1] != "sys[compact]: messages 10→4 · tokens_before=1234" {
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
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.dispatcher", Payload: map[string]any{"type": "dispatcher_lease_acquired", "worker_id": "worker-1"}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: inbound dispatcher lease acquired [worker-1]" {
		t.Fatalf("dispatcher lease transcript = %q", got)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.dispatcher", Payload: map[string]any{"type": "dispatcher_drain_completed", "worker_id": "worker-1", "processed_count": 3}})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: inbound dispatcher drain completed (3 processed) [worker-1]" {
		t.Fatalf("dispatcher drain transcript = %q", got)
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
	if c.status != "Compacted context" || c.transcript[len(c.transcript)-1] != "sys[compact]: messages 10→4 · tokens_before=1234" {
		t.Fatalf("compaction status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "routing_decision", "target_agent_id": "agent1", "target_session": "session_child"})
	if got := c.transcript[len(c.transcript)-1]; got != "sys: routed to @agent1 (session_child)" {
		t.Fatalf("legacy routing decision transcript = %q", got)
	}
}

func TestRestoreQueuedDraftMovesLastQueuedTextToEditor(t *testing.T) {
	c := &chatTUI{queuedDrafts: []string{"first", "second"}, status: "Queued follow-up"}
	c.input = newMultilineInput(80, "", c.onSubmit, nil)
	c.restoreQueuedDraft()
	if got := c.input.Text(); got != "second" {
		t.Fatalf("restored draft = %q", got)
	}
	if c.status != "Restored queued draft" || len(c.queuedDrafts) != 1 || c.queuedDrafts[0] != "first" {
		t.Fatalf("unexpected restore state status=%q queued=%#v", c.status, c.queuedDrafts)
	}
}

func TestMultilineInputAltBindingsAreAvailable(t *testing.T) {
	restored := false
	submitted := ""
	inp := newMultilineInput(80, "", func(text string) { submitted = text }, nil)
	inp.onRestoreQueued = func() { restored = true }
	inp.SetText("queued")
	for _, binding := range inp.KeyMap() {
		if binding.Pattern.Key == gotui.KeyEnter && binding.Pattern.Mod == gotui.ModAlt {
			binding.Handler(gotui.KeyEvent{Key: gotui.KeyEnter, Mod: gotui.ModAlt})
		}
		if binding.Pattern.Key == gotui.KeyUp && binding.Pattern.Mod == gotui.ModAlt {
			binding.Handler(gotui.KeyEvent{Key: gotui.KeyUp, Mod: gotui.ModAlt})
		}
	}
	if submitted != "queued" {
		t.Fatalf("Alt+Enter did not submit, got %q", submitted)
	}
	if !restored {
		t.Fatal("Alt+Up did not call restore callback")
	}
}

func TestCompleteInputPathCompletesWorkspaceRelativePath(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "readme.md"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: root}}
	got, cursor, ok := c.completeInputPath("open docs/re", len("open docs/re"))
	if !ok || got != "open docs/readme.md" || cursor != len("open docs/readme.md") {
		t.Fatalf("completion got=%q cursor=%d ok=%v", got, cursor, ok)
	}
}

func TestCompleteInputPathCompletesAtFileReference(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "chat.go"), []byte("package internal"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: root}}
	got, cursor, ok := c.completeInputPath("review @internal/ch", len("review @internal/ch"))
	if !ok || got != "review @internal/chat.go" || cursor != len("review @internal/chat.go") {
		t.Fatalf("@ completion got=%q cursor=%d ok=%v", got, cursor, ok)
	}
}

func TestMultilineInputTabCompletionCallback(t *testing.T) {
	inp := newMultilineInput(80, "", nil, nil)
	inp.SetText("abc")
	inp.onComplete = func(text string, cursor int) (string, int, bool) { return text + "def", cursor + 3, true }
	inp.complete()
	if got := inp.Text(); got != "abcdef" || inp.cursorPos != 6 {
		t.Fatalf("completion text=%q cursor=%d", got, inp.cursorPos)
	}
}

func TestLocalShellShortcutRunsLocally(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: t.TempDir()}}
	lines := strings.Join(c.localShellShortcutLines("printf hello"), "\n")
	for _, want := range []string{"local$ printf hello", "│ hello"} {
		if !strings.Contains(lines, want) {
			t.Fatalf("local shell output missing %q:\n%s", want, lines)
		}
	}
}

func TestOnSubmitBangShortcutSendsShellRequestToModel(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(context.Background(), "session_bang", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_bang", cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, running: true, draftLineIndex: -1, transcriptRef: gotui.NewRef()}
	c.input = newMultilineInput(80, "", c.onSubmit, nil)
	c.onSubmit("!pwd")
	joined := strings.Join(c.transcript, "\n")
	if !strings.Contains(joined, "you [queued]: Run this shell command and summarize the result: pwd") {
		t.Fatalf("bang shortcut did not transform to model request:\n%s", joined)
	}
}

func TestOnSubmitWhileRunningShowsQueuedFeedback(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(context.Background(), "session_queue_feedback", "@agent", map[string]any{"status": "running", "model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_queue_feedback", cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, running: true, draftLineIndex: -1, transcriptRef: gotui.NewRef()}
	c.input = newMultilineInput(80, "", c.onSubmit, nil)
	c.onSubmit("please continue")
	if c.status != "Queued follow-up" {
		t.Fatalf("queued status = %q", c.status)
	}
	if len(c.transcript) == 0 || c.transcript[len(c.transcript)-1] != "you [queued]: please continue" {
		t.Fatalf("queued transcript = %#v", c.transcript)
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

func TestHelpLinesAreGroupedAndDiscoverCoreCommands(t *testing.T) {
	c := &chatTUI{}
	joined := strings.Join(c.helpLines(), "\n")
	for _, want := range []string{
		"help: gi TUI",
		"keys:",
		"runtime:",
		"discovery:",
		"sessions:",
		"/session",
		"/new",
		"/name <name>",
		"/resume [index|session_id]",
		"/clone [@agentN]",
		"/copy",
		"/settings",
		"/where",
		"/tools [query|active|activate|reset]",
		"/skills [query]",
		"/approvals",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("help missing %q:\n%s", want, joined)
		}
	}
}

func TestNewSessionCommandCreatesAndSwitchesMainSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_existing_main")
	if _, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_existing_main", Title: "@agent", State: map[string]any{"status": "idle", "model": "old"}, Allocation: alloc}); err != nil {
		t.Fatalf("create existing main session: %v", err)
	}
	engine := turn.New(s)
	c := &chatTUI{store: s, engine: engine, sessionID: "session_existing_main", cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "new-model", DefaultProvider: "provider", DefaultThinkingLevel: "medium"}, draftLineIndex: -1, transcriptRef: gotui.NewRef()}
	c.eventCh = make(chan map[string]any, 64)
	c.topicEventCh = make(chan topics.Envelope, 64)
	lines := c.newSessionLines()
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "sys: new session @agent (session_") {
		t.Fatalf("unexpected new session output: %#v", lines)
	}
	if c.sessionID == "" || c.sessionID == "session_existing_main" {
		t.Fatalf("expected switched session, got %q", c.sessionID)
	}
	mainID, err := s.ResolveMainSessionID(ctx, "agent", "gi", "default")
	if err != nil {
		t.Fatalf("resolve main session: %v", err)
	}
	if mainID != c.sessionID {
		t.Fatalf("expected new session to become main, got main=%q current=%q", mainID, c.sessionID)
	}
	created, err := s.GetSession(ctx, c.sessionID)
	if err != nil {
		t.Fatalf("get new session: %v", err)
	}
	if created.State["model"] != "new-model" || created.State["provider"] != "provider" || created.State["thinking_level"] != "medium" {
		t.Fatalf("unexpected new session state: %#v", created.State)
	}
}

func TestReloadLinesRefreshesConfigOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"defaultProvider":"test","defaultModel":"model-a","defaultThinkingLevel":"high","enabledModels":["model-a"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultProvider: "old", DefaultModel: "old-model", DefaultThinkingLevel: "low"}, input: newMultilineInput(80, "old", nil, nil)}
	joined := strings.Join(c.reloadLines(), "\n")
	for _, want := range []string{"reload: config refreshed", "provider=test model=model-a thinking=high", "active runtime hooks/extensions remain mounted"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("reload output missing %q:\n%s", want, joined)
		}
	}
	if c.cfg.DefaultProvider != "test" || c.cfg.DefaultModel != "model-a" || c.cfg.DefaultThinkingLevel != "high" {
		t.Fatalf("config was not reloaded: %#v", c.cfg)
	}
	if c.input.placeholder != "Send a message…" {
		t.Fatalf("placeholder not refreshed: %q", c.input.placeholder)
	}
}

func TestCopyLastAssistantLinesFallsBackToTranscript(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_copy", "@agent", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_user_copy", "session_copy", "user", "question", nil); err != nil {
		t.Fatalf("add user message: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_assistant_copy", "session_copy", "assistant", "answer\nsecond line", nil); err != nil {
		t.Fatalf("add assistant message: %v", err)
	}
	c := &chatTUI{store: s, sessionID: "session_copy"}
	joined := strings.Join(c.copyLastAssistantLines(), "\n")
	for _, want := range []string{"copy: clipboard unavailable; last assistant message follows", "copy: answer", "  second line"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("copy fallback missing %q:\n%s", want, joined)
		}
	}
	if _, err := s.CreateSession(ctx, "session_copy_empty", "@agent", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create empty session: %v", err)
	}
	c.sessionID = "session_copy_empty"
	lines := c.copyLastAssistantLines()
	if len(lines) != 1 || lines[0] != "copy: no assistant message found" {
		t.Fatalf("unexpected empty copy output: %#v", lines)
	}
}

func TestCloneSessionCommandClonesAndSwitches(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_clone_source")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_clone_source", "", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}, &alloc.Scope, alloc.SessionAliases); err != nil {
		t.Fatalf("create source session: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_clone_source", "session_clone_source", "assistant", "hello", nil); err != nil {
		t.Fatalf("add source message: %v", err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_clone_source", cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, transcriptRef: gotui.NewRef(), draftLineIndex: -1}
	c.eventCh = make(chan map[string]any, 64)
	c.topicEventCh = make(chan topics.Envelope, 64)
	lines := c.cloneSessionLines([]string{"/clone", "@agent7"})
	if len(lines) != 1 || !strings.Contains(lines[0], "sys: cloned to @agent7 (session_") {
		t.Fatalf("unexpected clone output: %#v", lines)
	}
	if c.sessionID == "" || c.sessionID == "session_clone_source" {
		t.Fatalf("expected switch to cloned session, got %q", c.sessionID)
	}
	cloned, err := s.GetSession(ctx, c.sessionID)
	if err != nil {
		t.Fatalf("get cloned session: %v", err)
	}
	if cloned.ParentSessionID != "session_clone_source" || cloned.Title != "@agent7" || cloned.State["forked_from"] != "session_clone_source" {
		t.Fatalf("unexpected cloned session: %#v", cloned)
	}
	msgs, err := s.ListMessages(ctx, cloned.ID)
	if err != nil {
		t.Fatalf("list cloned messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "hello" || msgs[0].Payload["forked_from_message_id"] != "msg_clone_source" {
		t.Fatalf("unexpected cloned messages: %#v", msgs)
	}
}

func TestResumeLinesListsAndSwitchesRecentSessions(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	allocA := gisession.AllocateDefaultSession("agent", "gi", "default", "session_resume_a")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_resume_a", "", "Alpha", map[string]any{"status": "idle"}, &allocA.Scope, allocA.SessionAliases); err != nil {
		t.Fatalf("create session a: %v", err)
	}
	allocB := gisession.AllocateDefaultSession("agent", "gi", "default", "session_resume_b")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_resume_b", "", "Beta", map[string]any{"status": "queued"}, &allocB.Scope, allocB.SessionAliases); err != nil {
		t.Fatalf("create session b: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_resume_b", "session_resume_b", "user", "hello", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_resume_a", cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, transcriptRef: gotui.NewRef(), draftLineIndex: -1}
	c.eventCh = make(chan map[string]any, 64)
	c.topicEventCh = make(chan topics.Envelope, 64)
	listed := strings.Join(c.resumeLines([]string{"/resume"}), "\n")
	for _, want := range []string{"resume: recent sessions", "Beta (session_resume_b)", "messages=1", "resume: use /resume <index|session_id>"} {
		if !strings.Contains(listed, want) {
			t.Fatalf("resume list missing %q:\n%s", want, listed)
		}
	}
	lines := c.resumeLines([]string{"/resume", "session_resume_b"})
	if len(lines) != 1 || !strings.Contains(lines[0], "sys: resumed @agent (session_resume_b)") {
		t.Fatalf("unexpected resume output: %#v", lines)
	}
	if c.sessionID != "session_resume_b" {
		t.Fatalf("expected switch to session_resume_b, got %q", c.sessionID)
	}
	lines = c.resumeLines([]string{"/resume", "session_resume_a"})
	if len(lines) != 1 || !strings.Contains(lines[0], "session_resume_a") || c.sessionID != "session_resume_a" {
		t.Fatalf("unexpected resume by id output=%#v session=%q", lines, c.sessionID)
	}
}

func TestNameSessionCommandRenamesCurrentSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_rename", "@agent", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s, sessionID: "session_rename"}
	lines := c.nameSessionLines("/name Project Alpha", []string{"/name", "Project", "Alpha"})
	if len(lines) != 1 || lines[0] != "sys: session renamed to Project Alpha" {
		t.Fatalf("unexpected rename output: %#v", lines)
	}
	reloaded, err := s.GetSession(ctx, "session_rename")
	if err != nil {
		t.Fatalf("reload session: %v", err)
	}
	if reloaded.Title != "Project Alpha" {
		t.Fatalf("unexpected title: %q", reloaded.Title)
	}
	usage := c.nameSessionLines("/name", []string{"/name"})
	if len(usage) != 1 || usage[0] != "sys: usage /name <name>" {
		t.Fatalf("unexpected usage output: %#v", usage)
	}
}

func TestSessionLinesShowCurrentSessionRuntimeSummary(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_info")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_info", "", "@agent", map[string]any{"status": "running", "model": "m1", "provider": "p1", "thinking_level": "high", "queue_count": 2}, &alloc.Scope, alloc.SessionAliases); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_session_info", "session_info", "user", "hello", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_queued", "session_info", "queued", "queued prompt", nil); err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_info", "", "user", "steer", nil, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_active", "session_info", "running", "active prompt", nil); err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "session_info", "turn_active", "worker", "claim"); err != nil || !ok {
		t.Fatalf("claim active turn: ok=%v err=%v", ok, err)
	}
	c := &chatTUI{store: s, sessionID: "session_info", cfg: config.RuntimeConfig{DefaultModel: "fallback", DefaultProvider: "fallback", DefaultThinkingLevel: "low"}, running: true}
	joined := strings.Join(c.sessionLines(), "\n")
	for _, want := range []string{"session: current", "id=session_info", "title=@agent agent=@agent parent=root", "messages=1 turns=2 queued_turns=1 steering=1", "status=running running=true active_turn=turn_active (claimed)", "model=m1 provider=p1 thinking=high"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("session summary missing %q:\n%s", want, joined)
		}
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

func TestModelCommandListsAndSelectsEnabledModels(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, sessionID: "session_models", cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultProvider: "ollama", DefaultModel: "qwen3:latest", DefaultThinkingLevel: "medium", EnabledModels: []string{"qwen3:latest", "ollama/gemma4:latest"}}}
	listed := strings.Join(c.modelCommand([]string{"/model"}), "\n")
	for _, want := range []string{"model: current=qwen3:latest", "provider=ollama thinking=medium", "*1. qwen3:latest", "  2. ollama/gemma4:latest", "select with /model <index>"} {
		if !strings.Contains(listed, want) {
			t.Fatalf("model list missing %q:\n%s", want, listed)
		}
	}
	lines := c.modelCommand([]string{"/model", "2"})
	if c.cfg.DefaultModel != "ollama/gemma4:latest" || len(lines) == 0 || !strings.Contains(lines[0], "model set to ollama/gemma4:latest") {
		t.Fatalf("index selection failed cfg=%#v lines=%#v", c.cfg, lines)
	}
}

func TestCycleThinkingUsesKnownLevels(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, sessionID: "session_cycle_thinking", cfg: config.RuntimeConfig{DefaultThinkingLevel: "low"}}
	c.cycleThinking(1)
	if c.cfg.DefaultThinkingLevel != "medium" || !strings.Contains(strings.Join(c.transcript, "\n"), "thinking set to medium") {
		t.Fatalf("next thinking failed cfg=%#v transcript=%#v", c.cfg, c.transcript)
	}
	c.cycleThinking(-1)
	if c.cfg.DefaultThinkingLevel != "low" {
		t.Fatalf("previous thinking failed cfg=%#v", c.cfg)
	}
}

func TestCycleModelUsesEnabledModels(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, sessionID: "session_cycle_model", cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultModel: "a", EnabledModels: []string{"a", "b", "c"}}}
	c.cycleModel(1)
	if c.cfg.DefaultModel != "b" || !strings.Contains(strings.Join(c.transcript, "\n"), "model set to b") {
		t.Fatalf("next model failed cfg=%#v transcript=%#v", c.cfg, c.transcript)
	}
	c.cycleModel(-1)
	if c.cfg.DefaultModel != "a" {
		t.Fatalf("previous model failed cfg=%#v", c.cfg)
	}
}

func TestScopedModelsCommandManagesEnabledModels(t *testing.T) {
	root := t.TempDir()
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultProvider: "ollama", DefaultModel: "a", DefaultThinkingLevel: "medium", EnabledModels: []string{"a"}}}
	lines := strings.Join(c.scopedModelsCommand([]string{"/scoped-models", "add", "b"}), "\n")
	if !strings.Contains(lines, "scoped models updated (2 enabled)") || !containsString(c.cfg.EnabledModels, "b") {
		t.Fatalf("add scoped model failed lines=%s cfg=%#v", lines, c.cfg)
	}
	lines = strings.Join(c.scopedModelsCommand([]string{"/scoped-models", "remove", "1"}), "\n")
	if c.cfg.DefaultModel != "b" || len(c.cfg.EnabledModels) != 1 || c.cfg.EnabledModels[0] != "b" {
		t.Fatalf("remove/default update failed lines=%s cfg=%#v", lines, c.cfg)
	}
	c.scopedModelsCommand([]string{"/scoped-models", "set", "c", "d", "c"})
	if got := strings.Join(c.cfg.EnabledModels, ","); got != "c,d" {
		t.Fatalf("set/dedupe scoped models = %q", got)
	}
	cfg := config.Load(root)
	if got := strings.Join(cfg.EnabledModels, ","); got != "c,d" {
		t.Fatalf("persisted scoped models = %q", got)
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

func TestRenderMessageLinesFormatsCodeBlocksWithLineCount(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	lines := c.renderMessageLines(store.Message{Role: "assistant", Content: "```go\nfmt.Println(1)\nfmt.Println(2)\n```", Payload: map[string]any{"kind": "chat"}}, 80)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Neo: [code:go] 2 lines", "fmt.Println(1)", "fmt.Println(2)"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("code block render missing %q:\n%s", want, joined)
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
	multilineToolLine := c.renderMessageLine(store.Message{Role: "tool_result", Content: "first line\nsecond line\nthird line", Payload: map[string]any{"kind": "tool_result", "tool_name": "shell", "is_error": true}})
	if multilineToolLine != "tool[shell/error]: 3 lines · first line" {
		t.Fatalf("multiline tool line = %q", multilineToolLine)
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

func TestMultilineInputPlaceholderShowsFocusState(t *testing.T) {
	inp := newMultilineInput(40, "Send a message…", nil, nil)
	lines := inp.renderLines()
	if len(lines) != 1 || lines[0].text != "○ Send a message…" || !lines[0].placeholder {
		t.Fatalf("blurred placeholder lines = %#v", lines)
	}
	inp.Focus()
	inp.blink = true
	lines = inp.renderLines()
	if len(lines) != 1 || lines[0].text != "● ▌Send a message…" || !lines[0].placeholder {
		t.Fatalf("focused placeholder lines = %#v", lines)
	}
}

func TestMultilineInputHelpLineFollowsFocusAndText(t *testing.T) {
	inp := newMultilineInput(40, "Send a message…", nil, nil)
	if got := inp.helpLine(); got != "" {
		t.Fatalf("blurred helpLine = %q", got)
	}
	inp.Focus()
	if got := inp.helpLine(); got != "Enter send · Alt+Enter queue · Alt+Up restore · Shift+Enter newline" {
		t.Fatalf("empty focused helpLine = %q", got)
	}
	inp.SetText("hello")
	if got := inp.helpLine(); got != "Enter send · Alt+Enter queue · Alt+Up restore · Shift+Enter newline · Esc blur" {
		t.Fatalf("text focused helpLine = %q", got)
	}
}

func TestMultilineInputWordMovementAndDeletion(t *testing.T) {
	inp := newMultilineInput(80, "", nil, nil)
	inp.SetText("alpha beta  gamma")
	inp.moveEnd()
	inp.moveWordLeft()
	if inp.cursorPos != len([]rune("alpha beta  ")) {
		t.Fatalf("word-left cursor=%d", inp.cursorPos)
	}
	inp.moveWordLeft()
	if inp.cursorPos != len([]rune("alpha ")) {
		t.Fatalf("second word-left cursor=%d", inp.cursorPos)
	}
	inp.moveWordRight()
	if inp.cursorPos != len([]rune("alpha beta  ")) {
		t.Fatalf("word-right cursor=%d", inp.cursorPos)
	}
	inp.deleteWordBackward()
	if got := inp.Text(); got != "alpha gamma" {
		t.Fatalf("delete word backward = %q", got)
	}
	inp.deleteWordForward()
	if got := inp.Text(); got != "alpha " {
		t.Fatalf("delete word forward = %q", got)
	}
}

func TestMultilineInputLineDeletion(t *testing.T) {
	inp := newMultilineInput(80, "", nil, nil)
	inp.SetText("alpha beta gamma")
	inp.cursorPos = len([]rune("alpha beta"))
	inp.deleteToLineStart()
	if got := inp.Text(); got != " gamma" || inp.cursorPos != 0 {
		t.Fatalf("delete to line start text=%q cursor=%d", got, inp.cursorPos)
	}
	inp.SetText("alpha beta gamma")
	inp.cursorPos = len([]rune("alpha beta"))
	inp.deleteToLineEnd()
	if got := inp.Text(); got != "alpha beta" {
		t.Fatalf("delete to line end = %q", got)
	}
}

func TestMultilineInputUndoAndYank(t *testing.T) {
	inp := newMultilineInput(80, "", nil, nil)
	inp.SetText("alpha beta")
	inp.deleteWordBackward()
	if got := inp.Text(); got != "alpha " {
		t.Fatalf("after delete word = %q", got)
	}
	inp.undo()
	if got := inp.Text(); got != "alpha beta" {
		t.Fatalf("after undo = %q", got)
	}
	inp.moveHome()
	inp.yank()
	if got := inp.Text(); got != "betaalpha beta" {
		t.Fatalf("after yank = %q", got)
	}
}

func TestMultilineInputCursorRenderingWithinText(t *testing.T) {
	inp := newMultilineInput(40, "", nil, nil)
	inp.SetText("abcd")
	inp.Focus()
	inp.blink = true
	inp.cursorPos = 2
	lines := inp.renderLines()
	if len(lines) != 1 || lines[0].text != "ab▌cd" {
		t.Fatalf("cursor render lines = %#v", lines)
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
	for _, want := range []string{"Session:", "Agent:", "Model:", "Provider:", "Messages:", "Queued:", "Steering:"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("responsive summary missing %q:\n%s", want, joined)
		}
	}
}

func TestStatusLineClassifiesCommonStates(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, status: "Neo · bootstrap"}
	if got := c.statusLine(); got != "● idle · Neo · bootstrap" {
		t.Fatalf("idle status line = %q", got)
	}
	c.running = true
	c.status = "Running: read"
	if got := c.statusLine(); got != "▶ running · Running: read" {
		t.Fatalf("running status line = %q", got)
	}
	c.running = false
	checks := map[string]string{
		"Queued follow-up":   "◷ queued · Queued follow-up",
		"Tool failed: shell": "◆ tool · Tool failed: shell",
		"Hook denied read":   "◇ hook · Hook denied read",
		"Compacted context":  "◌ compact · Compacted context",
	}
	for status, want := range checks {
		c.status = status
		if got := c.statusLine(); got != want {
			t.Fatalf("statusLine(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestFooterTextContainsStableHints(t *testing.T) {
	c := &chatTUI{}
	footer := c.footerTextForWidth(100)
	for _, want := range []string{"Hints:", "/help", "/tools", "Tab focus", "F2/F3 history", "Ctrl-D quit"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("footer missing %q: %s", want, footer)
		}
	}
	narrow := c.footerTextForWidth(60)
	for _, want := range []string{"Hints:", "/help", "Enter send", "Ctrl-D quit"} {
		if !strings.Contains(narrow, want) {
			t.Fatalf("narrow footer missing %q: %s", want, narrow)
		}
	}
	if strings.Contains(narrow, "/tools") || strings.Contains(narrow, "F2/F3") {
		t.Fatalf("narrow footer should stay compact: %s", narrow)
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
