package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestAuthEnrollStartReturnsInternalServerErrorForCorruptState(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".gi"), 0o755); err != nil {
		t.Fatalf("create .gi dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gi", "auth.json"), []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt auth file: %v", err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/enroll/start", bytes.NewBufferString(`{"username":"rui"}`))
	req.RemoteAddr = "127.0.0.1:1234"
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for corrupt auth state, got %d body=%s", res.Code, res.Body.String())
	}
}
