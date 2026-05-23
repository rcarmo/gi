package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/routing"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/store/queue"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func TestServerSessionPromptTurnsFlow(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "test-model", DefaultThinkingLevel: "medium"})

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"title":"@agent","agent_id":"agent"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRes.Code, createRes.Body.String())
	}
	var created struct {
		ID    string `json:"id"`
		Scope struct {
			AgentID string `json:"agent_id"`
		} `json:"scope"`
	}
	if err := json.Unmarshal(createRes.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Scope.AgentID != "agent" {
		t.Fatalf("unexpected agent id: %+v", created)
	}

	promptReq := httptest.NewRequest(http.MethodPost, "/api/sessions/"+created.ID+"/prompt", bytes.NewBufferString(`{"prompt":"hello"}`))
	promptReq.Header.Set("Content-Type", "application/json")
	promptRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(promptRes, promptReq)
	if promptRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected prompt status: %d body=%s", promptRes.Code, promptRes.Body.String())
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		turnsReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+created.ID+"/turns", nil)
		turnsRes := httptest.NewRecorder()
		srv.Handler().ServeHTTP(turnsRes, turnsReq)
		if turnsRes.Code == http.StatusOK && bytes.Contains(turnsRes.Body.Bytes(), []byte("completed")) {
			messagesReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+created.ID+"/messages", nil)
			messagesRes := httptest.NewRecorder()
			srv.Handler().ServeHTTP(messagesRes, messagesReq)
			if messagesRes.Code == http.StatusOK && bytes.Contains(messagesRes.Body.Bytes(), []byte("Gi received: hello")) && bytes.Contains(messagesRes.Body.Bytes(), []byte(`"agent_id":"agent"`)) {
				break
			}
		}
		if time.Now().After(deadline) {
			turnsReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+created.ID+"/turns", nil)
			turnsRes := httptest.NewRecorder()
			srv.Handler().ServeHTTP(turnsRes, turnsReq)
			messagesReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+created.ID+"/messages", nil)
			messagesRes := httptest.NewRecorder()
			srv.Handler().ServeHTTP(messagesRes, messagesReq)
			t.Fatalf("timed out waiting for completed turn + assistant output; turns=%d %s messages=%d %s", turnsRes.Code, turnsRes.Body.String(), messagesRes.Code, messagesRes.Body.String())
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestSessionIntrospectionEndpoint(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "test-model", DefaultThinkingLevel: "medium"})

	session, err := s.CreateSession(t.Context(), store.NowID("session"), "Demo", map[string]any{"status": "idle", "model": "test-model", "provider": "test", "thinking_level": "medium"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.AddMessage(t.Context(), store.NowID("msg"), session.ID, "user", "hello", map[string]any{"kind": "chat"}); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(t.Context(), store.NowID("turn"), session.ID, "completed", "hello", map[string]any{"model": "test-model"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+session.ID+"/introspect", nil)
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("unexpected introspect status: %d body=%s", res.Code, res.Body.String())
	}
	if !bytes.Contains(res.Body.Bytes(), []byte("message_count")) || !bytes.Contains(res.Body.Bytes(), []byte("turn_count")) || !bytes.Contains(res.Body.Bytes(), []byte("test-model")) || !bytes.Contains(res.Body.Bytes(), []byte("scope")) {
		t.Fatalf("unexpected introspect body: %s", res.Body.String())
	}
}

func TestCreateSessionMarksMainSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "test-model", DefaultThinkingLevel: "medium"})

	createReqA := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"title":"@agent","agent_id":"agent"}`))
	createReqA.Header.Set("Content-Type", "application/json")
	createResA := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createResA, createReqA)
	if createResA.Code != http.StatusCreated {
		t.Fatalf("unexpected first create status: %d body=%s", createResA.Code, createResA.Body.String())
	}
	var createdA struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createResA.Body.Bytes(), &createdA); err != nil {
		t.Fatalf("decode first create response: %v", err)
	}
	mainSessionID, err := s.ResolveMainSessionID(t.Context(), "agent", "gi", "default")
	if err != nil {
		t.Fatalf("resolve main session id after first create: %v", err)
	}
	if mainSessionID != createdA.ID {
		t.Fatalf("expected first created session to be main, got %q", mainSessionID)
	}

	createReqB := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"title":"@agent","agent_id":"agent"}`))
	createReqB.Header.Set("Content-Type", "application/json")
	createResB := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createResB, createReqB)
	if createResB.Code != http.StatusCreated {
		t.Fatalf("unexpected second create status: %d body=%s", createResB.Code, createResB.Body.String())
	}
	var createdB struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createResB.Body.Bytes(), &createdB); err != nil {
		t.Fatalf("decode second create response: %v", err)
	}
	mainSessionID, err = s.ResolveMainSessionID(t.Context(), "agent", "gi", "default")
	if err != nil {
		t.Fatalf("resolve main session id after second create: %v", err)
	}
	if mainSessionID != createdB.ID {
		t.Fatalf("expected second created session to become main, got %q", mainSessionID)
	}
}

func TestForkSessionCreatesChildAgent(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "test-model", DefaultThinkingLevel: "medium"})

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"title":"@agent","agent_id":"agent"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRes.Code, createRes.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createRes.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if err := s.AddMessage(t.Context(), store.NowID("msg"), created.ID, "user", "hello", map[string]any{"kind": "chat"}); err != nil {
		t.Fatalf("seed source message: %v", err)
	}

	forkReq := httptest.NewRequest(http.MethodPost, "/api/sessions/"+created.ID+"/fork", bytes.NewBufferString(`{}`))
	forkReq.Header.Set("Content-Type", "application/json")
	forkRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(forkRes, forkReq)
	if forkRes.Code != http.StatusCreated {
		t.Fatalf("unexpected fork status: %d body=%s", forkRes.Code, forkRes.Body.String())
	}
	if !bytes.Contains(forkRes.Body.Bytes(), []byte(`"agent_id":"agent1"`)) {
		t.Fatalf("unexpected fork response: %s", forkRes.Body.String())
	}

	sessions, err := s.ListSessions(t.Context())
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	var child *store.Session
	for i := range sessions {
		if sessions[i].ParentSessionID == created.ID {
			child = &sessions[i]
			break
		}
	}
	if child == nil || child.Scope == nil || child.Scope.AgentID != "agent1" {
		t.Fatalf("unexpected child session: %#v", child)
	}
	msgs, err := s.ListMessages(t.Context(), child.ID)
	if err != nil {
		t.Fatalf("list child messages: %v", err)
	}
	if len(msgs) == 0 || msgs[0].Content != "hello" {
		t.Fatalf("unexpected child messages: %#v", msgs)
	}
}

