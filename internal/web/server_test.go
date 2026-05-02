package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
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

	time.Sleep(1500 * time.Millisecond)
	turnsReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+created.ID+"/turns", nil)
	turnsRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(turnsRes, turnsReq)
	if turnsRes.Code != http.StatusOK || !bytes.Contains(turnsRes.Body.Bytes(), []byte("completed")) {
		t.Fatalf("unexpected turns status/body: %d %s", turnsRes.Code, turnsRes.Body.String())
	}

	messagesReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+created.ID+"/messages", nil)
	messagesRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(messagesRes, messagesReq)
	if messagesRes.Code != http.StatusOK {
		t.Fatalf("unexpected messages status: %d body=%s", messagesRes.Code, messagesRes.Body.String())
	}
	if !bytes.Contains(messagesRes.Body.Bytes(), []byte("Gi received: hello")) || !bytes.Contains(messagesRes.Body.Bytes(), []byte(`"agent_id":"agent"`)) {
		t.Fatalf("unexpected messages body: %s", messagesRes.Body.String())
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
