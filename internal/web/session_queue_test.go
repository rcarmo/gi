package web

import (
	"bytes"
	"context"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSessionQueueAPIRejectsForeignRunningAndStaleItems(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := turn.New(s)
	defer e.Close()
	server := New(s, e, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"one", "two"} {
		if _, err := s.CreateTurnWithStatus(ctx, id, "A", "queued", id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateTurnWithStatus(ctx, "active", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "A", "active", "test", "token"); err != nil {
		t.Fatal(err)
	}
	check := func(method, path, body string, want int) {
		t.Helper()
		w := httptest.NewRecorder()
		server.handleSessionSubroutes(w, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
	}
	check("GET", "/api/sessions/A/queue", "", 200)
	check("GET", "/api/sessions/missing/queue", "", 404)
	check("PATCH", "/api/sessions/A/queue", `{"expected":["one","two"],"order":["two","one"]}`, 200)
	check("PATCH", "/api/sessions/A/queue", `{"expected":["one","two"],"order":["two","one"]}`, 409)
	check("DELETE", "/api/sessions/B/queue/one", "", 404)
	check("DELETE", "/api/sessions/A/queue/active", "", 409)
	check("DELETE", "/api/sessions/A/queue/one", "", 200)
	check("DELETE", "/api/sessions/A/queue/one", "", 200)
	check("POST", "/api/sessions/A/queue", "", 405)
	queue, err := s.ListQueuedTurns(ctx, "A")
	if err != nil || len(queue) != 1 || queue[0].ID != "two" {
		t.Fatalf("queue: %+v %v", queue, err)
	}
	active, _ := s.GetTurn(ctx, "active")
	if active.Status != "running" {
		t.Fatal("stale queue cancel stopped active turn")
	}
}