func TestNextForkAgentIDPrefersCanonicalIdentityOverScopeSnapshot(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx := t.Context()
	rootAlloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_root_identity")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_root_identity", "", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}, &rootAlloc.Scope, rootAlloc.SessionAliases); err != nil {
		t.Fatalf("create root session: %v", err)
	}
	childAlloc := gisession.AllocateDefaultSession("agent1", "gi", "default", "session_child_identity")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_child_identity", "session_root_identity", "@agent1", map[string]any{"status": "idle", "model": "bootstrap"}, &childAlloc.Scope, childAlloc.SessionAliases); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update sessions set state_json = state_json where id = ?`, "session_root_identity"); err != nil {
		t.Fatalf("mutate root scope snapshot: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update sessions set state_json = state_json where id = ?`, "session_child_identity"); err != nil {
		t.Fatalf("mutate child scope snapshot: %v", err)
	}
	agentID, err := srv.nextForkAgentID(ctx, "session_root_identity")
	if err != nil {
		t.Fatalf("next fork agent id: %v", err)
	}
	if agentID != "agent2" {
		t.Fatalf("expected canonical fork agent id agent2, got %q", agentID)
	}
}

func TestNextForkAgentIDFallsBackWhenSourceIdentityMissingFromLookupMap(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx := t.Context()
	rootAlloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_root_identity_missing")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_root_identity_missing", "", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}, &rootAlloc.Scope, rootAlloc.SessionAliases); err != nil {
		t.Fatalf("create root session: %v", err)
	}
	childAlloc := gisession.AllocateDefaultSession("agent1", "gi", "default", "session_child_identity_present")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_child_identity_present", "session_root_identity_missing", "@agent1", map[string]any{"status": "idle", "model": "bootstrap"}, &childAlloc.Scope, childAlloc.SessionAliases); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `delete from session_identities where session_id = ?`, "session_root_identity_missing"); err != nil {
		t.Fatalf("delete root identity row: %v", err)
	}
	agentID, err := srv.nextForkAgentID(ctx, "session_root_identity_missing")
	if err != nil {
		t.Fatalf("next fork agent id: %v", err)
	}
	if agentID != "agent2" {
		t.Fatalf("expected fallback fork agent id agent2, got %q", agentID)
	}
}

func TestNextForkAgentIDDefaultsWhenUsedSessionIdentityMissingFromLookupMap(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx := t.Context()
	rootAlloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_root_identity_used_missing")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_root_identity_used_missing", "", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}, &rootAlloc.Scope, rootAlloc.SessionAliases); err != nil {
		t.Fatalf("create root session: %v", err)
	}
	childAlloc := gisession.AllocateDefaultSession("agent1", "gi", "default", "session_child_identity_used_missing")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_child_identity_used_missing", "session_root_identity_used_missing", "@agent1", map[string]any{"status": "idle", "model": "bootstrap"}, &childAlloc.Scope, childAlloc.SessionAliases); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `delete from session_identities where session_id = ?`, "session_child_identity_used_missing"); err != nil {
		t.Fatalf("delete child identity row: %v", err)
	}
	agentID, err := srv.nextForkAgentID(ctx, "session_root_identity_used_missing")
	if err != nil {
		t.Fatalf("next fork agent id: %v", err)
	}
	if agentID != "agent1" {
		t.Fatalf("expected defaulted used-session fork agent id agent1, got %q", agentID)
	}
}

