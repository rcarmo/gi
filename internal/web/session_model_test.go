package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSessionModelCommandsAreValidatedLocalAndDoNotCreateTurns(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "models.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cfg := config.RuntimeConfig{DefaultProvider: "test", DefaultModel: "test-model", EnabledModels: []string{"test-model", "bootstrap", "test/unavailable-model"}, WorkspaceRoot: t.TempDir()}
	e := turn.NewWithRuntimeConfig(s, cfg, "")
	defer e.Close()
	server := New(s, e, cfg)
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, map[string]any{"model": "test-model", "provider": "test", "status": "idle", "thinking_level": "low"}); err != nil {
			t.Fatal(err)
		}
	}
	call := func(method, path, body string, status int) map[string]any {
		t.Helper()
		w := httptest.NewRecorder()
		server.handleSessionSubroutes(w, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
		if w.Code != status {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		var value map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	state := call("PATCH", "/api/sessions/A/model", `{"model":"test/bootstrap"}`, 200)
	if state["current"] != "test/bootstrap" || state["context_usage"] != nil {
		t.Fatalf("incorrect authoritative state %+v", state)
	}
	for _, name := range []string{"unknown", "test/unavailable-model"} {
		call("PATCH", "/api/sessions/A/model", `{"model":"`+name+`"}`, 400)
	}
	call("POST", "/api/sessions/A/prompt", `{"prompt":"/model unknown"}`, 400)
	call("POST", "/api/sessions/A/prompt", `{"prompt":"/model test/bootstrap"}`, 200)
	call("POST", "/api/sessions/A/prompt", `{"prompt":"/model"}`, 200)
	for _, id := range []string{"A", "B"} {
		turns, _ := s.ListTurns(ctx, id)
		if len(turns) != 0 {
			t.Fatal("model command created turn")
		}
	}
	a, _ := s.GetSession(ctx, "A")
	b, _ := s.GetSession(ctx, "B")
	if a.State["model"] != "bootstrap" || a.State["thinking_level"] != "" || b.State["model"] != "test-model" || server.cfg.DefaultModel != "test-model" {
		t.Fatalf("model scope changed A=%v B=%v", a.State, b.State)
	}
	// A completing old turn can update runtime model metadata; it must not
	// overwrite the separately persisted user selection.
	if err := s.TouchSessionState(ctx, "A", map[string]any{"model": "test-model"}); err != nil {
		t.Fatal(err)
	}
	if current := call("GET", "/api/sessions/A/model", "", 200)["current"]; current != "test/bootstrap" {
		t.Fatal("old runtime model replaced selection")
	}
	// Ordinary prompts without an override must use the selected session model.
	call("POST", "/api/sessions/A/prompt", `{"prompt":"selected model prompt"}`, 202)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		turns, _ := s.ListTurns(ctx, "A")
		if len(turns) > 0 && turns[0].Status == "completed" {
			if turns[0].Metadata["model"] != "bootstrap" {
				t.Fatalf("selection ignored: %v", turns[0].Metadata)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("selected-model turn did not complete")
}
