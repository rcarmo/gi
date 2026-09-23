package web

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/search/chunking"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestWorkspaceIndexAPIRefreshQueriesFailureAndAuth(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	write := func(path, text string) {
		t.Helper()
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("notes/a.md", "native orchid")
	write(".pi/skills/test/SKILL.md", "native violet")
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})
	call := func(method, path string) *httptest.ResponseRecorder {
		t.Helper()
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, httptest.NewRequest(method, path, nil))
		return res
	}
	status := func() *workspaceIndexStatus {
		t.Helper()
		res := call("GET", "/api/workspace/index")
		if res.Code != 200 {
			t.Fatal(res.Code, res.Body.String())
		}
		var out workspaceIndexStatus
		if err := json.Unmarshal(res.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return &out
	}
	query := func(scope, term string) []searchstore.LexicalHit {
		t.Helper()
		res := call("GET", "/api/workspace/search?scope="+scope+"&q="+term)
		if res.Code != 200 {
			t.Fatal(res.Code, res.Body.String())
		}
		var out struct {
			Hits []searchstore.LexicalHit `json:"hits"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out.Hits
	}
	if status().State != "never_indexed" || len(query("all", "native")) != 0 || status().State != "never_indexed" {
		t.Fatal("read mutated index")
	}
	res := call("POST", "/api/workspace/index")
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	ready := status()
	if ready.State != "ready" || ready.IndexedFileCount != 2 || ready.Generation != 1 || !ready.RequiredRoots {
		t.Fatal(ready)
	}
	if len(query("all", "native")) != 2 || len(query("notes", "native")) != 0 {
		t.Fatal("scope membership incorrectly inferred")
	}
	if res := call("POST", "/api/workspace/index?scope=notes"); res.Code != 200 {
		t.Fatal(res.Code)
	}
	if len(query("notes", "orchid")) != 1 || len(query("notes", "violet")) != 0 {
		t.Fatal("scope leak")
	}
	// Failed scan preserves committed content/status; GET never retries it.
	if err := os.Rename(filepath.Join(root, "notes"), filepath.Join(root, "saved-notes")); err != nil {
		t.Fatal(err)
	}
	if res := call("POST", "/api/workspace/index"); res.Code != 500 {
		t.Fatal(res.Code)
	}
	failed := status()
	if failed.State != "failed" || failed.Generation != ready.Generation || failed.LastIndexedAt != ready.LastIndexedAt || failed.IndexedFileCount != 2 || failed.LastError == "" {
		t.Fatal(failed)
	}
	if len(query("all", "orchid")) != 1 || status().State != "failed" {
		t.Fatal("read refreshed or lost snapshot")
	}
	if err := os.Rename(filepath.Join(root, "saved-notes"), filepath.Join(root, "notes")); err != nil {
		t.Fatal(err)
	}
	write("notes/a.md", "native changed")
	c, err := searchstore.DefaultScopeConfig(root, "all", nil, nil, chunking.LineVersion)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := searchstore.NewRefreshStore(s.DB()).Begin(t.Context(), c, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if res := call("POST", "/api/workspace/index"); res.Code != 409 {
		t.Fatal("competing API worker", res.Code)
	}
	if status().State != "indexing" || len(query("all", "orchid")) != 1 {
		t.Fatal("busy query altered snapshot")
	}
	if err := owner.Fail(t.Context(), os.ErrClosed); err != nil {
		t.Fatal(err)
	}
	if res := call("POST", "/api/workspace/index"); res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	if len(query("all", "orchid")) != 0 || len(query("all", "changed")) != 1 {
		t.Fatal("retry failed")
	}
	for _, path := range []string{"/api/workspace/index?scope=bad", "/api/workspace/search?q=x&scope=bad", "/api/workspace/search?q=x&limit=51", "/api/workspace/search?q=x&offset=-1", "/api/workspace/search", "/api/workspace/search?q=x&refresh=true"} {
		if res := call("GET", path); res.Code != 400 {
			t.Fatal(path, res.Code)
		}
	}
	if call("DELETE", "/api/workspace/index").Code != 405 || call("POST", "/api/workspace/search?q=x").Code != 405 {
		t.Fatal("methods")
	}
	// New server instance reads the same durable status; no implicit rebuild.
	generation := status().Generation
	srv = New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})
	if status().Generation != generation {
		t.Fatal("server recreation mutated generation")
	}
	pending, err := srv.auth.StartEnrollment("rui")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = srv.auth.VerifyEnrollment("rui", webTestTOTPCode(pending.Secret)); err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"GET", "POST"} {
		if call(method, "/api/workspace/index").Code != 401 {
			t.Fatal("unguarded")
		}
	}
	if call("GET", "/api/workspace/search?q=native").Code != 401 {
		t.Fatal("unguarded search")
	}
}