func TestForkAgentIDForSessionResolutionFallbackOrder(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx := t.Context()
	alloc := gisession.AllocateDefaultSession("agent-base", "gi", "default", "session_fork_helper")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_fork_helper", "", "@agent-base", map[string]any{"status": "idle", "model": "bootstrap"}, &alloc.Scope, alloc.SessionAliases); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized(normalizeForkSessionID("session_fork_helper"), map[string]string{"session_fork_helper": " agent-map "}); got != "agent-map" || !mapped {
		t.Fatalf("expected mapped agent id to win, got %q mapped=%v", got, mapped)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized(normalizeForkSessionID("session_fork_helper"), map[string]string{}); got != defaultForkAgentID || mapped {
		t.Fatalf("expected default fallback agent id for non-source resolution, got %q mapped=%v", got, mapped)
	}
	if got := srv.sourceForkAgentID(ctx, "session_fork_helper", map[string]string{}); got != "agent-base" {
		t.Fatalf("expected source store fallback agent id, got %q", got)
	}
	if got := srv.sourceForkAgentID(ctx, "session_fork_helper", map[string]string{"session_fork_helper": defaultForkAgentID}); got != defaultForkAgentID {
		t.Fatalf("expected mapped default agent id to win over store fallback, got %q", got)
	}
	if got := srv.sourceForkAgentID(ctx, "session_fork_helper", nil); got != "agent-base" {
		t.Fatalf("expected nil map source fallback to canonical store agent id, got %q", got)
	}
	if got := srv.sourceForkAgentID(nil, "session_fork_helper", nil); got != "agent-base" {
		t.Fatalf("expected nil context + nil map source fallback to canonical store agent id, got %q", got)
	}
	if got := srv.sourceForkAgentID(ctx, "session_fork_helper", map[string]string{"session_fork_helper": "   "}); got != "agent-base" {
		t.Fatalf("expected whitespace-only mapped source agent id to fallback to store value, got %q", got)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized(normalizeForkSessionID("session_fork_helper_missing"), map[string]string{}); got != defaultForkAgentID || mapped {
		t.Fatalf("expected default fallback agent id, got %q mapped=%v", got, mapped)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized(normalizeForkSessionID("session_fork_helper"), map[string]string{}); got != defaultForkAgentID || mapped {
		t.Fatalf("expected nil-context default fallback agent id, got %q mapped=%v", got, mapped)
	}
	if got := srv.sourceForkAgentID(nil, "session_fork_helper", map[string]string{}); got != "agent-base" {
		t.Fatalf("expected nil-context source store fallback agent id, got %q", got)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized(normalizeForkSessionID("  session_fork_helper  "), map[string]string{}); got != defaultForkAgentID || mapped {
		t.Fatalf("expected trimmed session id default fallback agent id, got %q mapped=%v", got, mapped)
	}
	if got := srv.sourceForkAgentID(ctx, "  session_fork_helper  ", map[string]string{}); got != "agent-base" {
		t.Fatalf("expected trimmed source session id store fallback agent id, got %q", got)
	}
	if got := srv.sourceForkAgentID(ctx, "   ", map[string]string{}); got != defaultForkAgentID {
		t.Fatalf("expected empty source session id to remain default base agent id, got %q", got)
	}
	if got := srv.sourceForkAgentID(ctx, "   ", map[string]string{"": "agent-leak"}); got != defaultForkAgentID {
		t.Fatalf("expected empty source session id to ignore blank-key map values, got %q", got)
	}
	if got := srv.sourceForkAgentID(nil, "   ", map[string]string{"": "agent-leak"}); got != defaultForkAgentID {
		t.Fatalf("expected nil-context empty source session id to ignore blank-key map values, got %q", got)
	}
}

func TestTrimAgentNumericSuffix(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "agent42", want: "agent"},
		{in: " agent42 ", want: "agent"},
		{in: "agent", want: "agent"},
		{in: "123", want: "123"},
		{in: "agent007x", want: "agent007x"},
	}
	for _, tc := range cases {
		if got := trimAgentNumericSuffix(tc.in); got != tc.want {
			t.Fatalf("trim agent numeric suffix(%q): expected %q, got %q", tc.in, tc.want, got)
		}
	}
}

func TestNormalizeForkSessionID(t *testing.T) {
	if got := normalizeForkSessionID("  session-a  "); got != "session-a" {
		t.Fatalf("expected trimmed session id, got %q", got)
	}
}

func TestNormalizeForkAgentID(t *testing.T) {
	if got := normalizeForkAgentID("  agent-a  "); got != "agent-a" {
		t.Fatalf("expected trimmed agent id, got %q", got)
	}
}

func TestMapForkAgentIDOrDefaultNormalized(t *testing.T) {
	if got, mapped := mapForkAgentIDOrDefaultNormalized("s1", map[string]string{"s1": " agent-a "}); got != "agent-a" || !mapped {
		t.Fatalf("expected normalized mapped agent id to win, got %q mapped=%v", got, mapped)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized("s2", map[string]string{"s2": "   "}); got != defaultForkAgentID || mapped {
		t.Fatalf("expected whitespace-only normalized value to default, got %q mapped=%v", got, mapped)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized("s3", nil); got != defaultForkAgentID || mapped {
		t.Fatalf("expected nil map to default for normalized session id, got %q mapped=%v", got, mapped)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized("", map[string]string{"": "agent-blank"}); got != defaultForkAgentID || mapped {
		t.Fatalf("expected blank normalized session id to fail closed to default, got %q mapped=%v", got, mapped)
	}
}

func TestMapForkAgentIDOrDefaultNormalizedFromRawSessionID(t *testing.T) {
	if got, mapped := mapForkAgentIDOrDefaultNormalized(normalizeForkSessionID(" s1 "), map[string]string{"s1": " agent-a "}); got != "agent-a" || !mapped {
		t.Fatalf("expected trimmed mapped agent id, got %q mapped=%v", got, mapped)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized(normalizeForkSessionID("missing"), map[string]string{}); got != defaultForkAgentID || mapped {
		t.Fatalf("expected default for missing key, got %q mapped=%v", got, mapped)
	}
	if got, mapped := mapForkAgentIDOrDefaultNormalized(normalizeForkSessionID("s2"), map[string]string{"s2": "   "}); got != defaultForkAgentID || mapped {
		t.Fatalf("expected whitespace-only value to be treated as unmapped default, got %q mapped=%v", got, mapped)
	}
}

func TestBuildUsedForkAgentIDs(t *testing.T) {
	if used := buildUsedForkAgentIDs(nil); len(used) != 0 {
		t.Fatalf("expected nil input to yield empty used set, got %#v", used)
	}
	used := buildUsedForkAgentIDs(map[string]string{"s1": " agent-a ", "s2": "", "s3": "agent", "s4": "   ", "   ": "agent-z"})
	if !used["agent-a"] {
		t.Fatalf("expected trimmed mapped agent id to be marked used: %#v", used)
	}
	if !used[defaultForkAgentID] {
		t.Fatalf("expected default fork agent id to be marked used for blank/missing entries: %#v", used)
	}
	if len(used) != 2 {
		t.Fatalf("expected blank session id entries to be ignored, got %#v", used)
	}
}

func TestNormalizeForkAgentBase(t *testing.T) {
	if got := normalizeForkAgentBase("  "); got != defaultForkAgentID {
		t.Fatalf("expected empty base to normalize to default, got %q", got)
	}
	if got := normalizeForkAgentBase("  agentX  "); got != "agentX" {
		t.Fatalf("expected trimmed non-empty base, got %q", got)
	}
}

func TestForkAgentIDCandidate(t *testing.T) {
	if got := forkAgentIDCandidate("agent", 3); got != "agent3" {
		t.Fatalf("expected candidate agent3, got %q", got)
	}
}

func TestChooseNextForkAgentID(t *testing.T) {
	used := map[string]bool{"agent1": true, "agent2": true}
	if got, ok := chooseNextForkAgentID("agent", used); !ok || got != "agent3" {
		t.Fatalf("expected agent3 with available suffix, got %q ok=%v", got, ok)
	}
	if got, ok := chooseNextForkAgentID("  ", map[string]bool{}); !ok || got != "agent1" {
		t.Fatalf("expected default-base candidate agent1 for empty base, got %q ok=%v", got, ok)
	}
	if got, ok := chooseNextForkAgentID("  ", nil); !ok || got != "agent1" {
		t.Fatalf("expected default-base candidate agent1 for nil used set, got %q ok=%v", got, ok)
	}
	if got, ok := chooseNextForkAgentID("agent", map[string]bool{}); !ok || got != "agent1" {
		t.Fatalf("expected first fork candidate agent1 for empty used set, got %q ok=%v", got, ok)
	}

	exhausted := map[string]bool{}
	for i := minForkAgentIDSuffix; i < maxForkAgentIDSuffixExclusive; i++ {
		exhausted["agent"+strconv.Itoa(i)] = true
	}
	if got, ok := chooseNextForkAgentID("agent", exhausted); ok || got != "" {
		t.Fatalf("expected exhaustion with empty candidate, got %q ok=%v", got, ok)
	}
}

func TestNextForkAgentIDNilContext(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx := t.Context()
	rootAlloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_root_nil_ctx")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_root_nil_ctx", "", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}, &rootAlloc.Scope, rootAlloc.SessionAliases); err != nil {
		t.Fatalf("create root session: %v", err)
	}
	if got, err := srv.nextForkAgentID(nil, "session_root_nil_ctx"); err != nil {
		t.Fatalf("next fork agent id (nil ctx): %v", err)
	} else if got != "agent1" {
		t.Fatalf("expected nil-context next fork agent id agent1, got %q", got)
	}
}

func TestNextForkAgentIDTrimsSourceSessionID(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx := t.Context()
	rootAlloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_root_trimmed_source")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_root_trimmed_source", "", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}, &rootAlloc.Scope, rootAlloc.SessionAliases); err != nil {
		t.Fatalf("create root session: %v", err)
	}
	if got, err := srv.nextForkAgentID(ctx, "  session_root_trimmed_source  "); err != nil {
		t.Fatalf("next fork agent id (trimmed source): %v", err)
	} else if got != "agent1" {
		t.Fatalf("expected trimmed-source next fork agent id agent1, got %q", got)
	}
}

func TestNextForkAgentIDEmptySourceSessionIDUsesDefaultBase(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx := t.Context()
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_existing_default_base")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_existing_default_base", "", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}, &alloc.Scope, alloc.SessionAliases); err != nil {
		t.Fatalf("create existing session: %v", err)
	}
	if got, err := srv.nextForkAgentID(ctx, "   "); err != nil {
		t.Fatalf("next fork agent id (empty source): %v", err)
	} else if got != "agent1" {
		t.Fatalf("expected default-base next fork agent id agent1, got %q", got)
	}
}

