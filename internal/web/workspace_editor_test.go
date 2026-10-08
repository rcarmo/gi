package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
)

func TestWorkspaceEditorCompleteReadsWritesAndCompatibilityRoutes(t *testing.T) {
	root := t.TempDir()
	content := strings.Repeat("a unicode line αβ\n", 2500)
	path := filepath.Join(root, "editable.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	before := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(path, before, before); err != nil {
		t.Fatal(err)
	}
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: root})
	request := func(method, target, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, target, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		return w
	}
	decode := func(w *httptest.ResponseRecorder) map[string]any {
		t.Helper()
		var p map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
			t.Fatal(err, w.Body.String())
		}
		return p
	}
	for _, prefix := range []string{"/api/workspace", "/workspace"} {
		w := request(http.MethodGet, prefix+"/file?path=editable.md&mode=edit&max=1000000", "")
		p := decode(w)
		if w.Code != 200 || p["text"] != content || p["truncated"] != false || p["mtime"] != before.Format(time.RFC3339Nano) {
			t.Fatal(w.Code, p)
		}
		w = request(http.MethodGet, prefix+"/file?path=editable.md&max=4096", "")
		p = decode(w)
		if w.Code != 200 || len(p["text"].(string)) > 4096 || p["truncated"] != true {
			t.Fatal(w.Code, p)
		}
		w = request(http.MethodGet, prefix+"/stat?path=editable.md", "")
		if w.Code != 200 || decode(w)["mtime"] != before.Format(time.RFC3339Nano) {
			t.Fatal(w.Code, w.Body.String())
		}
		w = request(http.MethodGet, prefix+"/raw?path=editable.md", "")
		if w.Code != 200 || w.Body.String() != content {
			t.Fatal(w.Code)
		}
	}
	missing := request(http.MethodGet, "/workspace/file?path=missing.md&mode=edit", "")
	if missing.Code != http.StatusNotFound {
		t.Fatal(missing.Code, missing.Body.String())
	}
	boundary := strings.Repeat("x", workspaceMaxEditBytes)
	if err := os.WriteFile(filepath.Join(root, "boundary.txt"), []byte(boundary), 0o600); err != nil {
		t.Fatal(err)
	}
	complete := request(http.MethodGet, "/workspace/file?path=boundary.txt&mode=edit&max=1", "")
	if complete.Code != 200 || decode(complete)["text"] != boundary {
		t.Fatal("edit read truncated at client preview limit", complete.Code)
	}
	body, _ := json.Marshal(map[string]string{"path": "editable.md", "content": content})
	w := request(http.MethodPut, "/workspace/file", string(body))
	if w.Code != 200 || decode(w)["mtime"] != before.Format(time.RFC3339Nano) {
		t.Fatal(w.Code, w.Body.String())
	}
	body, _ = json.Marshal(map[string]string{"path": "editable.md", "content": "changed"})
	w = request(http.MethodPut, "/workspace/file", string(body))
	if w.Code != 200 || decode(w)["size"] != float64(7) {
		t.Fatal(w.Code, w.Body.String())
	}
	data, _ := os.ReadFile(path)
	info, _ := os.Stat(path)
	if string(data) != "changed" || info.Mode().Perm() != 0o600 {
		t.Fatal(string(data), info.Mode())
	}
}

func TestWorkspaceEditorRejectsTruncatedBinaryOutsideAndUnauthenticatedAccess(t *testing.T) {
	root := t.TempDir()
	for name, data := range map[string][]byte{
		"large.txt":   bytes.Repeat([]byte("x"), workspaceMaxEditBytes+1),
		"binary.txt":  []byte("before\x00after"),
		"invalid.txt": {0xff, 0xfe},
	} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Fatal(err)
	}
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: root})
	for _, path := range []string{"large.txt", "binary.txt", "invalid.txt", "outside", "../secret.txt", outside} {
		r := httptest.NewRequest(http.MethodGet, "/workspace/file?mode=edit&path="+url.QueryEscape(path), nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		if w.Code != 400 || strings.Contains(w.Body.String(), "before") || strings.Contains(w.Body.String(), "outside\"") {
			t.Fatal(path, w.Code, w.Body.String())
		}
	}
	body := `{"path":"binary.txt","content":"hello\u0000world"}`
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/workspace/file", strings.NewReader(body)))
	if w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	authed := authSessionServer(t)
	for _, path := range []string{"/workspace/file?path=a.md", "/workspace/raw?path=a.md", "/workspace/stat?path=a.md"} {
		for _, method := range []string{http.MethodGet, http.MethodPut} {
			w := httptest.NewRecorder()
			authed.Handler().ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 401 || w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Fatal(path, method, w.Code)
			}
		}
	}
}

func TestWorkspaceEditorStaticCompatibilityPaths(t *testing.T) {
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	for _, path := range []string{"/static/dist/app.bundle.css", "/static/editor-vendor/codemirror.js", "/static/dist/editor.bundle.js", "/dist/editor.bundle.js", "/static/dist/app.bundle.js"} {
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 200 || w.Body.Len() == 0 {
			t.Fatal(path, w.Code)
		}
	}
	for _, path := range []string{"/static/workspace/file?path=secret", "/static/api/sessions"} {
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 404 {
			t.Fatal(path, w.Code)
		}
	}
}
