package web

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
)

func TestQuickActionsCatalogueIsReadOnlyAuthenticatedAndCapabilityLimited(t *testing.T) {
	root := t.TempDir()
	db, err := store.Open(filepath.Join(t.TempDir(), "actions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := New(db, nil, config.RuntimeConfig{WorkspaceRoot: root})
	call := func(method string) *httptest.ResponseRecorder {
		r := httptest.NewRecorder()
		s.Handler().ServeHTTP(r, httptest.NewRequest(method, "/api/quick-actions", nil))
		return r
	}
	r := call("GET")
	if r.Code != 200 || r.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal(r.Code, r.Body.String())
	}
	var payload struct {
		WorkspaceCommands, SlashCommands []string
		Commands                         []map[string]string
	}
	if err := json.Unmarshal(r.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.WorkspaceCommands) != 2 || len(payload.SlashCommands) != 2 || len(payload.Commands) != 2 {
		t.Fatal(payload)
	}
	if payload.Commands[0]["name"] != "/model" || payload.Commands[1]["name"] != "/compact" {
		t.Fatal(payload)
	}
	if call("POST").Code != 405 {
		t.Fatal("mutable catalogue")
	}
	var n int
	if err := db.DB().QueryRow("SELECT count(*) FROM turns").Scan(&n); err != nil || n != 0 {
		t.Fatal(n, err)
	}
	pending, err := s.auth.StartEnrollment("rui")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.auth.VerifyEnrollment("rui", webTestTOTPCode(pending.Secret)); err != nil {
		t.Fatal(err)
	}
	if call("GET").Code != 401 {
		t.Fatal("unguarded catalogue")
	}
}