func TestSessionContinueEndpointStartsQueuedSteering(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	srv := New(s, engine, config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})

	session, err := s.CreateSession(t.Context(), store.NowID("session"), "Demo", map[string]any{"status": "idle", "model": "bootstrap", "provider": "test", "thinking_level": "medium"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.EnqueueSteering(t.Context(), session.ID, "", "user", "continue me", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}

	continueReq := httptest.NewRequest(http.MethodPost, "/api/sessions/"+session.ID+"/continue", nil)
	continueRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(continueRes, continueReq)
	if continueRes.Code != http.StatusOK {
		t.Fatalf("unexpected continue status: %d body=%s", continueRes.Code, continueRes.Body.String())
	}
	if !bytes.Contains(continueRes.Body.Bytes(), []byte(`"continued":true`)) {
		t.Fatalf("unexpected continue response: %s", continueRes.Body.String())
	}

	time.Sleep(1500 * time.Millisecond)
	msgs, err := s.ListMessages(t.Context(), session.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if !bytes.Contains([]byte(fmt.Sprintf("%v", msgs)), []byte("continue me")) {
		t.Fatalf("expected continued steering message in history, got %#v", msgs)
	}
}

func TestDirectedPromptCreatesTargetAgentSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"title":"@agent","agent_id":"agent"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createRes, createReq)
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(createRes.Body.Bytes(), &created)

	promptReq := httptest.NewRequest(http.MethodPost, "/api/sessions/"+created.ID+"/prompt", bytes.NewBufferString(`{"prompt":"@agent1 hello directed","model":"bootstrap"}`))
	promptReq.Header.Set("Content-Type", "application/json")
	promptRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(promptRes, promptReq)
	if promptRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected directed prompt status: %d body=%s", promptRes.Code, promptRes.Body.String())
	}
	if !bytes.Contains(promptRes.Body.Bytes(), []byte(`"target_agent_id":"agent1"`)) {
		t.Fatalf("unexpected directed prompt response: %s", promptRes.Body.String())
	}
}

func TestSessionRouteEventsEndpoint(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"title":"@agent","agent_id":"agent"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create session status: %d body=%s", createRes.Code, createRes.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(createRes.Body.Bytes(), &created)

	promptReq := httptest.NewRequest(http.MethodPost, "/api/sessions/"+created.ID+"/prompt", bytes.NewBufferString(`{"prompt":"@agent1 hello routed","model":"bootstrap"}`))
	promptReq.Header.Set("Content-Type", "application/json")
	promptRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(promptRes, promptReq)
	if promptRes.Code != http.StatusAccepted {
		t.Fatalf("directed prompt status: %d body=%s", promptRes.Code, promptRes.Body.String())
	}
	eventsReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+created.ID+"/route-events", nil)
	eventsRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(eventsRes, eventsReq)
	if eventsRes.Code != http.StatusOK {
		t.Fatalf("route-events status: %d body=%s", eventsRes.Code, eventsRes.Body.String())
	}
	var payload struct {
		RouteEvents []routing.Event `json:"route_events"`
	}
	if err := json.Unmarshal(eventsRes.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode route-events: %v", err)
	}
	if len(payload.RouteEvents) == 0 {
		t.Fatalf("expected at least one route event, got none")
	}
	if payload.RouteEvents[0].TargetAgentID != "agent1" {
		t.Fatalf("unexpected target agent: %#v", payload.RouteEvents[0])
	}
}

