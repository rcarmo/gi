package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
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
	mainSess, err := s.ResolveMainSession(t.Context(), "agent", "gi", "default")
	if err != nil {
		t.Fatalf("resolve main session after first create: %v", err)
	}
	if mainSess.ID != createdA.ID {
		t.Fatalf("expected first created session to be main, got %#v", mainSess)
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
	mainSess, err = s.ResolveMainSession(t.Context(), "agent", "gi", "default")
	if err != nil {
		t.Fatalf("resolve main session after second create: %v", err)
	}
	if mainSess.ID != createdB.ID {
		t.Fatalf("expected second created session to become main, got %#v", mainSess)
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
	if _, err := s.DB().ExecContext(ctx, `update sessions set scope_json = ? where id = ?`, `{"version":1,"agent_id":"wrong99","channel":"gi","account":"default","dimensions":["chat"],"values":{"chat":"direct:session_root_identity"}}`, "session_root_identity"); err != nil {
		t.Fatalf("mutate root scope snapshot: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update sessions set scope_json = ? where id = ?`, `{"version":1,"agent_id":"wrong100","channel":"gi","account":"default","dimensions":["chat"],"values":{"chat":"direct:session_child_identity"}}`, "session_child_identity"); err != nil {
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
		RouteEvents []store.RouteEvent `json:"route_events"`
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
