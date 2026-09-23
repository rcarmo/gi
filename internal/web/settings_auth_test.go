package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

// @gi-settings-008: the derived UI reuses the existing authenticated routes.
func TestGiSettingsRoutesPreserveAuthentication(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gi")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	blob, _ := json.Marshal(map[string]any{"username": "admin", "totp_enabled": true, "totp_secret": "SECRET-NOT-FOR-SETTINGS"})
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), blob, 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	srv := New(st, turn.New(st), config.RuntimeConfig{WorkspaceRoot: workspace})
	for _, item := range []struct{ method, path, body string }{
		{"GET", "/api/runtime/config", ""},
		{"GET", "/api/settings/identity", ""},
		{"GET", "/api/settings/providers", ""},
		{"PATCH", "/api/settings/providers", "not json"},
		{"DELETE", "/api/settings/providers", "not json"},
		{"GET", "/api/settings/compaction", ""},
		{"PATCH", "/api/settings/compaction", "not json"},
		{"PATCH", "/api/settings/identity", "not json"},
		{"GET", "/api/sessions/missing/model", ""},
		{"GET", "/api/sessions/missing/compaction", ""},
		{"POST", "/api/sessions/missing/compaction", "not json"},
		{"POST", "/api/sessions/missing/activity", "not json"},
		{"PATCH", "/api/sessions/missing/model", "not json"},
	} {
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(item.method, item.path, bytes.NewBufferString(item.body)))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: %d %s", item.method, item.path, rec.Code, rec.Body.String())
		}
		if bytes.Contains(rec.Body.Bytes(), []byte("SECRET-NOT-FOR-SETTINGS")) {
			t.Fatal("authentication secret leaked")
		}
	}
}