func TestPeerMessageEndpointRoutesToTargetAgent(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})

	source, err := s.CreateSession(t.Context(), "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	_, err = s.CloneSession(t.Context(), source.ID, "session_child", "@agent1", "agent1")
	if err != nil {
		t.Fatalf("create target: %v", err)
	}

	peerReq := httptest.NewRequest(http.MethodPost, "/api/sessions/"+source.ID+"/peer-message", bytes.NewBufferString(`{"target_agent_id":"agent1","content":"hello peer","mode":"prompt","model":"bootstrap"}`))
	peerReq.Header.Set("Content-Type", "application/json")
	peerRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(peerRes, peerReq)
	if peerRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected peer status: %d body=%s", peerRes.Code, peerRes.Body.String())
	}
	if !bytes.Contains(peerRes.Body.Bytes(), []byte(`"target_agent_id":"agent1"`)) {
		t.Fatalf("unexpected peer response: %s", peerRes.Body.String())
	}
}

func TestRuntimeInboundWorkPublishesTopicLifecycleEvents(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	srv := New(s, engine, config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.inbound_work", topics.SubscribeOptions{Buffer: 16})
	defer unsub()
	enqueueReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work", bytes.NewBufferString(`{"kind":"prompt","prompt":"topic item"}`))
	enqueueReq.Header.Set("Content-Type", "application/json")
	enqueueRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(enqueueRes, enqueueReq)
	if enqueueRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected enqueue status: %d body=%s", enqueueRes.Code, enqueueRes.Body.String())
	}
	var enqueuePayload struct {
		Item queue.InboundWorkItem `json:"item"`
	}
	if err := json.Unmarshal(enqueueRes.Body.Bytes(), &enqueuePayload); err != nil {
		t.Fatalf("decode enqueue response: %v", err)
	}
	requeueReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work/requeue", bytes.NewBufferString(fmt.Sprintf(`{"id":%d}`, enqueuePayload.Item.ID)))
	requeueReq.Header.Set("Content-Type", "application/json")
	requeueRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(requeueRes, requeueReq)
	if requeueRes.Code != http.StatusBadRequest {
		t.Fatalf("expected queued item requeue rejection, got %d body=%s", requeueRes.Code, requeueRes.Body.String())
	}
	if err := queue.RecordInboundWorkFailure(t.Context(), s.DB(), enqueuePayload.Item.ID, 1, "bad"); err != nil {
		t.Fatalf("mark inbound work failed: %v", err)
	}
	requeueReq = httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work/requeue", bytes.NewBufferString(fmt.Sprintf(`{"id":%d}`, enqueuePayload.Item.ID)))
	requeueReq.Header.Set("Content-Type", "application/json")
	requeueRes = httptest.NewRecorder()
	srv.Handler().ServeHTTP(requeueRes, requeueReq)
	if requeueRes.Code != http.StatusOK {
		t.Fatalf("unexpected requeue status: %d body=%s", requeueRes.Code, requeueRes.Body.String())
	}
	discardReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work/discard", bytes.NewBufferString(fmt.Sprintf(`{"id":%d}`, enqueuePayload.Item.ID)))
	discardReq.Header.Set("Content-Type", "application/json")
	discardRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(discardRes, discardReq)
	if discardRes.Code != http.StatusOK {
		t.Fatalf("unexpected discard status: %d body=%s", discardRes.Code, discardRes.Body.String())
	}
	want := map[string]bool{"inbound_work_enqueued": false, "inbound_work_requeued": false, "inbound_work_discarded": false}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case env := <-ch:
			if env.Topic != "runtime.inbound_work" {
				t.Fatalf("unexpected inbound work topic envelope: %#v", env)
			}
			if typ, _ := env.Payload["type"].(string); typ != "" {
				if _, ok := want[typ]; ok {
					want[typ] = true
				}
			}
		case <-time.After(50 * time.Millisecond):
		}
		all := true
		for _, seen := range want {
			all = all && seen
		}
		if all {
			return
		}
	}
	t.Fatalf("missing inbound work lifecycle topic events: %#v", want)
}

func TestRuntimeDispatcherPublishesLeaseTopicEvents(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	srv := New(s, engine, config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium", InboundWork: config.InboundWorkSettings{Enabled: true, IntervalMS: 25, BatchSize: 1, WorkerID: "web-test-dispatcher", LeaseTTLMS: 500}})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.dispatcher", topics.SubscribeOptions{Buffer: 32})
	defer unsub()
	srv.StartInboundWorkDispatcher(ctx)
	seenAcquire := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case env := <-ch:
			if env.Topic != "runtime.dispatcher" {
				t.Fatalf("unexpected runtime.dispatcher topic envelope: %#v", env)
			}
			if env.Payload["type"] == "dispatcher_lease_acquired" {
				seenAcquire = true
				return
			}
		case <-time.After(50 * time.Millisecond):
		}
	}
	if !seenAcquire {
		t.Fatal("expected dispatcher lease topic event")
	}
}

func TestTopicSSEStreamsRuntimeTopicEvents(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	srv := New(s, engine, config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/sse/topics?topic=runtime.inbound_work&session_id=session_topic_sse", nil).WithContext(ctx)
	res := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.Handler().ServeHTTP(res, req)
	}()
	time.Sleep(25 * time.Millisecond)
	engine.PublishRuntimeInboundWorkEvent("inbound_work_test", &queue.InboundWorkItem{ID: 1, Status: "queued", SessionID: "session_topic_sse", SourceKind: "ipc"}, nil)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		body := res.Body.String()
		if strings.Contains(body, "event: connected") && strings.Contains(body, `"last_sequence"`) && strings.Contains(body, "event: runtime.inbound_work") && strings.Contains(body, "inbound_work_test") && strings.Contains(body, "id: ") && strings.Contains(body, `"sequence":`) {
			cancel()
			<-done
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	cancel()
	<-done
	t.Fatalf("expected topic SSE stream body to contain runtime topic event, got %s", res.Body.String())
}

func TestTopicSSERejectsInvalidBuffer(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	for _, raw := range []string{"nope", "0", "2048"} {
		req := httptest.NewRequest(http.MethodGet, "/sse/topics?buffer="+raw, nil)
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request for invalid topic SSE buffer %q, got %d body=%s", raw, res.Code, res.Body.String())
		}
	}
}

