package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSessionMetadataMutations(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	root, err := s.CreateSession(ctx, "main", "@main", map[string]any{"status": "idle", "model": "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := s.CloneSession(ctx, root.ID, "child", "@child", "child-agent")
	if err != nil {
		t.Fatal(err)
	}
	server := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	defer server.turns.Close()
	mutate := func(id, body string, status int) {
		t.Helper()
		r := httptest.NewRequest(http.MethodPatch, "/api/sessions/"+id, bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		server.handleSession(w, r, id)
		if w.Code != status {
			t.Fatalf("%s %s: status %d, want %d: %s", id, body, w.Code, status, w.Body.String())
		}
		if status != http.StatusOK {
			var result map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result["error"] == nil {
				t.Fatalf("missing mutation error: %s", w.Body.String())
			}
		}
	}
	mutate("missing", `{"action":"restore"}`, 404)
	mutate(child.ID, `{"action":"rename","title":"  "}`, 400)
	mutate(child.ID, `{"action":"pin"}`, 400)
	mutate(child.ID, `{"action":"delete"}`, 400)
	mutate(child.ID, `{"action":"restore","unknown":true}`, 400)
	mutate(child.ID, `{"action":"restore"}{"action":"archive"}`, 400)
	mutate(child.ID, `{"action":"pin","pinned":"yes"}`, 400)
	mutate(root.ID, `{"action":"archive"}`, 409)
	mutate(child.ID, `{"action":"rename","title":" New display name "}`, 200)
	mutate(child.ID, `{"action":"pin","pinned":true}`, 200)
	mutate(child.ID, `{"action":"archive"}`, 200)
	after, err := s.GetSession(ctx, child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Title != "New display name" || after.State["pinned"] != true || after.State["archived_at"] == nil {
		t.Fatalf("metadata not persisted: %+v", after)
	}
	if after.Scope.AgentID != child.Scope.AgentID || after.ParentSessionID != child.ParentSessionID || after.State["model"] != child.State["model"] {
		t.Fatalf("mutation changed identity/ancestry/model: %+v", after)
	}
	cloned, err := s.CloneSession(ctx, child.ID, "new-child", "New", "new-agent")
	if err != nil {
		t.Fatal(err)
	}
	if cloned.State["archived_at"] != nil || cloned.State["pinned"] != nil || cloned.State["queue_count"] != float64(0) {
		t.Fatalf("fork inherited picker metadata or queue: %+v", cloned.State)
	}
	archivedAt := after.State["archived_at"]
	mutate(child.ID, `{"action":"archive"}`, 200)
	after, _ = s.GetSession(ctx, child.ID)
	if after.State["archived_at"] != archivedAt {
		t.Fatal("archive is not idempotent")
	}
	mutate(child.ID, `{"action":"rename","title":"Blocked"}`, 409)
	mutate(child.ID, `{"action":"pin","pinned":false}`, 409)
	mutate(child.ID, `{"action":"restore"}`, 200)
	mutate(child.ID, `{"action":"pin","pinned":false}`, 200)
	after, _ = s.GetSession(ctx, child.ID)
	if after.State["archived_at"] != nil || after.State["pinned"] != false {
		t.Fatalf("restore/unpin failed: %+v", after.State)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "queued", child.ID, "queued", "pending", nil); err != nil {
		t.Fatal(err)
	}
	mutate(child.ID, `{"action":"archive"}`, 409)
	if err := s.UpdateTurnStatus(ctx, "queued", "completed"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, child.ID, "queued", "test", "token"); err != nil {
		t.Fatal(err)
	}
	mutate(child.ID, `{"action":"archive"}`, 409)
	if err := s.ReleaseSessionActiveTurn(ctx, child.ID, "token"); err != nil {
		t.Fatal(err)
	}
	mutate(child.ID, `{"action":"archive"}`, 200)
}
