package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestMediaLookupNativeBytesMetadataAndAuth(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "media.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	session, err := s.CreateSession(t.Context(), "media-session", "Media", nil)
	if err != nil {
		t.Fatal(err)
	}
	raw := bytes.Repeat([]byte("stored bytes\n"), 100)
	media, err := s.CreateMedia(t.Context(), session.ID, "image α.png", "image/png", raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := New(s, turn.New(s), config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	path := fmt.Sprintf("/api/media/%d", media.ID)
	call := func(method, path string) *httptest.ResponseRecorder {
		t.Helper()
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, httptest.NewRequest(method, path, nil))
		return res
	}
	info := call("GET", path)
	if info.Code != 200 {
		t.Fatal(info.Code, info.Body.String())
	}
	var decoded store.Media
	if err := json.Unmarshal(info.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID != media.ID || decoded.Filename != media.Filename || decoded.SessionID != session.ID {
		t.Fatalf("metadata mismatch: %+v", decoded)
	}
	for _, suffix := range []string{"", "/raw"} {
		head := call("HEAD", path+suffix)
		if head.Code != 200 || head.Body.Len() != 0 {
			t.Fatalf("HEAD: %d %q", head.Code, head.Body.String())
		}
	}
	res := call("GET", path+"/raw")
	if res.Code != 200 || !bytes.Equal(res.Body.Bytes(), raw) {
		t.Fatal("raw bytes mismatch", res.Code)
	}
	for key, want := range map[string]string{"Content-Type": "image/png", "X-Content-Type-Options": "nosniff", "Cache-Control": "private, no-store", "Content-Security-Policy": "sandbox; default-src 'none'"} {
		if res.Header().Get(key) != want {
			t.Fatalf("%s: %q", key, res.Header().Get(key))
		}
	}
	if !strings.HasPrefix(res.Header().Get("Content-Disposition"), "inline;") {
		t.Fatal(res.Header())
	}
	req := httptest.NewRequest("GET", path+"/raw", nil)
	req.Header.Set("Range", "bytes=1-4")
	res = httptest.NewRecorder()
	srv.Handler().ServeHTTP(res, req)
	if res.Code != 206 || !bytes.Equal(res.Body.Bytes(), raw[1:5]) {
		t.Fatal("range response", res.Code)
	}
	for _, bad := range []string{"/api/media/", "/api/media/-1", "/api/media/nope", "/api/media/999999", path + "/unknown", path + "/raw/extra"} {
		if got := call("GET", bad).Code; got != 404 {
			t.Fatalf("%s: %d", bad, got)
		}
	}
	if got := call("POST", path).Code; got != 405 {
		t.Fatal(got)
	}
	for _, kind := range []string{"text/html", "image/svg+xml", "application/octet-stream"} {
		file, err := s.CreateMedia(t.Context(), session.ID, "active.html", kind, []byte("<script>alert(1)</script>"), nil)
		if err != nil {
			t.Fatal(err)
		}
		res := call("GET", fmt.Sprintf("/api/media/%d/raw", file.ID))
		if res.Code != 200 || !strings.HasPrefix(res.Header().Get("Content-Disposition"), "attachment;") || !strings.Contains(res.Header().Get("Content-Security-Policy"), "sandbox") {
			t.Fatal(kind, res.Code, res.Header())
		}
	}
	pending, err := srv.auth.StartEnrollment("rui")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.auth.VerifyEnrollment("rui", webTestTOTPCode(pending.Secret)); err != nil {
		t.Fatal(err)
	}
	token, _, err := srv.auth.VerifyLogin("rui", webTestTOTPCode(pending.Secret))
	if err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"", "/raw"} {
		for _, method := range []string{"GET", "HEAD"} {
			if got := call(method, path+suffix).Code; got != 401 {
				t.Fatalf("unguarded %s %s: %d", method, suffix, got)
			}
			req := httptest.NewRequest(method, path+suffix, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()
			srv.Handler().ServeHTTP(res, req)
			if res.Code != 200 {
				t.Fatalf("authorised %s %s: %d", method, suffix, res.Code)
			}
		}
	}
}
