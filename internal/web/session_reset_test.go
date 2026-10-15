package web

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSessionResetHTTP(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "reset-http.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := turn.New(s)
	defer e.Close()
	h := New(s, e, config.RuntimeConfig{WorkspaceRoot: t.TempDir()}).Handler()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "old", "s", "user", "old", nil); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/api/sessions/s/reset", "", 405},
		{"POST", "/api/sessions/s/reset", `{"unknown":true}`, 400},
		{"POST", "/api/sessions/s/reset", `{} {}`, 400},
		{"POST", "/api/sessions/s/reset", `{"text":"fresh"}`, 200},
	} {
		rec := httptest.NewRecorder()
		request := httptest.NewRequest(tc.method, "http://127.0.0.1"+tc.path, strings.NewReader(tc.body))
		request.RemoteAddr = "127.0.0.1:1234"
		request.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(rec, request)
		if rec.Code != tc.status {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}
	snapshot, err := s.ContextSnapshot(ctx, "s")
	if err != nil || len(snapshot.Messages) != 1 || snapshot.Messages[0].Content != "fresh" {
		t.Fatalf("reset API context: %+v %v", snapshot, err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "test", "claim"); err != nil || !ok {
		t.Fatalf("claim %t %v", ok, err)
	}
	rec := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "http://127.0.0.1/api/sessions/s/reset", strings.NewReader(`{}`))
	request.RemoteAddr = "127.0.0.1:1234"
	request.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, request)
	if rec.Code != 202 || !strings.Contains(rec.Body.String(), `"queued":true`) {
		t.Fatalf("busy response %d %s", rec.Code, rec.Body.String())
	}
}
