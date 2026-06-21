package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func transcriptLastBlockMeta(t *testing.T, lines []string) transcriptBlockMeta {
	t.Helper()
	for i := len(lines) - 1; i >= 0; i-- {
		if meta, ok := parseTranscriptBlockMarker(lines[i]); ok {
			return meta
		}
	}
	t.Fatalf("no transcript block marker found in %#v", lines)
	return transcriptBlockMeta{}
}

func transcriptContainsBody(lines []string, want string) bool {
	for _, line := range lines {
		if strings.Contains(line, want) {
			return true
		}
	}
	return false
}

func transcriptContainsBlockKind(lines []string, kind string) bool {
	for _, line := range lines {
		if meta, ok := parseTranscriptBlockMarker(line); ok && meta.Kind == kind {
			return true
		}
	}
	return false
}

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

func TestHandleMouseWheelUsesTranscriptRegionScrollEvent(t *testing.T) {
	transcript := gotui.New(
		gotui.WithWidth(20),
		gotui.WithHeight(3),
		gotui.WithScrollable(gotui.ScrollVertical),
		gotui.WithScrollOffset(0, 2),
	)
	for _, line := range []string{"1", "2", "3", "4", "5", "6"} {
		transcript.AddChild(gotui.New(gotui.WithWidth(20), gotui.WithHeight(1), gotui.WithText(line)))
	}
	buf := gotui.NewBuffer(20, 10)
	transcript.Render(buf, 20, 10)

	c := &chatTUI{
		transcript:       []string{"1", "2", "3", "4", "5", "6"},
		transcriptScroll: 2,
		transcriptRef:    gotui.NewRef(),
		transcriptRegion: transcript,
	}
	c.transcriptRef.Set(transcript)

	if !c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseWheelUp, Action: gotui.MousePress, X: 0, Y: 0}) {
		t.Fatal("expected transcript-region wheel event to be consumed")
	}
	if c.transcriptScroll != 1 {
		t.Fatalf("expected transcript scroll to sync from standard scroll event, got %d", c.transcriptScroll)
	}
}

func TestVisibleTranscriptDoesNotMutateDraftLineIndex(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{ScrollbackLimit: 3}, transcript: []string{"1", "2", "3", "4"}, draftLineIndex: 3}
	_ = c.visibleTranscript()
	if c.draftLineIndex != 3 {
		t.Fatalf("expected visibleTranscript to leave draftLineIndex unchanged, got %d", c.draftLineIndex)
	}
}

func TestInputChangeScrollsTranscriptToBottom(t *testing.T) {
	c := &chatTUI{transcript: []string{"1", "2", "3", "4", "5", "6", "7", "8"}, transcriptScroll: 0}
	c.ensureInput()
	c.input.SetText("hello")
	if !c.stickToBottom {
		t.Fatal("expected typing to restore stick-to-bottom mode")
	}
	if c.transcriptScroll != 4 {
		t.Fatalf("expected typing to scroll transcript to bottom, got %d", c.transcriptScroll)
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
	allocA := gisession.AllocateDefaultSession(defaultTUIAgentID, "gi", "default", "session_main_a")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_main_a", "", "@"+defaultTUIAgentID, map[string]any{"model": "bootstrap", "status": "idle"}, &allocA.Scope, allocA.SessionAliases); err != nil {
		t.Fatalf("create session a: %v", err)
	}
	allocB := gisession.AllocateDefaultSession(defaultTUIAgentID, "gi", "default", "session_main_b")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_main_b", "", "@"+defaultTUIAgentID, map[string]any{"model": "bootstrap", "status": "idle"}, &allocB.Scope, allocB.SessionAliases); err != nil {
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

func TestInitialSessionIDCreatesAgentInsteadOfReusingUnrelatedSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_unrelated", "Unrelated", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create unrelated session: %v", err)
	}
	sessionID, err := initialSessionID(ctx, s)
	if err != nil {
		t.Fatalf("initial session id: %v", err)
	}
	if sessionID != "" {
		t.Fatalf("expected no initial session so TUI creates @agent, got %q", sessionID)
	}
}