func TestRuntimeInboundWorkEndpoints(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	srv := New(s, engine, config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	session, err := s.CreateSession(t.Context(), store.NowID("session"), "Demo", map[string]any{"status": "idle", "model": "bootstrap", "provider": "test", "thinking_level": "medium"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	enqueueReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work", bytes.NewBufferString(fmt.Sprintf(`{"kind":"prompt","session_id":%q,"prompt":"hello from runtime queue"}`, session.ID)))
	enqueueReq.Header.Set("Content-Type", "application/json")
	enqueueRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(enqueueRes, enqueueReq)
	if enqueueRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected enqueue status: %d body=%s", enqueueRes.Code, enqueueRes.Body.String())
	}
	if !bytes.Contains(enqueueRes.Body.Bytes(), []byte(`"status":"queued"`)) {
		t.Fatalf("unexpected enqueue response: %s", enqueueRes.Body.String())
	}
	listReq := httptest.NewRequest(http.MethodGet, "/api/runtime/inbound-work?status=queued&limit=10", nil)
	listRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("unexpected inbound-work list status: %d body=%s", listRes.Code, listRes.Body.String())
	}
	if !bytes.Contains(listRes.Body.Bytes(), []byte("hello from runtime queue")) {
		t.Fatalf("expected queued inbound work in list response, got %s", listRes.Body.String())
	}
	drainReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work/drain", bytes.NewBufferString(`{"claimed_by":"web-test","limit":10}`))
	drainReq.Header.Set("Content-Type", "application/json")
	drainRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(drainRes, drainReq)
	if drainRes.Code != http.StatusOK {
		t.Fatalf("unexpected drain status: %d body=%s", drainRes.Code, drainRes.Body.String())
	}
	if !bytes.Contains(drainRes.Body.Bytes(), []byte(`"processed":1`)) || !bytes.Contains(drainRes.Body.Bytes(), []byte(`"status":"completed"`)) {
		t.Fatalf("unexpected drain response: %s", drainRes.Body.String())
	}
	msgs, err := s.ListMessages(t.Context(), session.ID)
	if err != nil {
		t.Fatalf("list session messages after drain: %v", err)
	}
	if !bytes.Contains([]byte(fmt.Sprintf("%v", msgs)), []byte("hello from runtime queue")) {
		t.Fatalf("expected drained inbound prompt in session history, got %#v", msgs)
	}
}

func TestRuntimeInboundWorkDispatcherProcessesQueuedItems(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	srv := New(s, engine, config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium", InboundWork: config.InboundWorkSettings{Enabled: true, IntervalMS: 25, BatchSize: 4, WorkerID: "web-test-dispatcher"}})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	srv.StartInboundWorkDispatcher(ctx)
	session, err := s.CreateSession(t.Context(), store.NowID("session"), "Demo", map[string]any{"status": "idle", "model": "bootstrap", "provider": "test", "thinking_level": "medium"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	enqueueReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work", bytes.NewBufferString(fmt.Sprintf(`{"kind":"prompt","session_id":%q,"prompt":"hello from dispatcher"}`, session.ID)))
	enqueueReq.Header.Set("Content-Type", "application/json")
	enqueueRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(enqueueRes, enqueueReq)
	if enqueueRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected enqueue status: %d body=%s", enqueueRes.Code, enqueueRes.Body.String())
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		msgs, err := s.ListMessages(t.Context(), session.ID)
		if err == nil && bytes.Contains([]byte(fmt.Sprintf("%v", msgs)), []byte("hello from dispatcher")) {
			break
		}
		if time.Now().After(deadline) {
			items, listErr := queue.ListInboundWork(t.Context(), s.DB(), "", 10)
			if listErr != nil {
				t.Fatalf("timed out waiting for dispatcher; list inbound work: %v", listErr)
			}
			t.Fatalf("timed out waiting for dispatcher to process queued item; inbound=%#v msgsErr=%v", items, err)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func TestRuntimeInboundWorkDispatcherContinuesPastRetryingItem(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	srv := New(s, engine, config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium", InboundWork: config.InboundWorkSettings{Enabled: true, IntervalMS: 25, BatchSize: 4, WorkerID: "web-test-dispatcher"}})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	srv.StartInboundWorkDispatcher(ctx)
	session, err := s.CreateSession(t.Context(), store.NowID("session"), "Demo", map[string]any{"status": "idle", "model": "bootstrap", "provider": "test", "thinking_level": "medium"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	badReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work", bytes.NewBufferString(`{"kind":"prompt","prompt":"missing session"}`))
	badReq.Header.Set("Content-Type", "application/json")
	badRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(badRes, badReq)
	if badRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected bad enqueue status: %d body=%s", badRes.Code, badRes.Body.String())
	}
	goodReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work", bytes.NewBufferString(fmt.Sprintf(`{"kind":"prompt","session_id":%q,"prompt":"hello after retry item"}`, session.ID)))
	goodReq.Header.Set("Content-Type", "application/json")
	goodRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(goodRes, goodReq)
	if goodRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected good enqueue status: %d body=%s", goodRes.Code, goodRes.Body.String())
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		msgs, err := s.ListMessages(t.Context(), session.ID)
		if err == nil && bytes.Contains([]byte(fmt.Sprintf("%v", msgs)), []byte("hello after retry item")) {
			break
		}
		if time.Now().After(deadline) {
			items, listErr := queue.ListInboundWork(t.Context(), s.DB(), "", 10)
			if listErr != nil {
				t.Fatalf("timed out waiting for dispatcher after retry item; list inbound work: %v", listErr)
			}
			t.Fatalf("timed out waiting for dispatcher to continue past retrying item; inbound=%#v msgsErr=%v", items, err)
		}
		time.Sleep(25 * time.Millisecond)
	}
	items, err := queue.ListInboundWork(t.Context(), s.DB(), "", 10)
	if err != nil {
		t.Fatalf("list inbound work: %v", err)
	}
	if len(items) < 2 || items[0].Status != "retry" {
		t.Fatalf("expected first inbound item to be retrying while later item completed, got %#v", items)
	}
}

func TestRuntimeInboundWorkEligibleFilterAndCounts(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	if _, err := queue.EnqueueInboundWork(t.Context(), s.DB(), "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "ready now"}); err != nil {
		t.Fatalf("enqueue queued item: %v", err)
	}
	retrying, err := queue.EnqueueInboundWork(t.Context(), s.DB(), "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "retry later"})
	if err != nil {
		t.Fatalf("enqueue retry item: %v", err)
	}
	if err := queue.RecordInboundWorkRetry(t.Context(), s.DB(), retrying.ID, 1, "temporary failure", time.Minute); err != nil {
		t.Fatalf("mark retry item: %v", err)
	}
	failed, err := queue.EnqueueInboundWork(t.Context(), s.DB(), "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "failed"})
	if err != nil {
		t.Fatalf("enqueue failed item: %v", err)
	}
	if err := queue.RecordInboundWorkFailure(t.Context(), s.DB(), failed.ID, 3, "permanent failure"); err != nil {
		t.Fatalf("mark failed item: %v", err)
	}
	readyReq := httptest.NewRequest(http.MethodGet, "/api/runtime/inbound-work?eligible=true&limit=10", nil)
	readyRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(readyRes, readyReq)
	if readyRes.Code != http.StatusOK {
		t.Fatalf("unexpected eligible list status: %d body=%s", readyRes.Code, readyRes.Body.String())
	}
	if !bytes.Contains(readyRes.Body.Bytes(), []byte("ready now")) || bytes.Contains(readyRes.Body.Bytes(), []byte("retry later")) {
		t.Fatalf("unexpected eligible list response: %s", readyRes.Body.String())
	}
	if !bytes.Contains(readyRes.Body.Bytes(), []byte(`"eligible_count":1`)) || !bytes.Contains(readyRes.Body.Bytes(), []byte(`"queued":1`)) || !bytes.Contains(readyRes.Body.Bytes(), []byte(`"retry":1`)) || !bytes.Contains(readyRes.Body.Bytes(), []byte(`"failed":1`)) {
		t.Fatalf("expected eligible/count metadata in response, got %s", readyRes.Body.String())
	}
	blockedReq := httptest.NewRequest(http.MethodGet, "/api/runtime/inbound-work?eligible=false&limit=10", nil)
	blockedRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(blockedRes, blockedReq)
	if blockedRes.Code != http.StatusOK {
		t.Fatalf("unexpected ineligible list status: %d body=%s", blockedRes.Code, blockedRes.Body.String())
	}
	if !bytes.Contains(blockedRes.Body.Bytes(), []byte("retry later")) || !bytes.Contains(blockedRes.Body.Bytes(), []byte("failed")) || bytes.Contains(blockedRes.Body.Bytes(), []byte("ready now")) {
		t.Fatalf("unexpected ineligible list response: %s", blockedRes.Body.String())
	}
}

func TestRuntimeInboundWorkRequeueEndpoint(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	item, err := queue.EnqueueInboundWork(t.Context(), s.DB(), "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "manual requeue"})
	if err != nil {
		t.Fatalf("enqueue inbound work: %v", err)
	}
	if err := queue.RecordInboundWorkFailure(t.Context(), s.DB(), item.ID, 3, "boom"); err != nil {
		t.Fatalf("record failure: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work/requeue", bytes.NewBufferString(fmt.Sprintf(`{"id":%d,"reset_attempts":true}`, item.ID)))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("unexpected requeue status: %d body=%s", res.Code, res.Body.String())
	}
	if !bytes.Contains(res.Body.Bytes(), []byte(`"status":"queued"`)) || !bytes.Contains(res.Body.Bytes(), []byte(`"attempt_count":0`)) {
		t.Fatalf("unexpected requeue response: %s", res.Body.String())
	}
}

func TestRuntimeInboundWorkRequeueRejectsQueuedItem(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	item, err := queue.EnqueueInboundWork(t.Context(), s.DB(), "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "still queued"})
	if err != nil {
		t.Fatalf("enqueue inbound work: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work/requeue", bytes.NewBufferString(fmt.Sprintf(`{"id":%d}`, item.ID)))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request requeueing queued item, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestRuntimeInboundWorkDiscardEndpoint(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	item, err := queue.EnqueueInboundWork(t.Context(), s.DB(), "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "discard me"})
	if err != nil {
		t.Fatalf("enqueue inbound work: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work/discard", bytes.NewBufferString(fmt.Sprintf(`{"id":%d}`, item.ID)))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("unexpected discard status: %d body=%s", res.Code, res.Body.String())
	}
	if !bytes.Contains(res.Body.Bytes(), []byte(`"status":"discarded"`)) {
		t.Fatalf("unexpected discard response: %s", res.Body.String())
	}
}

func TestRuntimeInboundWorkDiscardRejectsCompletedItem(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	item, err := queue.EnqueueInboundWork(t.Context(), s.DB(), "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "complete then discard"})
	if err != nil {
		t.Fatalf("enqueue inbound work: %v", err)
	}
	if err := queue.UpdateInboundWorkStatus(t.Context(), s.DB(), item.ID, "completed"); err != nil {
		t.Fatalf("mark inbound work completed: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work/discard", bytes.NewBufferString(fmt.Sprintf(`{"id":%d}`, item.ID)))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request discarding completed item, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestRuntimeInboundWorkDispatcherStartsOnlyOncePerServer(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	cfg := config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium", InboundWork: config.InboundWorkSettings{Enabled: true, IntervalMS: 1000, BatchSize: 1, WorkerID: "web-test-dispatcher-once", LeaseTTLMS: 500}}
	srv := New(s, engine, cfg)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.dispatcher", topics.SubscribeOptions{Buffer: 16})
	defer unsub()
	srv.StartInboundWorkDispatcher(ctx)
	srv.StartInboundWorkDispatcher(ctx)
	acquires := 0
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		select {
		case env := <-ch:
			if env.Payload["type"] == "dispatcher_lease_acquired" {
				acquires++
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	if acquires != 1 {
		t.Fatalf("expected a single dispatcher startup acquisition, got %d", acquires)
	}
}

func TestRuntimeInboundWorkDispatcherAcceptsNilContext(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	cfg := config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium", InboundWork: config.InboundWorkSettings{Enabled: true, IntervalMS: 25, BatchSize: 1, WorkerID: "web-test-dispatcher-nil", LeaseTTLMS: 500}}
	srv := New(s, engine, cfg)
	defer engine.Close()
	srv.StartInboundWorkDispatcher(nil)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var owner string
		row := s.DB().QueryRowContext(context.Background(), `select value from kv_store where namespace = ? and key = ?`, "runtime_leases", "inbound_dispatcher")
		if err := row.Scan(&owner); err == nil && owner != "" {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("expected dispatcher lease acquisition with nil context")
}

func TestRuntimeInboundWorkDispatcherReleasesLeaseAfterContextCancel(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	cfg := config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium", InboundWork: config.InboundWorkSettings{Enabled: true, IntervalMS: 25, BatchSize: 1, WorkerID: "web-test-dispatcher-release", LeaseTTLMS: 500}}
	srv := New(s, engine, cfg)
	ctx, cancel := context.WithCancel(t.Context())
	srv.StartInboundWorkDispatcher(ctx)
	var owner string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		row := s.DB().QueryRowContext(context.Background(), `select value from kv_store where namespace = ? and key = ?`, "runtime_leases", "inbound_dispatcher")
		owner = ""
		if err := row.Scan(&owner); err == nil && owner != "" {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if owner == "" {
		t.Fatal("expected dispatcher lease owner before cancel")
	}
	cancel()
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		row := s.DB().QueryRowContext(context.Background(), `select value from kv_store where namespace = ? and key = ?`, "runtime_leases", "inbound_dispatcher")
		owner = ""
		if err := row.Scan(&owner); err != nil || owner == "" {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("expected dispatcher lease release after cancel")
}

func TestRuntimeInboundWorkDispatcherUsesSingleLeaseHolder(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engineA := turn.New(s)
	engineB := turn.New(s)
	cfg := config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium", InboundWork: config.InboundWorkSettings{Enabled: true, IntervalMS: 25, BatchSize: 1, WorkerID: "web-test-dispatcher", LeaseTTLMS: 500}}
	srvA := New(s, engineA, cfg)
	srvB := New(s, engineB, cfg)
	ctxA, cancelA := context.WithCancel(t.Context())
	defer cancelA()
	ctxB, cancelB := context.WithCancel(t.Context())
	defer cancelB()
	srvA.StartInboundWorkDispatcher(ctxA)
	srvB.StartInboundWorkDispatcher(ctxB)
	session, err := s.CreateSession(t.Context(), store.NowID("session"), "Demo", map[string]any{"status": "idle", "model": "bootstrap", "provider": "test", "thinking_level": "medium"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	enqueueReq := httptest.NewRequest(http.MethodPost, "/api/runtime/inbound-work", bytes.NewBufferString(fmt.Sprintf(`{"kind":"prompt","session_id":%q,"prompt":"lease holder only"}`, session.ID)))
	enqueueReq.Header.Set("Content-Type", "application/json")
	enqueueRes := httptest.NewRecorder()
	srvA.Handler().ServeHTTP(enqueueRes, enqueueReq)
	if enqueueRes.Code != http.StatusAccepted {
		t.Fatalf("unexpected enqueue status: %d body=%s", enqueueRes.Code, enqueueRes.Body.String())
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		msgs, err := s.ListMessages(t.Context(), session.ID)
		if err == nil && bytes.Contains([]byte(fmt.Sprintf("%v", msgs)), []byte("lease holder only")) {
			break
		}
		if time.Now().After(deadline) {
			items, listErr := queue.ListInboundWork(t.Context(), s.DB(), "", 10)
			if listErr != nil {
				t.Fatalf("timed out waiting for leased dispatcher; list inbound work: %v", listErr)
			}
			t.Fatalf("timed out waiting for leased dispatcher to process queued item; inbound=%#v msgsErr=%v", items, err)
		}
		time.Sleep(25 * time.Millisecond)
	}
	items, err := queue.ListInboundWork(t.Context(), s.DB(), "completed", 10)
	if err != nil {
		t.Fatalf("list completed inbound work: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected exactly one completed inbound item, got %#v", items)
	}
}

func TestRuntimeInboundWorkRejectsInvalidEligibleFlag(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	req := httptest.NewRequest(http.MethodGet, "/api/runtime/inbound-work?eligible=maybe", nil)
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid eligible flag, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestRuntimeInboundWorkRejectsInvalidListLimit(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "test", DefaultModel: "bootstrap", DefaultThinkingLevel: "medium"})
	req := httptest.NewRequest(http.MethodGet, "/api/runtime/inbound-work?limit=nope", nil)
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid limit, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestWorkspaceEndpoints(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.md"), []byte("# hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})

	treeReq := httptest.NewRequest(http.MethodGet, "/api/workspace/tree", nil)
	treeRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(treeRes, treeReq)
	if treeRes.Code != http.StatusOK || !bytes.Contains(treeRes.Body.Bytes(), []byte("hello.md")) {
		t.Fatalf("unexpected tree response: %d %s", treeRes.Code, treeRes.Body.String())
	}

	fileReq := httptest.NewRequest(http.MethodGet, "/api/workspace/file?path=hello.md", nil)
	fileRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(fileRes, fileReq)
	if fileRes.Code != http.StatusOK || !bytes.Contains(fileRes.Body.Bytes(), []byte("# hi")) {
		t.Fatalf("unexpected file response: %d %s", fileRes.Code, fileRes.Body.String())
	}
}

func TestWorkspaceFileRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink not supported in test env: %v", err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})

	fileReq := httptest.NewRequest(http.MethodGet, "/api/workspace/file?path=link/secret.txt", nil)
	fileRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(fileRes, fileReq)
	if fileRes.Code != http.StatusBadRequest {
		t.Fatalf("expected symlink escape rejection, got %d body=%s", fileRes.Code, fileRes.Body.String())
	}
}

func TestWorkspaceEndpointsRequireAuthWhenEnrolled(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".gi"), 0o755); err != nil {
		t.Fatalf("create .gi dir: %v", err)
	}
	authState := `{"username":"admin","totp_secret":"JBSWY3DPEHPK3PXP","totp_enabled":true,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","sessions":[]}`
	if err := os.WriteFile(filepath.Join(root, ".gi", "auth.json"), []byte(authState), 0o600); err != nil {
		t.Fatalf("write auth state: %v", err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})

	treeReq := httptest.NewRequest(http.MethodGet, "/api/workspace/tree", nil)
	treeRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(treeRes, treeReq)
	if treeRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized workspace tree without bearer token, got %d body=%s", treeRes.Code, treeRes.Body.String())
	}
}
