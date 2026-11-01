package web

import (
	"context"
	"encoding/json"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestConversationDoesNotExposeProviderReplayPayload(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "privacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: t.TempDir()}).Handler()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "private-replay", "session", nil); err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{
		"kind": "chat", "provider_message": map[string]any{"content": "REPLAY_SECRET"},
		"provider_partial":           map[string]any{"content": "PARTIAL_SECRET"},
		"provider_received_terminal": map[string]any{"content": "TERMINAL_SECRET"},
		"provider_content_blocks":    []any{map[string]any{"type": "image", "data": "IMAGE_SECRET"}},
		"added_tool_names":           []string{"INTERNAL_LOADOUT"}, "content_blocks": []any{map[string]any{"type": "text", "text": "public block"}},
	}
	if err := s.AddMessage(ctx, "private-row", "private-replay", "assistant", "visible", payload); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/sessions/private-replay/messages?view=conversation&limit=50", "/api/sessions/private-replay/search?view=conversation&q=visible&scope=current"} {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != 200 {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		for _, secret := range []string{"REPLAY_SECRET", "PARTIAL_SECRET", "TERMINAL_SECRET", "provider_received_terminal", "IMAGE_SECRET", "INTERNAL_LOADOUT", "provider_message", "provider_partial", "provider_content_blocks"} {
			if strings.Contains(body, secret) {
				t.Fatalf("%s leaked %s", path, secret)
			}
		}
		if !strings.Contains(body, "public block") {
			t.Fatalf("display blocks lost: %s", body)
		}
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/sessions/private-replay/messages", nil))
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rec.Body.String(), "REPLAY_SECRET") || !strings.Contains(rec.Body.String(), "PARTIAL_SECRET") || !strings.Contains(rec.Body.String(), "TERMINAL_SECRET") {
		t.Fatal("raw export lost replay payload")
	}
	snapshot, err := s.ContextSnapshot(ctx, "private-replay")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Messages[0].Payload["provider_message"] == nil || snapshot.Messages[0].Payload["provider_partial"] == nil || snapshot.Messages[0].Payload["provider_received_terminal"] == nil {
		t.Fatal("provider history lost replay payload")
	}
}
