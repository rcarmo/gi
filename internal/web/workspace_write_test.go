package web

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestWorkspaceWritesFollowPiclawExplorerAPI(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o700); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Fatal(err)
	}
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})
	call := func(method, target, contentType string, body []byte) (int, map[string]any) {
		t.Helper()
		if method == "PUT" && strings.HasPrefix(target, "/api/workspace/file") {
			var input map[string]any
			if json.Unmarshal(body, &input) == nil {
				name, _ := input["path"].(string)
				rw := httptest.NewRecorder()
				srv.Handler().ServeHTTP(rw, httptest.NewRequest("GET", "/api/workspace/file?mode=edit&path="+url.QueryEscape(name), nil))
				var loaded map[string]any
				json.Unmarshal(rw.Body.Bytes(), &loaded)
				input["expected_revision"] = loaded["revision"]
				body, _ = json.Marshal(input)
			}
		}
		req := httptest.NewRequest(method, target, bytes.NewReader(body))
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		r := httptest.NewRecorder()
		srv.Handler().ServeHTTP(r, req)
		var out map[string]any
		_ = json.Unmarshal(r.Body.Bytes(), &out)
		return r.Code, out
	}
	upload := func(query, name, content string) (int, map[string]any) {
		t.Helper()
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, _ := mw.CreateFormFile("file", name)
		fw.Write([]byte(content))
		mw.Close()
		return call("POST", "/api/workspace/upload"+query, mw.FormDataContentType(), buf.Bytes())
	}
	jsonBody := func(v any) []byte { b, _ := json.Marshal(v); return b }
	read := func(path string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	// Upload to the root, then conflict, then overwrite.
	if code, out := upload("", "a.md", "one"); code != 200 || out["path"] != "a.md" || out["overwritten"] != false || out["size"] != float64(3) {
		t.Fatal(code, out)
	}
	if code, out := upload("", "a.md", "two"); code != 409 || out["code"] != "file_exists" {
		t.Fatal(code, out)
	}
	if code, out := upload("?overwrite=1", "a.md", "two"); code != 200 || out["overwritten"] != true || read("a.md") != "two" {
		t.Fatal(code, out)
	}
	if code, out := upload("?path=dir", "../b.md", "b"); code != 200 || out["path"] != "dir/b.md" || read("dir/b.md") != "b" {
		t.Fatal(code, out)
	}
	if code, _ := upload("?path=outside", "c.md", "c"); code == 200 {
		t.Fatal("upload followed a symlink out of the workspace")
	}
	if code, out := upload("?path=a.md", "c.md", "c"); code != 400 || out["error"] != "Path is not a directory" {
		t.Fatal(code, out)
	}

	// Create, update, stat.
	if code, out := call("POST", "/api/workspace/file", "application/json", jsonBody(map[string]any{"path": "dir", "name": "new.md", "content": ""})); code != 200 || out["path"] != "dir/new.md" {
		t.Fatal(code, out)
	}
	if code, out := call("POST", "/api/workspace/file", "application/json", jsonBody(map[string]any{"path": "dir", "name": "new.md", "content": ""})); code != 409 || out["code"] != "file_exists" {
		t.Fatal(code, out)
	}
	if code, out := call("POST", "/api/workspace/file", "application/json", jsonBody(map[string]any{"path": "dir", "name": "../x.md", "content": ""})); code != 400 || out["error"] != "Invalid filename" {
		t.Fatal(code, out)
	}
	if code, out := call("PUT", "/api/workspace/file", "application/json", jsonBody(map[string]any{"path": "dir/new.md", "content": "saved"})); code != 200 || out["size"] != float64(5) || read("dir/new.md") != "saved" {
		t.Fatal(code, out)
	}
	if code, out := call("GET", "/api/workspace/stat?path=dir/new.md", "", nil); code != 200 || out["size"] != float64(5) || out["mtime"] == nil {
		t.Fatal(code, out)
	}

	// Rename and move, including their conflicts and root guards.
	if code, out := call("POST", "/api/workspace/rename", "application/json", jsonBody(map[string]any{"path": "dir/new.md", "name": "renamed.md"})); code != 200 || out["path"] != "dir/renamed.md" || out["old_path"] != "dir/new.md" {
		t.Fatal(code, out)
	}
	if code, out := call("POST", "/api/workspace/rename", "application/json", jsonBody(map[string]any{"path": "dir/renamed.md", "name": "b.md"})); code != 409 {
		t.Fatal(code, out)
	}
	if code, out := call("POST", "/api/workspace/rename", "application/json", jsonBody(map[string]any{"path": "", "name": "x"})); code != 400 || out["error"] != "Cannot rename workspace root" {
		t.Fatal(code, out)
	}
	if code, out := call("POST", "/api/workspace/move", "application/json", jsonBody(map[string]any{"path": "dir/renamed.md", "target": ""})); code != 200 || out["path"] != "renamed.md" || read("renamed.md") != "saved" {
		t.Fatal(code, out)
	}
	if code, out := call("POST", "/api/workspace/move", "application/json", jsonBody(map[string]any{"path": "dir", "target": "dir"})); code != 400 || !strings.Contains(out["error"].(string), "into itself") {
		t.Fatal(code, out)
	}
	if code, out := call("POST", "/api/workspace/move", "application/json", jsonBody(map[string]any{"path": "a.md", "target": "../"})); code != 400 || out["error"] != "Invalid target" {
		t.Fatal(code, out)
	}

	// Delete files only.
	if code, out := call("DELETE", "/api/workspace/file?path=renamed.md", "", nil); code != 200 || out["deleted"] != true {
		t.Fatal(code, out)
	}
	if _, err := os.Stat(filepath.Join(root, "renamed.md")); !os.IsNotExist(err) {
		t.Fatal("file not deleted", err)
	}
	if code, out := call("DELETE", "/api/workspace/file?path=dir", "", nil); code != 400 || out["error"] != "Path is a directory" {
		t.Fatal(code, out)
	}
	if code, _ := call("DELETE", "/api/workspace/file?path=missing.md", "", nil); code != 404 {
		t.Fatal(code)
	}
	entries, _ := os.ReadDir(outside)
	if len(entries) != 0 {
		t.Fatal("wrote outside the workspace", entries)
	}
}
