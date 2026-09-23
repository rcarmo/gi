package web

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestProviderSettingsTransportAuthSecretAndPersistence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".pi", "agent")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{"other":{"type":"oauth","access":"fixture-other"}}`), 0600)
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	srv := New(st, turn.New(st), config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	call := func(method, body, remote, origin, ct string, secure bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/settings/providers", bytes.NewBufferString(body))
		req.RemoteAddr = remote
		req.Host = "127.0.0.1:8090"
		req.Header.Set("Content-Type", ct)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if secure {
			req.TLS = &tls.ConnectionState{}
		}
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, req)
		return res
	}
	get := call("GET", "", "127.0.0.1:1", "", "", false)
	if get.Code != 200 || get.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal(get.Code)
	}
	var meta struct {
		Revision string `json:"revision"`
		CanWrite bool   `json:"can_write"`
	}
	json.Unmarshal(get.Body.Bytes(), &meta)
	if !meta.CanWrite {
		t.Fatal("loopback blocked")
	}
	payload := map[string]any{"provider": "openai", "revision": meta.Revision, "key": "fixture-sensitive-key"}
	body, _ := json.Marshal(payload)
	for _, tc := range []struct {
		remote, origin, ct string
		code               int
	}{{"192.0.2.1:2", "", "application/json", 403}, {"127.0.0.1:2", "https://evil.invalid", "application/json", 403}, {"127.0.0.1:2", "", "text/plain", 415}} {
		res := call("PATCH", string(body), tc.remote, tc.origin, tc.ct, false)
		if res.Code != tc.code {
			t.Fatal(res.Code)
		}
		if bytes.Contains(res.Body.Bytes(), []byte("fixture-sensitive-key")) {
			t.Fatal("secret leak")
		}
	}
	saved := call("PATCH", string(body), "192.0.2.1:2", "", "application/json", true)
	if saved.Code != 200 {
		t.Fatalf("save %d %s", saved.Code, saved.Body.String())
	}
	if bytes.Contains(saved.Body.Bytes(), []byte("fixture-sensitive-key")) || bytes.Contains(saved.Body.Bytes(), []byte("fixture-other")) {
		t.Fatal("response leak")
	}
	if call("PATCH", string(body), "127.0.0.1:2", "", "application/json", false).Code != 409 {
		t.Fatal("stale accepted")
	}
	current, err := inference.ReadProviderSettings()
	if err != nil {
		t.Fatal(err)
	}
	payload["revision"] = current.Revision
	delete(payload, "key")
	body, _ = json.Marshal(payload)
	removed := call("DELETE", string(body), "127.0.0.1:2", "", "application/json", false)
	if removed.Code != 200 {
		t.Fatal(removed.Code)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "auth.json"))
	if bytes.Contains(raw, []byte("fixture-sensitive-key")) || !bytes.Contains(raw, []byte("fixture-other")) {
		t.Fatal("wrong delete scope")
	}
	for _, bad := range []string{`{}`, `{"provider":"openai","key":"fixture","surprise":true}`, `{} {}`} {
		res := call("PATCH", bad, "127.0.0.1:2", "", "application/json", false)
		if res.Code != 400 {
			t.Fatal(res.Code)
		}
	}
	if call("POST", "", "127.0.0.1:2", "", "", false).Code != 405 {
		t.Fatal("method accepted")
	}
}

func TestProviderWriteTransportDoesNotTrustProxyHeaders(t *testing.T) {
	for _, tc := range []struct {
		host, remote string
		secure, want bool
	}{
		{"127.0.0.1:8090", "127.0.0.1:1234", false, true}, {"localhost:8090", "[::1]:1234", false, true}, {"[::1]:8090", "[::1]:1234", false, true},
		{"public.invalid", "127.0.0.1:1234", false, false}, {"127.0.0.1:8090", "192.0.2.1:1234", false, false}, {"public.invalid", "192.0.2.1:1234", true, true},
	} {
		r := httptest.NewRequest("PATCH", "/api/settings/providers", nil)
		r.Host = tc.host
		r.RemoteAddr = tc.remote
		r.Header.Set("X-Forwarded-Proto", "https")
		r.Header.Set("X-Forwarded-For", "127.0.0.1")
		if tc.secure {
			r.TLS = &tls.ConnectionState{}
		}
		if got := providerWriteTransport(r); got != tc.want {
			t.Fatalf("%+v => %v", tc, got)
		}
	}
}
