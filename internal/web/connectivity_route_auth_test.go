package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestConnectivityRouteReturnsInternalServerErrorForCorruptTOTPAuthState(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".gi"), 0o755); err != nil {
		t.Fatalf("create .gi dir: %v", err)
	}
	authState := `{"username":"admin","totp_secret":"JBSWY3DPEHPK3PXP","totp_enabled":true,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","sessions":[]}`
	if err := os.WriteFile(filepath.Join(root, ".gi", "auth.json"), []byte(authState), 0o600); err != nil {
		t.Fatalf("write auth state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gi", "auth.json"), []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("corrupt auth state: %v", err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.New(s)
	defer engine.Close()
	srv := New(s, engine, config.RuntimeConfig{WorkspaceRoot: root})
	info, err := engine.Connectivity().Register(t.Context(), connectivity.RouteSpec{Name: "totp", Transport: "http", Auth: map[string]any{"type": "totp"}}, nil)
	if err != nil {
		t.Fatalf("register route: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/connect/routes/"+info.ID, nil)
	req.RemoteAddr = "203.0.113.10:1234"
	res := httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for corrupt connectivity auth state, got %d body=%s", res.Code, res.Body.String())
	}
}
