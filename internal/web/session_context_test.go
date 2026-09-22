package web

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSessionModelPayloadExposesMeasuredInputAndIsolatesUnknown(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "context.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := turn.New(s)
	defer e.Close()
	server := New(s, e, config.RuntimeConfig{DefaultProvider: "opencode-zen", DefaultModel: "minimax-m2.5-free", EnabledModels: []string{"opencode-zen/minimax-m2.5-free"}})
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, map[string]any{"model": "opencode-zen/minimax-m2.5-free"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateTurn(ctx, "turn", "A", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendTurnEvent(ctx, "turn", "A", "context.measured", map[string]any{"tokens": 96000, "input": 80000, "cache_read": 16000, "model": "opencode-zen/minimax-m2.5-free", "iteration": 2}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"A", "B"} {
		w := httptest.NewRecorder()
		server.handleSessionModel(w, httptest.NewRequest("GET", "/api/sessions/"+id+"/model", nil), id)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var payload map[string]any
		json.Unmarshal(w.Body.Bytes(), &payload)
		usage := payload["context_usage"].(map[string]any)
		if usage["contextWindow"] != float64(128000) {
			t.Fatal("missing catalogue window", usage)
		}
		if id == "A" {
			if usage["tokens"] != float64(96000) || usage["percent"] != float64(75) || usage["source"] != "provider_request" {
				t.Fatal(usage)
			}
		} else if usage["tokens"] != nil || usage["percent"] != nil {
			t.Fatal("measurement crossed session", usage)
		}
	}
}
