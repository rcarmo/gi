package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	srv := New(s, turn.New(s), config.RuntimeConfig{AssistantName: "Neo", UserName: "Rui", DefaultProvider: "ollama", DefaultModel: "gemma4:latest", DefaultThinkingLevel: "medium"})

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"title":"Demo"}`))
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
	if !bytes.Contains(messagesRes.Body.Bytes(), []byte("Gi received: hello")) {
		t.Fatalf("unexpected messages body: %s", messagesRes.Body.String())
	}
}
