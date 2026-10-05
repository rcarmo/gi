package web

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	goai "github.com/rcarmo/go-ai"
)

func sidePromptFixture(t *testing.T, onCancel ...func()) *Server {
	t.Helper()
	auth := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", auth)
	t.Setenv("PI_CODING_AGENT_DIR", auth)
	if err := os.WriteFile(filepath.Join(auth, "auth.json"), []byte(`{"web-side-fixture":{"type":"api_key","key":"fixture-only"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Role    string `json:"role"`
				Content any    `json:"content"`
			} `json:"messages"`
			Tools []any `json:"tools"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(body.Messages)
		if !bytes.Contains(raw, []byte("BTW answer:")) && !bytes.Contains(raw, []byte("main-held")) && len(body.Tools) > 0 {
			t.Error("side request declared tools")
		}
		if bytes.Contains(raw, []byte("other-session-secret")) {
			t.Error("cross session context")
		}
		if len(body.Tools) > 0 && bytes.Contains(raw, []byte("main-held")) {
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data: {\"id\":\"fixture\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"main waiting\"},\"finish_reason\":null}]}\n\n")
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			return
		}
		if bytes.Contains(raw, []byte("wait-side")) {
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data: {\"id\":\"fixture\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"waiting\"},\"finish_reason\":null}]}\n\n")
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			for _, callback := range onCancel {
				callback()
			}
			return
		}
		if bytes.Contains(raw, []byte("fail-side")) {
			http.Error(w, "fixture failure", 500)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		for _, delta := range []map[string]any{{"role": "assistant", "reasoning_content": "side thought"}, {"content": "side answer"}} {
			payload, _ := json.Marshal(map[string]any{"id": "fixture", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": delta, "finish_reason": nil}}})
			fmt.Fprintf(w, "data: %s\n\n", payload)
			w.(http.Flusher).Flush()
		}
		fmt.Fprint(w, `data: {"id":"fixture","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`+"\n\ndata: [DONE]\n\n")
	}))
	t.Cleanup(provider.Close)
	inference.Init()
	goai.RegisterModel(&goai.Model{ID: t.Name() + "-" + strings.TrimPrefix(provider.URL, "http://127.0.0.1:"), Provider: "web-side-fixture", Name: "Side fixture", Api: goai.ApiOpenAICompletions, BaseURL: provider.URL + "/v1", ContextWindow: 32000, MaxTokens: 1024})
	db, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.CreateSession(context.Background(), "s", "s", nil)
	db.CreateSession(context.Background(), "other", "other", nil)
	db.AddMessage(context.Background(), "history", "s", "user", "main history", nil)
	db.AddMessage(context.Background(), "other-history", "other", "user", "other-session-secret", nil)
	cfg := config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultProvider: "web-side-fixture", DefaultModel: t.Name() + "-" + strings.TrimPrefix(provider.URL, "http://127.0.0.1:"), DefaultThinkingLevel: "off"}
	engine := turn.NewWithRuntimeConfig(db, cfg, "")
	t.Cleanup(func() { engine.Close() })
	return New(db, engine, cfg)
}
func sideRequest(s *Server, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "http://localhost"+path, strings.NewReader(body))
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
func TestSidePromptHTTPStreamingRetryNoPersistenceAndExplicitInjection(t *testing.T) {
	s := sidePromptFixture(t)
	for i := 0; i < 2; i++ {
		w := sideRequest(s, "/agent/side-prompt/stream", `{"prompt":"side question","chat_jid":"gi:s"}`)
		if w.Code != 200 || !strings.Contains(w.Body.String(), "side_prompt_start") || !strings.Contains(w.Body.String(), "side_prompt_text_delta") || !strings.Contains(w.Body.String(), "side_prompt_done") || !strings.Contains(w.Body.String(), "side answer") {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	w := sideRequest(s, "/agent/side-prompt", `{"prompt":"fail-side","chat_jid":"gi:s"}`)
	if w.Code != 502 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = sideRequest(s, "/agent/side-prompt/stream", `{"prompt":"fail-side","chat_jid":"gi:s"}`)
	if !strings.Contains(w.Body.String(), "side_prompt_error") || strings.Contains(w.Body.String(), "side_prompt_done") {
		t.Fatal(w.Body.String())
	}
	w = sideRequest(s, "/agent/side-prompt", `{"prompt":"side question","chat_jid":"gi:s"}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"status":"success"`) || w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal(w.Code, w.Body.String())
	}
	messages, _ := s.store.ListMessages(context.Background(), "s")
	turns, _ := s.store.ListTurns(context.Background(), "s")
	if len(messages) != 1 || len(turns) != 0 {
		t.Fatal("side answer persisted")
	}
	// Injection is normal admission; no side-specific write or execution path.
	w = sideRequest(s, "/api/sessions/s/prompt", `{"prompt":"BTW question: side question\n\nBTW answer: side answer"}`)
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	var admitted struct {
		TurnID string `json:"turn_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &admitted); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		row, err := s.store.GetTurn(context.Background(), admitted.TurnID)
		if err != nil {
			t.Fatal(err)
		}
		if row.Status == "completed" {
			break
		}
		if row.Status == "failed" || time.Now().After(deadline) {
			t.Fatalf("injection turn: %+v", row)
		}
		time.Sleep(10 * time.Millisecond)
	}
	messages, err := s.store.ListMessages(context.Background(), "s")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range messages {
		if m.Role == "user" && strings.Contains(m.Content, "BTW answer: side answer") {
			found = true
		}
	}
	if !found {
		t.Fatalf("normal injection missing response=%s messages=%+v", w.Body.String(), messages)
	}
}
func TestSidePromptHTTPValidationAuthAndConcurrency(t *testing.T) {
	s := sidePromptFixture(t)
	for _, body := range []string{`{}`, `{"prompt":"q","chat_jid":"other:s"}`, `{"prompt":"q","chat_jid":"gi:s","tools":[]}`, `{"prompt":"q","chat_jid":"gi:s"}{}`} {
		w := sideRequest(s, "/agent/side-prompt", body)
		if w.Code != 400 {
			t.Fatal(body, w.Code)
		}
	}
	if w := sideRequest(s, "/agent/side-prompt", `{"prompt":"q","chat_jid":"gi:missing"}`); w.Code != 404 {
		t.Fatal(w.Code)
	}
	for i := 0; i < cap(s.sidePromptSlots); i++ {
		s.sidePromptSlots <- struct{}{}
	}
	w := sideRequest(s, "/agent/side-prompt", `{"prompt":"q","chat_jid":"gi:s"}`)
	if w.Code != 429 {
		t.Fatal(w.Code)
	}
	for i := 0; i < cap(s.sidePromptSlots); i++ {
		<-s.sidePromptSlots
	}
	auth := authSessionServer(t)
	w = sideRequest(auth, "/agent/side-prompt", `{"prompt":"q","chat_jid":"gi:s"}`)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
	r := httptest.NewRequest(http.MethodPost, "http://localhost/agent/side-prompt", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("Origin", "https://evil.example")
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
}

func TestSidePromptHTTPDisconnectCancelsProviderAndReleasesSlot(t *testing.T) {
	cancelled := make(chan struct{})
	s := sidePromptFixture(t, func() { close(cancelled) })
	server := httptest.NewServer(s.Handler())
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/agent/side-prompt/stream", strings.NewReader(`{"prompt":"wait-side","chat_jid":"gi:s"}`))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		data, _ := io.ReadAll(response.Body)
		t.Fatal(response.StatusCode, string(data))
	}
	reader := bufio.NewReader(response.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(line, "side_prompt_text_delta") {
			break
		}
	}
	cancel()
	response.Body.Close()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("provider remained active after disconnect")
	}
	deadline := time.Now().Add(time.Second)
	for len(s.sidePromptSlots) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("side slot leaked")
		}
		time.Sleep(time.Millisecond)
	}
	messages, err := s.store.ListMessages(context.Background(), "s")
	if err != nil || len(messages) != 1 {
		t.Fatal(messages, err)
	}
	// Retry reuses no retained side state and can complete normally.
	retry, err := http.Post(server.URL+"/agent/side-prompt", "application/json", strings.NewReader(`{"prompt":"side question","chat_jid":"gi:s"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer retry.Body.Close()
	data, err := io.ReadAll(retry.Body)
	if err != nil || retry.StatusCode != 200 || !bytes.Contains(data, []byte("side answer")) {
		t.Fatal(retry.StatusCode, string(data), err)
	}
}

func TestSidePromptCompletesBesideRunningMainTurn(t *testing.T) {
	s := sidePromptFixture(t)
	w := sideRequest(s, "/api/sessions/s/prompt", `{"prompt":"main-held"}`)
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	var main struct {
		TurnID string `json:"turn_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &main); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		events, err := s.store.ListTurnEvents(context.Background(), main.TurnID)
		if err != nil {
			t.Fatal(err)
		}
		started := false
		for _, event := range events {
			if event.Type == "inference.started" {
				started = true
			}
		}
		if started {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("main inference not started", events)
		}
		time.Sleep(time.Millisecond)
	}
	before, err := s.store.ListMessages(context.Background(), "s")
	if err != nil {
		t.Fatal(err)
	}
	w = sideRequest(s, "/agent/side-prompt", `{"prompt":"independent question","chat_jid":"gi:s"}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	active, err := s.store.GetTurn(context.Background(), main.TurnID)
	if err != nil || active.Status != "running" {
		t.Fatal(active, err)
	}
	after, err := s.store.ListMessages(context.Background(), "s")
	if err != nil || len(after) != len(before) {
		t.Fatal(after, err)
	}
	turns, err := s.store.ListTurns(context.Background(), "s")
	if err != nil || len(turns) != 1 {
		t.Fatal(turns, err)
	}
	if err := s.turns.CancelActiveTurn(context.Background(), "s", main.TurnID); err != nil {
		t.Fatal(err)
	}
}
