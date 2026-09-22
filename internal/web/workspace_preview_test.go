package web

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestWorkspacePreviewBoundedKindsAndConfinement(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: root})
	write := func(name string, data []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var pngBytes bytes.Buffer
	if err := png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 2, 3))); err != nil {
		t.Fatal(err)
	}
	write("picture.png", pngBytes.Bytes())
	write("readme.md", []byte("# Hello\n\n**Markdown**"))
	write("source.html", []byte("<script>alert(1)</script>"))
	write("binary.dat", []byte{0, 255, 1})
	write("empty", nil)
	write("large.txt", []byte(strings.Repeat("界", 100000)))
	outside := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(outside, []byte("outside secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("readme.md", filepath.Join(root, "inside")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "folder"), 0700); err != nil {
		t.Fatal(err)
	}
	call := func(method, endpoint, path, extra string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRecorder()
		srv.Handler().ServeHTTP(r, httptest.NewRequest(method, endpoint+"?path="+url.QueryEscape(path)+extra, nil))
		return r
	}
	preview := func(path, extra string) map[string]any {
		t.Helper()
		res := call("GET", "/api/workspace/file", path, extra)
		if res.Code != 200 {
			t.Fatal(path, res.Code, res.Body.String())
		}
		var p map[string]any
		if err := json.Unmarshal(res.Body.Bytes(), &p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	for path, kind := range map[string]string{"picture.png": "image", "readme.md": "text", "source.html": "text", "binary.dat": "binary", "empty": "text", "inside": "text"} {
		p := preview(path, "")
		if p["kind"] != kind || p["path"] != path || p["mtime"] == nil || p["size"] == nil {
			t.Fatalf("%s: %#v", path, p)
		}
	}
	if p := preview("readme.md", ""); p["content_type"] != "text/markdown" || p["text"] != p["content"] {
		t.Fatal(p)
	}
	for _, limit := range []string{"1", "2", "4", "511", "512", "513", "20000", "262144"} {
		p := preview("large.txt", "&max_bytes="+limit)
		text := p["text"].(string)
		if !utf8.ValidString(text) || p["truncated"] != true || len(text) > workspacePreviewLimit {
			t.Fatal(limit, p)
		}
	}
	for _, limit := range []string{"0", "-1", "262145", "nope"} {
		if r := call("GET", "/api/workspace/file", "readme.md", "&max_bytes="+limit); r.Code != 400 {
			t.Fatal(limit, r.Code)
		}
	}
	for _, path := range []string{"", outside, "../secret", "outside", "folder"} {
		for _, endpoint := range []string{"/api/workspace/file", "/api/workspace/raw"} {
			res := call("GET", endpoint, path, "")
			if res.Code == 200 || strings.Contains(res.Body.String(), "outside secret") {
				t.Fatal(path, endpoint, res.Code)
			}
		}
	}
	raw := call("GET", "/api/workspace/raw", "picture.png", "")
	if raw.Code != 200 || !bytes.Equal(raw.Body.Bytes(), pngBytes.Bytes()) || !strings.HasPrefix(raw.Header().Get("Content-Disposition"), "inline;") {
		t.Fatal(raw.Code, raw.Header())
	}
	for _, path := range []string{"source.html", "binary.dat"} {
		r := call("GET", "/api/workspace/raw", path, "")
		if r.Code != 200 || !strings.HasPrefix(r.Header().Get("Content-Disposition"), "attachment;") || r.Header().Get("Content-Security-Policy") != "sandbox; default-src 'none'" || r.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatal(r.Code, r.Header())
		}
	}
	if r := call("GET", "/api/workspace/raw", "picture.png", "&download=1"); !strings.HasPrefix(r.Header().Get("Content-Disposition"), "attachment;") {
		t.Fatal(r.Header())
	}
	if r := call("HEAD", "/api/workspace/raw", "picture.png", ""); r.Code != 200 || r.Body.Len() != 0 {
		t.Fatal(r.Code)
	}
	req := httptest.NewRequest("GET", "/api/workspace/raw?path=picture.png", nil)
	req.Header.Set("Range", "bytes=1-4")
	r := httptest.NewRecorder()
	srv.Handler().ServeHTTP(r, req)
	if r.Code != 206 || !bytes.Equal(r.Body.Bytes(), pngBytes.Bytes()[1:5]) {
		t.Fatal(r.Code)
	}
	for _, endpoint := range []string{"/api/workspace/file", "/api/workspace/raw"} {
		if r := call("POST", endpoint, "readme.md", ""); r.Code != 405 {
			t.Fatal(r.Code)
		}
	}
	pending, err := srv.auth.StartEnrollment("rui")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = srv.auth.VerifyEnrollment("rui", webTestTOTPCode(pending.Secret)); err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []string{"/api/workspace/file", "/api/workspace/raw"} {
		if r := call("GET", endpoint, "readme.md", ""); r.Code != 401 {
			t.Fatal("unguarded", r.Code)
		}
	}
}