func TestHandleEventStreamsDraftIntoTranscript(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": "hello"})
	if c.status != "" || len(c.transcript) != 1 || c.transcript[0] != "Neo: hello" {
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
	if !c.running || c.status != "" || len(c.transcript) != 1 || c.transcript[0] != "Neo: hello" {
		t.Fatalf("turn.draft rendering = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
}

func TestHandleTopicEventTurnThoughtRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.thought", Payload: map[string]any{"delta": "pondering"}})
	if !c.running || c.status != "" || !transcriptContainsBlockKind(c.transcript, "thought") || !transcriptContainsBody(c.transcript, "pondering") {
		t.Fatalf("turn.thought rendering = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
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
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.running = false
	c.handleTopicEvent(topics.Envelope{Topic: "turn.status", Payload: map[string]any{"title": "Thinking…", "status": "running"}})
	if !c.running || c.status != "" || !transcriptContainsBlockKind(c.transcript, "thinking_indicator") {
		t.Fatalf("turn.status rendering = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "turn.status", Payload: map[string]any{"status": "idle"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != 0 || len(c.transcript) != 1 || c.transcript[0] != "Neo: partial" {
		t.Fatalf("turn.status idle cleanup should preserve streamed draft = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
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
	if c.status != "" {
		t.Fatalf("steering should not update status, got %q", c.status)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.subturn", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "subturn_created", "child_turn_id": "turn_child"}})
	meta := transcriptLastBlockMeta(t, c.transcript)
	if meta.Kind != "subturn" || meta.Title != "Sub-turn started" || !transcriptContainsBody(c.transcript, "turn=turn_child") {
		t.Fatalf("unexpected subturn created transcript meta=%#v transcript=%#v", meta, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.subturn", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "subturn_status", "child_turn_id": "turn_child", "status": "completed"}})
	meta = transcriptLastBlockMeta(t, c.transcript)
	if meta.Kind != "subturn" || meta.Title != "Sub-turn update" || !transcriptContainsBody(c.transcript, "status=completed") {
		t.Fatalf("unexpected subturn status transcript meta=%#v transcript=%#v", meta, c.transcript)
	}
}

func TestHandleTopicEventCompactionAndRoutingRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "session.compaction", Timestamp: time.Now().UTC(), Payload: map[string]any{"messages_before": 10, "messages_after": 4, "tokens_before": 1234}})
	meta := transcriptLastBlockMeta(t, c.transcript)
	if c.status != "Compacted context" || meta.Kind != "compact" || meta.Title != "Context compacted" || !transcriptContainsBody(c.transcript, "messages=10→4") {
		t.Fatalf("compaction topic status/transcript = %q %#v meta=%#v", c.status, c.transcript, meta)
	}
	before := len(c.transcript)
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.routing", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "routing_decision", "target_agent_id": "agent1", "target_session_id": "session_child"}})
	if c.status != "routed to @agent1" || len(c.transcript) != before {
		t.Fatalf("routing decision should update status only, status=%q transcript=%#v", c.status, c.transcript)
	}
}

func TestHandleTopicEventInboundWorkRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "inbound_work_enqueued", "source_kind": "ipc", "status": "queued"}})
	if c.status != "inbound work queued (ipc) [queued]" {
		t.Fatalf("inbound work enqueue status = %q transcript=%#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "inbound_work_retry_scheduled", "source_kind": "ipc", "status": "retry", "attempt_count": 2}})
	if c.status != "inbound work retry scheduled (ipc) attempt 2 [retry]" {
		t.Fatalf("inbound work retry status = %q transcript=%#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "inbound_work_requeued", "source_kind": "ipc", "status": "queued"}})
	if c.status != "inbound work requeued (ipc) [queued]" {
		t.Fatalf("inbound work requeued status = %q transcript=%#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "inbound_work_discarded", "source_kind": "ipc", "status": "discarded"}})
	if c.status != "inbound work discarded (ipc) [discarded]" {
		t.Fatalf("inbound work discarded status = %q transcript=%#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.inbound_work", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "inbound_work_failed", "source_kind": "ipc", "status": "failed", "error": "decode failed"}})
	if c.status != "inbound work failed (ipc) [failed]: decode failed" || !transcriptContainsBody(c.transcript, "error=decode failed") {
		t.Fatalf("inbound work failed status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.dispatcher", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "dispatcher_lease_acquired", "worker_id": "worker-1"}})
	if c.status != "inbound dispatcher lease acquired [worker-1]" {
		t.Fatalf("dispatcher lease status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.dispatcher", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "dispatcher_drain_completed", "worker_id": "worker-1", "processed_count": 3}})
	if c.status != "inbound dispatcher drain completed (3 processed) [worker-1]" || !transcriptContainsBody(c.transcript, "processed=3") {
		t.Fatalf("dispatcher drain status/transcript = %q %#v", c.status, c.transcript)
	}
}

func TestHandleTopicEventTurnAndSessionRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.running = false
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.turn", Payload: map[string]any{"type": "turn_state", "status": "running", "phase": "waiting_on_tools", "tool": "read"}})
	if !c.running || c.status != "" {
		t.Fatalf("turn state status = running=%v status=%q", c.running, c.status)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.turn", Payload: map[string]any{"type": "turn_completed", "status": "completed"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != 0 || len(c.transcript) != 1 || c.transcript[0] != "Neo: partial" {
		t.Fatalf("turn completed cleanup should preserve streamed draft = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.turn", Payload: map[string]any{"type": "turn_terminal", "status": "failed"}})
	if c.status != "Turn failed" || c.running || c.draft != "" || c.draftLineIndex != 0 || len(c.transcript) != 1 || c.transcript[0] != "Neo: partial" {
		t.Fatalf("turn terminal cleanup should preserve streamed draft = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.running = false
	c.status = ""
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.session", Payload: map[string]any{"type": "session_running", "status": "running"}})
	if !c.running || c.status != "" || !transcriptContainsBlockKind(c.transcript, "thinking_indicator") {
		t.Fatalf("session running state = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.session", Payload: map[string]any{"type": "session_idle", "status": "idle"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != 0 || len(c.transcript) != 1 || c.transcript[0] != "Neo: partial" {
		t.Fatalf("session idle cleanup should preserve streamed draft = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.session", Payload: map[string]any{"type": "session_state", "status": "queued"}})
	if c.status != "Queued" || c.running || c.draft != "" || c.draftLineIndex != 0 || len(c.transcript) != 1 || c.transcript[0] != "Neo: partial" {
		t.Fatalf("session state queued rendering should preserve streamed draft = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
	c.running = true
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.session", Payload: map[string]any{"type": "session_state", "status": "idle"}})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != 0 || len(c.transcript) != 1 || c.transcript[0] != "Neo: partial" {
		t.Fatalf("session state idle cleanup should preserve streamed draft = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
	}
}

func TestHandleTopicEventHookInvocationErrorRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.hook", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "hook_invocation", "hook": "tool_call", "tool": "grep", "error": "timed out after 1500ms"}})
	if c.status != "hook invocation error via tool_call for grep: timed out after 1500ms" {
		t.Fatalf("hook invocation status/transcript = %q %#v", c.status, c.transcript)
	}
	meta := transcriptLastBlockMeta(t, c.transcript)
	if meta.Kind != "hook" || meta.Status != "error" || !transcriptContainsBody(c.transcript, "error=timed out after 1500ms") {
		t.Fatalf("unexpected hook block meta=%#v transcript=%#v", meta, c.transcript)
	}
}

func TestHandleTopicEventHookDecisionRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.hook", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "hook_modify", "hook": "tool_call", "tool": "grep"}})
	if c.status != "hook modified via tool_call for grep" {
		t.Fatalf("hook modify status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.hook", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "hook_respond", "hook": "tool_call", "tool": "grep"}})
	if c.status != "hook responded directly via tool_call for grep" {
		t.Fatalf("hook respond status/transcript = %q %#v", c.status, c.transcript)
	}
}

func TestHandleTopicEventStatusRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.running = false
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.tool", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "tool_started", "tool": "read", "turn_id": "turn_1", "tool_call_id": "call_1"}})
	if !c.running || c.status != "" || !transcriptContainsBlockKind(c.transcript, "tool") {
		t.Fatalf("tool started status = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.tool", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "tool_skipped", "tool": "shell", "reason": "queued user steering message", "turn_id": "turn_1", "tool_call_id": "call_2"}})
	if c.status != "" || !transcriptContainsBody(c.transcript, "reason=queued user steering message") {
		t.Fatalf("tool skipped status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.hook", Timestamp: time.Now().UTC(), Payload: map[string]any{"type": "hook_deny", "hook": "approve_tool", "tool": "shell", "reason": "tool not approved"}})
	if c.status != "hook denied via approve_tool for shell: tool not approved" || !transcriptContainsBody(c.transcript, "reason=tool not approved") {
		t.Fatalf("hook deny status/transcript = %q %#v", c.status, c.transcript)
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
	if !c.running || c.status != "" || !transcriptContainsBlockKind(c.transcript, "thinking_indicator") {
		t.Fatalf("agent_status without app = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
	c.draft = "partial"
	c.transcript = []string{"Neo: partial"}
	c.draftLineIndex = 0
	c.handleEvent(map[string]any{"type": "agent_status"})
	if c.status != "Neo · bootstrap" || c.running || c.draft != "" || c.draftLineIndex != 0 || len(c.transcript) != 1 || c.transcript[0] != "Neo: partial" {
		t.Fatalf("agent_status idle cleanup should preserve streamed draft = status=%q running=%v draft=%q draftLineIndex=%d transcript=%#v", c.status, c.running, c.draft, c.draftLineIndex, c.transcript)
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
	c.handleEvent(map[string]any{"type": "routing_decision", "target_agent_id": "agent1", "target_session": "session_child"})
	if c.status != "" || len(c.transcript) != 0 || c.draft != "" {
		t.Fatalf("expected topic-native path to suppress duplicate legacy routing events, got status=%q draft=%q transcript=%#v", c.status, c.draft, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "agent_status", "title": "Thinking…"})
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": "hello"})
	c.handleEvent(map[string]any{"type": "agent_thought_delta", "delta": "pondering"})
	c.handleEvent(map[string]any{"type": "new_post", "data": map[string]any{"content": "hello world"}})
	if c.status != "Neo · bootstrap" || c.draft != "" || !strings.Contains(strings.Join(c.transcript, "\n"), "Neo: hello world") {
		t.Fatalf("expected legacy stream events to remain live fallback with topic-native active, got status=%q draft=%q transcript=%#v", c.status, c.draft, c.transcript)
	}
}

func TestHandleEventStatusRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.running = false
	c.handleEvent(map[string]any{"type": "agent_thought_delta"})
	if !c.running || c.status != "" || !transcriptContainsBlockKind(c.transcript, "thinking_indicator") {
		t.Fatalf("thinking status = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
	c.running = false
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": "hello"})
	if !c.running || c.status != "" || len(c.transcript) == 0 || c.transcript[len(c.transcript)-1] != "Neo: hello" {
		t.Fatalf("draft status = running=%v status=%q transcript=%#v", c.running, c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "tool_finished", "tool": "read", "turn_id": "turn_1", "tool_call_id": "call_1"})
	if c.status != "" || !transcriptContainsBlockKind(c.transcript, "tool") {
		t.Fatalf("tool finished status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "tool_failed", "tool": "shell", "error": "boom", "turn_id": "turn_1", "tool_call_id": "call_2"})
	if c.status != "" || !transcriptContainsBody(c.transcript, "error=boom") {
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
	if c.status != "Compacted context" || !transcriptContainsBody(c.transcript, "messages=10→4") {
		t.Fatalf("compaction status/transcript = %q %#v", c.status, c.transcript)
	}
	before := len(c.transcript)
	c.handleEvent(map[string]any{"type": "routing_decision", "target_agent_id": "agent1", "target_session": "session_child"})
	if c.status != "routed to @agent1" || len(c.transcript) != before {
		t.Fatalf("legacy routing decision should update status only, status=%q transcript=%#v", c.status, c.transcript)
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

func TestSkillCommandLoadsDiscoveredSkill(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".gi", "skills", "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("Name: demo\nDescription: Demo skill\n\nUse it."), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: root}}
	lines := strings.Join(c.skillCommandLines("/skill:demo --flag"), "\n")
	for _, want := range []string{"skill:demo loaded", "args: --flag", "Name: demo", "Use it."} {
		if !strings.Contains(lines, want) {
			t.Fatalf("skill command missing %q:\n%s", want, lines)
		}
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
	out := c.localShellShortcutLines("printf hello")
	meta, ok := parseTranscriptBlockMarker(out[0])
	if !ok || meta.Kind != "bash" || meta.Title != "$ printf hello" || meta.Status != "ok" {
		t.Fatalf("unexpected bash block meta: %#v ok=%v", meta, ok)
	}
	lines := strings.Join(out, "\n")
	if !strings.Contains(lines, "│ hello") {
		t.Fatalf("local shell output missing body:\n%s", lines)
	}
}

func TestBashBlockPreviewTailAndExpand(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: t.TempDir()}}
	var sb strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&sb, "line-%02d\n", i)
	}
	lines := c.bashBlockLines("seq", sb.String(), "ok", nil, time.Now(), time.Now())
	blocks := c.buildTranscriptRenderableBlocks(lines)
	var bash transcriptRenderableBlock
	for _, b := range blocks {
		if b.Kind == "bash" {
			bash = b
		}
	}
	if bash.Kind != "bash" {
		t.Fatalf("no bash block built: %#v", blocks)
	}
	if !bash.Expandable || bash.PreviewLimit != bashPreviewLines || !bash.PreviewTail {
		t.Fatalf("unexpected bash preview config: %#v", bash)
	}
	if len(bash.Body) != 40 {
		t.Fatalf("expected full body retained, got %d lines", len(bash.Body))
	}
	if bash.Body[len(bash.Body)-1] != "line-39" {
		t.Fatalf("unexpected tail line: %q", bash.Body[len(bash.Body)-1])
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
	if c.status != "" {
		t.Fatalf("queued follow-up should not update status, got %q", c.status)
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

func TestCommandPaletteLinesFilterCommands(t *testing.T) {
	c := &chatTUI{}
	all := strings.Join(c.commandPaletteLines(""), "\n")
	for _, want := range []string{"commands: palette", "/session", "/skill:name [args]", "/copy [--osc52|--native|--auto|--fallback]", "!!cmd"} {
		if !strings.Contains(all, want) {
			t.Fatalf("palette missing %q:\n%s", want, all)
		}
	}
	filtered := strings.Join(c.commandPaletteLines("model"), "\n")
	if !strings.Contains(filtered, "/model [name|index]") || strings.Contains(filtered, "/session") {
		t.Fatalf("filtered palette mismatch:\n%s", filtered)
	}
}

func TestHelpLinesAreGroupedAndDiscoverCoreCommands(t *testing.T) {
	c := &chatTUI{}
	joined := strings.Join(c.helpLines(), "\n")
	for _, want := range []string{
		"help",
		"enter send",
		"/commands",
		"/model",
		"/session",
		"/where",
		"/attach",
		"!cmd",
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
	for _, want := range []string{"reload: config refreshed", "provider=test model=model-a thinking=high", "discovery: skills=0 tools=0", "extensions: discovered=0 mounted=0", "reload refreshes config and skill/tool discovery safely"} {
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

func TestOSC52SequenceEncodesAndCapsPayload(t *testing.T) {
	seq, err := osc52Sequence("hello")
	if err != nil {
		t.Fatal(err)
	}
	if seq != "\x1b]52;c;aGVsbG8=\x07" {
		t.Fatalf("unexpected OSC 52 sequence: %q", seq)
	}
	if _, err := osc52Sequence(strings.Repeat("x", osc52PayloadLimit+1)); err == nil {
		t.Fatal("expected payload cap error")
	}
}

func TestSelectNativeClipboardHelper(t *testing.T) {
	available := map[string]bool{"wl-copy": true, "pbcopy": true}
	look := func(name string) (string, error) {
		if available[name] {
			return "/usr/bin/" + name, nil
		}
		return "", os.ErrNotExist
	}
	if helper, ok := selectNativeClipboardHelper("linux", true, false, look); !ok || helper.name != "wl-copy" {
		t.Fatalf("expected wl-copy, got %#v ok=%v", helper, ok)
	}
	if helper, ok := selectNativeClipboardHelper("darwin", false, false, look); !ok || helper.name != "pbcopy" {
		t.Fatalf("expected pbcopy, got %#v ok=%v", helper, ok)
	}
	if _, ok := selectNativeClipboardHelper("linux", false, false, func(string) (string, error) { return "", os.ErrNotExist }); ok {
		t.Fatal("expected no helper")
	}
}

func TestCopyLastAssistantLinesOSC52OptInDoesNotStoreEscapeInTranscript(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_copy_osc", "@agent", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "msg_copy_osc", "session_copy_osc", "assistant", "answer", nil); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	c := &chatTUI{store: s, sessionID: "session_copy_osc", osc52Writer: &out}
	lines := strings.Join(c.copyLastAssistantLines("--osc52"), "\n")
	if !strings.Contains(out.String(), "\x1b]52;c;") {
		t.Fatalf("OSC 52 sequence not written: %q", out.String())
	}
	if strings.Contains(lines, "\x1b]52") || !strings.Contains(lines, "copy: sent 6 chars using OSC 52") {
		t.Fatalf("unexpected copy lines: %q", lines)
	}
}

func TestCopyLastAssistantLinesNativeOptInUsesInjectedRunner(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_copy_native", "@agent", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "msg_copy_native", "session_copy_native", "assistant", "answer", nil); err != nil {
		t.Fatal(err)
	}
	runContent := ""
	c := &chatTUI{store: s, sessionID: "session_copy_native", clipboardLookPath: func(name string) (string, error) {
		if name == "clip.exe" {
			return "/usr/bin/clip.exe", nil
		}
		return "", os.ErrNotExist
	}, clipboardRun: func(ctx context.Context, name string, args []string, content string) error {
		runContent = content
		return nil
	}}
	lines := strings.Join(c.copyLastAssistantLines("--native"), "\n")
	if runContent != "answer" || !strings.Contains(lines, "native clipboard helper") {
		t.Fatalf("native copy failed content=%q lines=%s", runContent, lines)
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
	if strings.Contains(joined, "\x1b]52") {
		t.Fatalf("copy fallback should not emit OSC 52 escape sequences: %q", joined)
	}
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
	if len(lines) == 0 || !strings.Contains(lines[0], "model: ollama/gemma4:latest") {
		t.Fatalf("unexpected model command output: %#v", lines)
	}
	cfg := config.Load(root)
	if cfg.DefaultProvider != "ollama" || cfg.DefaultModel != "ollama/gemma4:latest" {
		t.Fatalf("unexpected persisted model config: %#v", cfg)
	}
}

func TestCompactOutputShortensDenseModelAndResumeLines(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(context.Background(), "session_very_long_identifier_for_compact_mode", "Very Long Session Title For Compact Mode", map[string]any{"status": "idle"}); err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{store: s, sessionID: "session_very_long_identifier_for_compact_mode", outputWidth: 60, cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultProvider: "very-long-provider-name", DefaultModel: "provider/very-long-model-name-that-would-wrap", EnabledModels: []string{"provider/very-long-model-name-that-would-wrap"}}}
	resume := strings.Join(c.resumeLines([]string{"/resume"}), "\n")
	if !strings.Contains(resume, "m=0 t=0") || !strings.Contains(resume, "…_mode]") {
		t.Fatalf("compact resume missing dense markers:\n%s", resume)
	}
	models := strings.Join(c.modelListLines(), "\n")
	if !strings.Contains(models, "provider/very-long-model-name") || !strings.Contains(models, "…") {
		t.Fatalf("compact model output unexpected:\n%s", models)
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
	for _, want := range []string{"model qwen3:latest · medium · ollama", "› 1  qwen3:latest", "  2  ollama/gemma4:latest", "enabled: 2 · /scoped-models list to manage pinned models", "/model <n> to switch · ctrl-l cycles enabled models"} {
		if !strings.Contains(listed, want) {
			t.Fatalf("model list missing %q:\n%s", want, listed)
		}
	}
	lines := c.modelCommand([]string{"/model", "2"})
	if c.cfg.DefaultModel != "ollama/gemma4:latest" || len(lines) == 0 || !strings.Contains(lines[0], "model: ollama/gemma4:latest") {
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

func TestModelListIncludesAuthBackedProviderModels(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".pi", "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".pi", "agent", "auth.json"), []byte(`{"github-copilot":{"type":"oauth","refresh":"token","access":"token","expires":9999999999999}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, sessionID: "session_runtime_models", cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultProvider: "opencode-zen", DefaultModel: "opencode-zen/minimax-m2.5-free", DefaultThinkingLevel: "low", EnabledModels: []string{"opencode-zen/minimax-m2.5-free"}}}
	listed := strings.Join(c.modelListLines(), "\n")
	if !strings.Contains(listed, "github-copilot/") {
		t.Fatalf("expected auth-backed github-copilot models in list:\n%s", listed)
	}
	if !strings.Contains(listed, "opencode-zen/minimax-m2.5-free") {
		t.Fatalf("expected existing enabled model in list:\n%s", listed)
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
	if c.cfg.DefaultModel != "b" || !strings.Contains(strings.Join(c.transcript, "\n"), "model: b") {
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
	for _, want := range []string{"settings: runtime", "workspace: /tmp/ws", "settings: model", "provider: test", "model: m", "thinking: low", "settings: editor", "scrollback_limit: 1000", "clipboard_mode:", "settings: session", "settings: discovery", "active_tools:"} {
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

func TestRenderMessageLinesProjectsInlineMarkdownWithoutMarkers(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	content := "This is **bold**, *emph*, `code`, ~~gone~~, and [a link](https://example.com).\n- [x] done\n- [ ] todo"
	lines := c.renderMessageLines(store.Message{Role: "assistant", Content: content, Payload: map[string]any{"kind": "chat"}}, 80)
	joined := strings.Join(lines, "\n")
	for _, bad := range []string{"**bold**", "*emph*", "`code`", "~~gone~~"} {
		if strings.Contains(joined, bad) {
			t.Fatalf("markdown marker %q leaked into render:\n%s", bad, joined)
		}
	}
	for _, want := range []string{"bold", "emph", "code", "gone", "a link (https://example.com)", "☑", "☐"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("markdown render missing %q:\n%s", want, joined)
		}
	}
}

func TestRenderMessageLinesPreservesInlineCodeLeadingSpaces(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	lines := c.renderMessageLines(store.Message{Role: "assistant", Content: "Run `  --flag` now", Payload: map[string]any{"kind": "chat"}}, 80)
	joined := strings.Join(lines, "\n")
	plain := stripMarkdownInlineStyleMarkers(joined)
	if strings.Contains(plain, "`") {
		t.Fatalf("inline code backticks leaked into render:\n%s", plain)
	}
	if !strings.Contains(plain, "  --flag") {
		t.Fatalf("inline code leading spaces were not preserved: raw=%q plain=%q", joined, plain)
	}
	segments := parseTUIInlineSegments(joined)
	found := false
	for _, seg := range segments {
		if seg.Code && seg.Text == "  --flag" {
			found = true
		}
	}
	if !found {
		t.Fatalf("inline code segment did not preserve leading spaces: %#v", segments)
	}
}

func TestRenderMessageLinesPreservesCodeBlockSpacing(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	lines := c.renderMessageLines(store.Message{Role: "assistant", Content: "```\nif  x  {\n    y :=  1\n}\n```", Payload: map[string]any{"kind": "chat"}}, 80)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"if  x  {", "    y :=  1"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("code block spacing not preserved for %q:\n%s", want, joined)
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
	toolLines := c.renderToolResultLines(store.Message{ID: "m1", Role: "tool_result", Content: "long tool output", Payload: map[string]any{"kind": "tool_result", "tool_name": "shell", "is_error": false}})
	if len(toolLines) != 2 {
		t.Fatalf("tool lines = %#v", toolLines)
	}
	meta, ok := parseTranscriptBlockMarker(toolLines[0])
	if !ok || meta.Kind != "tool" || meta.Title != "shell" || meta.Status != "ok" || toolLines[1] != "│ long tool output" {
		t.Fatalf("tool lines = %#v meta=%#v ok=%v", toolLines, meta, ok)
	}
	multilineToolLines := c.renderToolResultLines(store.Message{ID: "m2", Role: "tool_result", Content: "first line\nsecond line\nthird line", Payload: map[string]any{"kind": "tool_result", "tool_name": "shell", "is_error": true}})
	joined := strings.Join(multilineToolLines, "\n")
	for _, want := range []string{"│ first line", "│ second line", "│ third line"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("multiline tool lines missing %q in %#v", want, multilineToolLines)
		}
	}
	compactionLine := c.renderMessageLine(store.Message{Role: "assistant", Content: "summary", Payload: map[string]any{"kind": "compaction", "tokens_before": 1200}})
	if compactionLine != "compact: summary (tokens_before=1200)" {
		t.Fatalf("compaction line = %q", compactionLine)
	}
}

func TestBuildTranscriptRenderableBlocksCollapsesToolBlocks(t *testing.T) {
	startedAt := time.Now().Add(-1500 * time.Millisecond).UTC().Format(time.RFC3339Nano)
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}, transcriptExpanded: map[string]bool{}}
	lines := []string{encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "tool:test", Kind: "tool", Title: "shell", Status: "ok", StartedAt: startedAt, EndedAt: startedAt}), "│ first", "│ second", "│ third", "│ fourth"}
	blocks := c.buildTranscriptRenderableBlocks(lines)
	if len(blocks) != 1 {
		t.Fatalf("expected one block, got %#v", blocks)
	}
	if !blocks[0].Expandable || blocks[0].Expanded || blocks[0].Kind != "tool" || blocks[0].Subheader != "" || !strings.Contains(blocks[0].Header, "shell ·") {
		t.Fatalf("unexpected collapsed block state: %#v", blocks[0])
	}
}

func TestToolOKStatusLabelIsHidden(t *testing.T) {
	if shouldRenderTranscriptStatusLabel("tool", "ok") {
		t.Fatalf("tool ok status should not render as an inline label")
	}
	if !shouldRenderTranscriptStatusLabel("tool", "error") {
		t.Fatalf("tool error status should still render")
	}
}

func TestTranscriptBlockLifecycleAppendReplaceDelete(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	c.appendTranscriptBlock(transcriptBlockMeta{Key: "k1", Kind: "tool", Title: "first", Status: "running"}, []string{"a", "b"})
	if span, ok := c.transcriptBlockSpans["k1"]; !ok || span.BodyCount != 2 {
		t.Fatalf("append did not index block: %#v", c.transcriptBlockSpans)
	}
	// replace updates header meta and body in place (same key, no duplicate)
	c.replaceTranscriptBlock(transcriptBlockMeta{Key: "k1", Kind: "tool", Title: "first", Status: "ok"}, []string{"a", "b", "c"})
	if order := c.transcriptBlockOrder(); len(order) != 1 || order[0] != "k1" {
		t.Fatalf("replace should keep a single block: %#v", order)
	}
	if body := c.readTranscriptBlockBody("k1"); len(body) != 3 || body[2] != "c" {
		t.Fatalf("replace did not update body: %#v", body)
	}
	meta, ok := parseTranscriptBlockMarker(c.transcript[c.transcriptBlockSpans["k1"].HeaderIndex])
	if !ok || meta.Status != "ok" {
		t.Fatalf("replace did not update meta: %#v ok=%v", meta, ok)
	}
	// delete removes the block and its state
	if !c.deleteTranscriptBlock("k1") {
		t.Fatalf("delete should report removal")
	}
	if _, ok := c.transcriptBlockSpans["k1"]; ok || len(c.transcript) != 0 {
		t.Fatalf("delete left residue: spans=%#v transcript=%#v", c.transcriptBlockSpans, c.transcript)
	}
}

func TestSelectTranscriptBlockCyclesSelection(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	c.appendTranscriptBlock(transcriptBlockMeta{Key: "a", Kind: "tool", Title: "A"}, []string{"x"})
	c.appendTranscriptBlock(transcriptBlockMeta{Key: "b", Kind: "tool", Title: "B"}, []string{"y"})
	c.appendTranscriptBlock(transcriptBlockMeta{Key: "c", Kind: "tool", Title: "C"}, []string{"z"})
	c.selectedTranscriptBlock = ""
	c.selectTranscriptBlock(1)
	first := c.selectedTranscriptBlock
	if first == "" {
		t.Fatalf("select did not pick a block")
	}
	c.selectTranscriptBlock(1)
	if c.selectedTranscriptBlock == first {
		t.Fatalf("select forward did not move selection from %q", first)
	}
	c.selectTranscriptBlock(-1)
	if c.selectedTranscriptBlock != first {
		t.Fatalf("select backward did not return to %q, got %q", first, c.selectedTranscriptBlock)
	}
}

func TestErrorLinesRenderAsDedupedBlocks(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}, transcriptExpanded: map[string]bool{}}
	blocks := c.buildTranscriptRenderableBlocks([]string{
		"sys: inference error: max retries exceeded",
		"max retries exceeded",
		"Neo: after",
	})
	if len(blocks) != 2 {
		t.Fatalf("expected deduped error block plus assistant line, got %#v", blocks)
	}
	if blocks[0].Kind != "error" || blocks[0].Header != "Error" || len(blocks[0].Body) != 1 || !strings.Contains(blocks[0].Body[0], "max retries exceeded") {
		t.Fatalf("unexpected error block: %#v", blocks[0])
	}
}

func TestToggleSelectedTranscriptBlockExpandsBlock(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}, transcriptExpanded: map[string]bool{}, transcript: []string{
		encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "tool:test", Kind: "tool", Title: "shell", Status: "ok"}),
		"│ first",
		"│ second",
		"│ third",
	}}
	c.reindexTranscriptBlocks()
	c.selectedTranscriptBlock = "tool:test"
	c.toggleSelectedTranscriptBlock()
	blocks := c.buildTranscriptRenderableBlocks(c.transcript)
	if len(blocks) != 1 || !blocks[0].Expanded {
		t.Fatalf("expected selected block to expand, got %#v", blocks)
	}
}

func TestRenderToolEventUpdatesExistingRuntimeBlock(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}}
	startedAt := time.Now().Add(-2 * time.Second).UTC()
	c.renderToolEvent(map[string]any{"type": "tool_started", "tool": "shell", "turn_id": "turn_1", "tool_call_id": "call_1", "iteration": 1}, startedAt)
	c.renderToolEvent(map[string]any{"type": "tool_failed", "tool": "shell", "turn_id": "turn_1", "tool_call_id": "call_1", "iteration": 1, "error": "boom"}, startedAt.Add(1500*time.Millisecond))
	meta := transcriptLastBlockMeta(t, c.transcript)
	if meta.Kind != "tool" || meta.Status != "error" || !transcriptContainsBody(c.transcript, "error=boom") {
		t.Fatalf("unexpected runtime tool block meta=%#v transcript=%#v", meta, c.transcript)
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
	if len(lines) != 1 || lines[0].text != "" || !lines[0].placeholder {
		t.Fatalf("blurred placeholder lines = %#v", lines)
	}
	inp.Focus()
	inp.blink = false
	lines = inp.renderLines()
	if len(lines) != 1 || lines[0].text != "▌" || !lines[0].placeholder {
		t.Fatalf("focused placeholder lines = %#v", lines)
	}
}

func TestMultilineInputHelpLineIsHidden(t *testing.T) {
	inp := newMultilineInput(40, "Send a message…", nil, nil)
	if got := inp.helpLine(); got != "" {
		t.Fatalf("blurred helpLine = %q", got)
	}
	inp.Focus()
	if got := inp.helpLine(); got != "" {
		t.Fatalf("focused empty helpLine = %q", got)
	}
	inp.SetText("hello")
	if got := inp.helpLine(); got != "" {
		t.Fatalf("focused text helpLine = %q", got)
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
	inp.blink = false
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
	if len(lines) == 0 || len(lines) > 2 {
		t.Fatalf("expected compact summary lines, got %#v", lines)
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"@agent", "opencode-zen/minimax-m2.5-free", "low", "m0/t0"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("compact summary missing %q:\n%s", want, joined)
		}
	}
}

func TestSessionSelectorOpensFiltersAndSwitches(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_alpha", "@alpha", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create alpha: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_beta", "@beta", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create beta: %v", err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_alpha", cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, transcriptRef: gotui.NewRef(), draftLineIndex: -1}
	c.eventCh = make(chan map[string]any, 64)
	c.topicEventCh = make(chan topics.Envelope, 64)
	c.openSessionMenu()
	if !c.modelMenuOpen || c.modelMenuKind != "session" {
		t.Fatalf("session menu not open: open=%v kind=%q", c.modelMenuOpen, c.modelMenuKind)
	}
	if len(c.modelMenuChoices) != 2 {
		t.Fatalf("expected 2 session choices, got %#v", c.modelMenuChoices)
	}
	c.modelMenuTypeRune('b')
	c.modelMenuTypeRune('e')
	c.modelMenuTypeRune('t')
	if len(c.modelMenuChoices) != 1 || !strings.Contains(c.modelMenuChoices[0], "@beta") {
		t.Fatalf("filter to beta failed: %#v", c.modelMenuChoices)
	}
	c.acceptModelMenuSelection()
	if c.modelMenuOpen {
		t.Fatalf("menu should close after accept")
	}
	if c.sessionID != "session_beta" {
		t.Fatalf("expected switch to session_beta, got %q", c.sessionID)
	}
}

func TestModelMenuFuzzyFilter(t *testing.T) {
	all := []string{"openai/gpt-5.2", "anthropic/claude-sonnet", "opencode-zen/minimax", "openai/gpt-4.1-mini"}
	if got := filterModelMenuChoices(all, ""); len(got) != 4 {
		t.Fatalf("empty query should return all, got %#v", got)
	}
	got := filterModelMenuChoices(all, "gpt")
	if len(got) != 2 || got[0] != "openai/gpt-5.2" || got[1] != "openai/gpt-4.1-mini" {
		t.Fatalf("gpt filter unexpected: %#v", got)
	}
	if got := filterModelMenuChoices(all, "clsn"); len(got) != 0 {
		t.Fatalf("non-substring query should not match: %#v", got)
	}
	if got := filterModelMenuChoices(all, "claude sonnet"); len(got) != 1 || got[0] != "anthropic/claude-sonnet" {
		t.Fatalf("multi-token filter unexpected: %#v", got)
	}
	if got := filterModelMenuChoices(all, "zzz"); len(got) != 0 {
		t.Fatalf("no-match filter should be empty, got %#v", got)
	}
}

func TestModelMenuTypeAndBackspaceFiltersChoices(t *testing.T) {
	c := &chatTUI{modelMenuOpen: true, modelMenuAll: []string{"openai/gpt-5.2", "anthropic/claude", "openai/gpt-4.1"}}
	c.modelMenuChoices = append([]string(nil), c.modelMenuAll...)
	c.modelMenuTypeRune('g')
	c.modelMenuTypeRune('p')
	c.modelMenuTypeRune('t')
	if len(c.modelMenuChoices) != 2 {
		t.Fatalf("expected 2 gpt matches, got %#v", c.modelMenuChoices)
	}
	if c.modelMenuQuery != "gpt" || c.modelMenuSelected != 0 {
		t.Fatalf("unexpected menu state query=%q sel=%d", c.modelMenuQuery, c.modelMenuSelected)
	}
	c.modelMenuBackspace()
	c.modelMenuBackspace()
	c.modelMenuBackspace()
	if c.modelMenuQuery != "" || len(c.modelMenuChoices) != 3 {
		t.Fatalf("backspace should restore all choices, got query=%q choices=%#v", c.modelMenuQuery, c.modelMenuChoices)
	}
}

func TestExtensionToolRenderSlotControlsToolBody(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}, transcriptExpanded: map[string]bool{}}
	lines := []string{
		encodeTranscriptBlockMarker(transcriptBlockMeta{Key: "tool:1", Kind: "tool", Title: "grep", Status: "ok"}),
		"│ match 1",
		"│ match 2",
		"│ match 3",
	}
	// default: full body, expandable
	if blocks := c.buildTranscriptRenderableBlocks(lines); len(blocks) != 1 || len(blocks[0].Body) != 3 || !blocks[0].Expandable {
		t.Fatalf("default tool body unexpected: %#v", blocks)
	}
	// compact: first body line only
	c.handleTopicEvent(topics.Envelope{Topic: "extension.tool_render", Payload: map[string]any{"tool": "grep", "mode": "compact"}})
	if blocks := c.buildTranscriptRenderableBlocks(lines); len(blocks[0].Body) != 1 || blocks[0].Body[0] != "match 1" || blocks[0].Expandable {
		t.Fatalf("compact tool body unexpected: %#v", blocks[0])
	}
	// hidden: header only
	c.handleTopicEvent(topics.Envelope{Topic: "extension.tool_render", Payload: map[string]any{"tool": "grep", "mode": "hidden"}})
	if blocks := c.buildTranscriptRenderableBlocks(lines); len(blocks[0].Body) != 0 || blocks[0].Expandable {
		t.Fatalf("hidden tool body unexpected: %#v", blocks[0])
	}
	// full restores
	c.handleTopicEvent(topics.Envelope{Topic: "extension.tool_render", Payload: map[string]any{"tool": "grep", "mode": "full"}})
	if blocks := c.buildTranscriptRenderableBlocks(lines); len(blocks[0].Body) != 3 {
		t.Fatalf("full restore unexpected: %#v", blocks[0])
	}
}

func TestExtensionWidgetSlotRendersAboveEditorOnly(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap", DefaultThinkingLevel: "low", WorkspaceRoot: t.TempDir()}}
	if lines := c.extensionWidgetLines(); len(lines) != 0 {
		t.Fatalf("expected no widget lines initially, got %#v", lines)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "extension.widget", Payload: map[string]any{"key": "plan", "lines": []any{"plan: step 1", "plan: step 2"}}})
	lines := c.extensionWidgetLines()
	if len(lines) != 2 || lines[0] != "plan: step 1" {
		t.Fatalf("unexpected widget lines: %#v", lines)
	}
	// Widget slot must not write transcript rows (cannot create top chrome).
	if len(c.transcript) != 0 {
		t.Fatalf("widget slot must not write transcript rows, got %#v", c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "extension.widget", Payload: map[string]any{"key": "plan", "lines": []any{}}})
	if lines := c.extensionWidgetLines(); len(lines) != 0 {
		t.Fatalf("clearing widget should remove lines, got %#v", lines)
	}
	// text payload fallback
	c.setExtensionWidget("note", nil)
	c.handleTopicEvent(topics.Envelope{Topic: "extension.widget", Payload: map[string]any{"key": "note", "text": "a\nb"}})
	if lines := c.extensionWidgetLines(); len(lines) != 2 || lines[1] != "b" {
		t.Fatalf("text widget payload unexpected: %#v", lines)
	}
}

func TestExtensionStatusSlotAddsFooterRowsOnly(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap", DefaultThinkingLevel: "low", WorkspaceRoot: t.TempDir()}}
	base := len(c.footerLines(100))
	c.setExtensionStatus("lint", "lint: 0 errors")
	c.setExtensionStatus("build", "build ok")
	lines := c.footerLines(100)
	if len(lines) != base+2 {
		t.Fatalf("expected %d footer lines, got %d: %#v", base+2, len(lines), lines)
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "build ok") || !strings.Contains(joined, "lint: 0 errors") {
		t.Fatalf("extension status missing from footer: %#v", lines)
	}
	// sorted by key: build before lint
	if lines[base] != "build ok" {
		t.Fatalf("expected sorted extension statuses, got %#v", lines[base:])
	}
	c.handleTopicEvent(topics.Envelope{Topic: "extension.status", Payload: map[string]any{"key": "lint", "text": ""}})
	if lines := c.footerLines(100); len(lines) != base+1 {
		t.Fatalf("clearing one status should leave %d lines, got %#v", base+1, lines)
	}
	// Slot output is footer-only: setting status must not add any transcript row.
	if len(c.transcript) != 0 {
		t.Fatalf("extension status must not write transcript rows, got %#v", c.transcript)
	}
}

func TestFooterLinesExpandWithUsageAndNotice(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "provider/model", DefaultThinkingLevel: "low", WorkspaceRoot: t.TempDir()}, lastInputTokens: 1200, lastOutputTokens: 340, lastContextTokens: 5000, lastCacheRead: 2000, lastCacheWrite: 800, lastCostTotal: 0.0123}
	lines := c.footerLines(120)
	if len(lines) < 2 {
		t.Fatalf("expected at least path + stats lines, got %#v", lines)
	}
	stats := lines[1]
	for _, want := range []string{"m0/t0", "↑1K", "↓340", "R2K", "W800", "$0.012", "ctx 5K", "provider/model", "low"} {
		if !strings.Contains(stats, want) {
			t.Fatalf("stats footer missing %q: %s", want, stats)
		}
	}
	c.status = "Running: read"
	lines = c.footerLines(120)
	if len(lines) != 3 || !strings.Contains(lines[2], "Running: read") {
		t.Fatalf("expected transient notice line, got %#v", lines)
	}
}

func TestFooterNotificationTextClassifiesCommonStates(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap", DefaultThinkingLevel: "low"}, status: "Neo · bootstrap"}
	data := c.contextSummaryData()
	if got := c.footerNotificationText(data); got != "m0/t0" {
		t.Fatalf("idle footer notification = %q", got)
	}
	for _, status := range []string{"Running: read", "Queued follow-up", "Tool failed: shell", "Hook denied read", "Compacted context"} {
		c.status = status
		want := "m0/t0 · " + status
		if got := c.footerNotificationText(data); got != want {
			t.Fatalf("footerNotificationText(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestFooterContextFollowsThinkingLevel(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{DefaultModel: "provider/model", DefaultThinkingLevel: "medium", Compaction: config.CompactionSettings{ContextWindow: 20000}}, lastContextTokens: 5000}
	line := c.footerStatusLineForWidth(120)
	want := "provider/model • medium • ctx 5K/20K 25%"
	if !strings.Contains(line, want) {
		t.Fatalf("footer context should follow thinking level: got %q, want substring %q", line, want)
	}
	if strings.Contains(c.footerCountsText(c.contextSummaryData()), "ctx") {
		t.Fatalf("footer counts should not include context usage: %q", c.footerCountsText(c.contextSummaryData()))
	}
}

func TestFooterContextWindowUsesSelectedModelMetadata(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{DefaultProvider: "opencode-zen", DefaultModel: "minimax-m2.5-free", DefaultThinkingLevel: "low", Compaction: config.CompactionSettings{ContextWindow: 20000}}, lastContextTokens: 64000}
	line := c.footerStatusLineForWidth(140)
	want := "minimax-m2.5-free • low • ctx 64K/128K 50%"
	if !strings.Contains(line, want) {
		t.Fatalf("footer context should use selected model context window: got %q, want substring %q", line, want)
	}
}

func TestFooterStatusLineRefreshesCountsFromStore(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_footer_counts", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s, sessionID: "session_footer_counts", cfg: config.RuntimeConfig{DefaultModel: "bootstrap", DefaultThinkingLevel: "low"}}
	if got := c.footerStatusLineForWidth(80); !strings.Contains(got, "m0/t0") {
		t.Fatalf("initial footer missing m0/t0: %q", got)
	}
	if err := s.AddMessage(ctx, "msg_footer_counts", "session_footer_counts", "user", "hello", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_footer_counts", "session_footer_counts", "completed", "hello", nil); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if got := c.footerStatusLineForWidth(80); !strings.Contains(got, "m1/t1") {
		t.Fatalf("updated footer missing m1/t1: %q", got)
	}
	c.status = "Running: shell"
	if got := c.footerStatusLineForWidth(80); !strings.Contains(got, "m1/t1") || !strings.Contains(got, "Running: shell") {
		t.Fatalf("status footer should include dynamic counts and activity: %q", got)
	}
}

func TestFooterTextContainsStableHints(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultModel: "provider/model", DefaultThinkingLevel: "low"}}
	pathLine := c.footerPathLineForWidth(80)
	if !strings.Contains(pathLine, filepath.Base(root)) || !strings.Contains(pathLine, "(main)") {
		t.Fatalf("path footer missing workspace/branch: %s", pathLine)
	}
	statusLine := c.footerStatusLineForWidth(80)
	for _, want := range []string{"m0/t0", "provider/model", "low"} {
		if !strings.Contains(statusLine, want) {
			t.Fatalf("status footer missing %q: %s", want, statusLine)
		}
	}
	if strings.Contains(statusLine, "enter send") || strings.Contains(statusLine, "/ commands") {
		t.Fatalf("footer should be status-only: %s", statusLine)
	}
	c.status = "Running: read file with a very long path that must not wrap onto another physical footer row"
	statusLine = c.footerStatusLineForWidth(48)
	if strings.Contains(statusLine, "\n") || utf8.RuneCountInString(statusLine) > 48 {
		t.Fatalf("status footer should stay one truncated line: len=%d %q", utf8.RuneCountInString(statusLine), statusLine)
	}
	if !strings.Contains(statusLine, "provider/model") && !strings.Contains(statusLine, "…") {
		t.Fatalf("status footer should retain right-side model/truncation cue: %s", statusLine)
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

func TestExtensionCommandLinesDispatchesRegisteredCommand(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := turn.New(s)
	if _, err := e.RegisterExtensionCommand(turn.ExtensionCommandSpec{Name: "demo", Description: "Demo", Source: "test", Engine: "js"}, func(ctx context.Context, cmd turn.ExtensionCommandContext) (turn.ExtensionCommandResult, error) {
		return turn.ExtensionCommandResult{Type: "message", Lines: []string{"demo: " + cmd.Args}}, nil
	}); err != nil {
		t.Fatalf("register command: %v", err)
	}
	c := &chatTUI{engine: e, sessionID: "session_demo"}
	lines, handled := c.extensionCommandLines("/demo hello", []string{"/demo", "hello"})
	if !handled || len(lines) != 1 || lines[0] != "demo: hello" {
		t.Fatalf("extension command lines handled=%v lines=%#v", handled, lines)
	}
	palette := strings.Join(c.commandPaletteLines("demo"), "\n")
	if !strings.Contains(palette, "/demo") {
		t.Fatalf("palette missing extension command:\n%s", palette)
	}
	plugins := strings.Join(c.pluginLines(), "\n")
	if !strings.Contains(plugins, "plugins: commands: 1") || !strings.Contains(plugins, "/demo") {
		t.Fatalf("plugins missing extension command:\n%s", plugins)
	}
}

func TestStreamingDraftRendersMarkdownDynamically(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.draft", Payload: map[string]any{"delta": "# Plan\n\n- one"}})
	joined := strings.Join(c.transcript, "\n")
	if !strings.Contains(joined, "Neo: PLAN") || !strings.Contains(joined, "• one") {
		t.Fatalf("streaming markdown was not rendered dynamically:\n%s", joined)
	}
	if c.draftLineCount <= 1 {
		t.Fatalf("expected multi-line streaming draft span, got %d lines: %#v", c.draftLineCount, c.transcript)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.draft", Payload: map[string]any{"delta": "\n- two"}})
	joined = strings.Join(c.transcript, "\n")
	if strings.Count(joined, "Neo: PLAN") != 1 || !strings.Contains(joined, "• two") {
		t.Fatalf("streaming markdown replacement should update in place:\n%s", joined)
	}
	c.handleTopicEvent(topics.Envelope{Topic: "turn.response", Payload: map[string]any{"data": map[string]any{"content": "# Done\n\n- final"}}})
	joined = strings.Join(c.transcript, "\n")
	if strings.Contains(joined, "• one") || !strings.Contains(joined, "Neo: DONE") || !strings.Contains(joined, "• final") {
		t.Fatalf("final markdown should replace entire streaming span:\n%s", joined)
	}
	if c.draftLineIndex != -1 || c.draftLineCount != 0 || c.draft != "" || c.running {
		t.Fatalf("unexpected final state: running=%v draft=%q index=%d count=%d", c.running, c.draft, c.draftLineIndex, c.draftLineCount)
	}
}

func TestModelCommandOpensCursorNavigableMenu(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{DefaultProvider: "openai", DefaultModel: "openai/a", EnabledModels: []string{"openai/a", "openai/b", "openai/c"}}}
	c.handleCommand("/model")
	if !c.modelMenuOpen {
		t.Fatal("expected /model to open model menu")
	}
	if c.modelMenuSelected != 0 {
		t.Fatalf("expected current model selected, got %d", c.modelMenuSelected)
	}
	c.moveModelMenuSelection(1)
	if c.modelMenuSelected != 1 {
		t.Fatalf("expected cursor navigation to move selection, got %d", c.modelMenuSelected)
	}
	c.acceptModelMenuSelection()
	if c.modelMenuOpen {
		t.Fatal("expected model menu to close after selection")
	}
	if c.cfg.DefaultModel != "openai/b" {
		t.Fatalf("expected selected model openai/b, got %q", c.cfg.DefaultModel)
	}
}

func TestScrollbarCommandDefaultsOffAndPersists(t *testing.T) {
	root := t.TempDir()
	c := &chatTUI{cfg: config.RuntimeConfig{WorkspaceRoot: root}}
	if c.cfg.TUIScrollbar {
		t.Fatal("expected scrollbar to default off")
	}
	lines := c.scrollbarCommand([]string{"/scrollbar", "on"})
	if !c.cfg.TUIScrollbar {
		t.Fatal("expected scrollbar command to enable option")
	}
	if got := strings.Join(lines, "\n"); !strings.Contains(got, "scrollbar set to on") {
		t.Fatalf("expected confirmation, got %q", got)
	}
	cfg := config.Load(root)
	if !cfg.TUIScrollbar {
		t.Fatal("expected scrollbar setting to persist")
	}
}
