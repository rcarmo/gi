package web

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/search/chunking"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	"github.com/rcarmo/gi/internal/store"
)

func TestWorkspaceIndexAPIRefreshQueriesFailureAndAuth(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
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
	srv := newIndexTestServer(t, s, config.RuntimeConfig{WorkspaceRoot: root})
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
	srv.CloseWorkspaceIndex()
	srv = newIndexTestServer(t, s, config.RuntimeConfig{WorkspaceRoot: root})
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

func TestWorkspaceIndexStartupSettingsOptionalPolicyAndConfigIsolation(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "settings.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
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
	write(".pi/settings.json", `{"workspaceIndex":{"extraRoots":["docs"],"extraExtensions":["nim"],"optionalRoots":["notes",".pi/skills"]}}`)
	write("docs/sample.nim", "nimorchid evidence")
	srv := newIndexTestServer(t, s, config.Load(root))
	call := func(method, path string) *httptest.ResponseRecorder {
		t.Helper()
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, httptest.NewRequest(method, path, nil))
		return res
	}
	current := func() workspaceIndexStatus {
		t.Helper()
		r := call("GET", "/api/workspace/index")
		if r.Code != 200 {
			t.Fatal(r.Code, r.Body.String())
		}
		var st workspaceIndexStatus
		if err := json.Unmarshal(r.Body.Bytes(), &st); err != nil {
			t.Fatal(err)
		}
		return st
	}
	if res := call("POST", "/api/workspace/index"); res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	before := current()
	if before.RequiredRoots || len(before.OptionalRoots) != 2 || before.IndexedFileCount != 1 {
		t.Fatal(before)
	}
	if res := call("GET", "/api/workspace/search?q=nimorchid"); res.Code != 200 || !strings.Contains(res.Body.String(), "docs/sample.nim") {
		t.Fatal(res.Code, res.Body.String())
	}
	// A caller cannot override roots by query parameters. A settings edit requires
	// a reload/new server, and new fingerprints hide the previous configuration.
	write(".pi/settings.json", `{"workspaceIndex":{"extraRoots":["other"],"extraExtensions":["nim"],"optionalRoots":["notes",".pi/skills"]}}`)
	if st := current(); st.ConfigHash != before.ConfigHash || st.State != "ready" {
		t.Fatal("hot config unexpectedly changed", st)
	}
	write("other/new.nim", "newviolet content")
	srv.CloseWorkspaceIndex()
	srv = newIndexTestServer(t, s, config.Load(root))
	if st := current(); st.State != "stale" {
		t.Fatal(st)
	}
	var hits struct {
		Hits []any `json:"hits"`
	}
	res := call("GET", "/api/workspace/search?q=nimorchid")
	if err := json.Unmarshal(res.Body.Bytes(), &hits); err != nil || len(hits.Hits) != 0 {
		t.Fatal(hits, err)
	}
	if res := call("POST", "/api/workspace/index"); res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	after := current()
	if after.ConfigHash == before.ConfigHash || after.Generation != before.Generation+1 {
		t.Fatal(after)
	}
	write("notes/kept.md", "retained optional content")
	if res := call("POST", "/api/workspace/index"); res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	kept := current()
	if err := os.Rename(filepath.Join(root, "notes"), filepath.Join(root, "held")); err != nil {
		t.Fatal(err)
	}
	if res := call("POST", "/api/workspace/index"); res.Code != 500 || !strings.Contains(res.Body.String(), "disappeared") {
		t.Fatal(res.Code, res.Body.String())
	}
	failed := current()
	if failed.Generation != kept.Generation || failed.LastIndexedAt != kept.LastIndexedAt || failed.IndexedFileCount != kept.IndexedFileCount || failed.State != "failed" {
		t.Fatal(failed)
	}
	if res := call("GET", "/api/workspace/search?q=retained"); !strings.Contains(res.Body.String(), "notes/kept.md") {
		t.Fatal(res.Body.String())
	}
	// Invalid options fail requests; no silently broadened all-scope fallback.
	write(".pi/settings.json", `{"workspaceIndex":{"extraRoots":["../escape"]}}`)
	srv.CloseWorkspaceIndex()
	srv = newIndexTestServer(t, s, config.Load(root))
	for _, method := range []string{"GET", "POST"} {
		if res := call(method, "/api/workspace/index"); res.Code != 400 {
			t.Fatal(res.Code)
		}
	}
}
