package web

import (
	"bytes"
	"encoding/json"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
