package web

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestWorkspaceTreeQueriesConfinementAndBounds(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, path := range []string{"a/nested/deep/leaf.txt", "a/.hidden.txt", ".pi/config", ".git/config", ".gitignore", "node_modules/pkg/index.js", "z.txt"} {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(path), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "empty"), 0700); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("a", filepath.Join(root, "inside")); err != nil {
		t.Fatal(err)
	}
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})
	call := func(method, query string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRecorder()
		srv.Handler().ServeHTTP(r, httptest.NewRequest(method, "/api/workspace/tree"+query, nil))
		return r
	}
	get := func(query string) workspaceNode {
		t.Helper()
		r := call("GET", query)
		if r.Code != 200 {
			t.Fatal(query, r.Code, r.Body.String())
		}
		var n workspaceNode
		if err := json.Unmarshal(r.Body.Bytes(), &n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	names := func(n workspaceNode) string {
		var names []string
		for _, c := range n.Children {
			names = append(names, c.Name)
		}
		return strings.Join(names, ",")
	}
	legacy := get("")
	if !strings.Contains(names(legacy), ".pi") || strings.Contains(names(legacy), ".git") || strings.Contains(names(legacy), "node_modules") {
		t.Fatal(names(legacy))
	}
	visible := get("?path=.&depth=1&show_hidden=false")
	if names(visible) != "a,empty,inside,node_modules,z.txt" {
		t.Fatal(names(visible))
	}
	all := get("?path=.&depth=1&show_hidden=true")
	if !strings.Contains(names(all), ".git,") || !strings.Contains(names(all), ".gitignore") {
		t.Fatal(names(all))
	}
	for _, c := range all.Children {
		if c.Type == "dir" && c.Children != nil {
			t.Fatal("depth-1 stub expanded", c)
		}
	}
	empty := get("?path=empty&depth=1&show_hidden=false")
	if empty.Children == nil || len(empty.Children) != 0 {
		t.Fatal("missing explicit empty children", empty)
	}
	deep := get("?path=a/nested/deep&depth=1&show_hidden=false")
	if deep.Path != "a/nested/deep" || names(deep) != "leaf.txt" {
		t.Fatal(deep)
	}
	if names(get("?path=a&depth=1&show_hidden=false")) != "nested" {
		t.Fatal("hidden included")
	}
	if names(get("?path=a&depth=1&show_hidden=true")) != "nested,.hidden.txt" {
		t.Fatal("hidden omitted")
	}
	if names(get("?path=inside/nested/deep&depth=1&show_hidden=true")) != "leaf.txt" {
		t.Fatal("in-root symlink unavailable")
	}
	get("?path=.&depth=8&show_hidden=true")
	for _, q := range []string{"?path=../", "?path=/etc", "?path=outside", "?path=missing", "?depth=9", "?depth=-1", "?depth=x", "?show_hidden=maybe"} {
		if r := call("GET", q); r.Code != 400 {
			t.Fatal(q, r.Code)
		}
	}
	if r := call("POST", ""); r.Code != 405 {
		t.Fatal(r.Code)
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()
	budget := 1
	if _, err := readWorkspaceTree(dir, ".", 1, true, false, &budget); err != errWorkspaceTreeLimit {
		t.Fatal("budget not enforced", err)
	}
	pending, err := srv.auth.StartEnrollment("rui")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = srv.auth.VerifyEnrollment("rui", webTestTOTPCode(pending.Secret)); err != nil {
		t.Fatal(err)
	}
	if r := call("GET", "?path=.&depth=1&show_hidden=true"); r.Code != 401 {
		t.Fatal("unguarded", r.Code)
	}
}
