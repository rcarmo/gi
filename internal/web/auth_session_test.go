package web

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
)

func authSessionServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Load(dir)
	s := New(nil, nil, cfg)
	pending, err := s.auth.StartEnrollment("admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.auth.VerifyEnrollment("admin", webTestTOTPCode(pending.Secret)); err != nil {
		t.Fatal(err)
	}
	return s
}

func sessionRequest(s *Server, method, path, body string, cookie *http.Cookie, tlsOn bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://localhost"+path, strings.NewReader(body))
	r.RemoteAddr = "127.0.0.1:1234"
	if tlsOn {
		r.TLS = &tls.ConnectionState{}
	}
	r.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}

func loginBrowser(t *testing.T, s *Server, tlsOn bool) *http.Cookie {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	var state struct {
		Secret string `json:"totp_secret"`
	}
	if err = json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	w := sessionRequest(s, "POST", "/api/auth/session", `{"code":"`+webTestTOTPCode(state.Secret)+`"}`, nil, tlsOn)
	if w.Code != 200 {
		t.Fatalf("login status=%d %s", w.Code, w.Body)
	}
	if strings.Contains(w.Body.String(), "token") {
		t.Fatal("browser token exposed")
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("missing cookie")
	}
	c := cookies[0]
	if c.Name != browserSessionCookie || c.Path != "/" || c.Domain != "" || !c.HttpOnly || c.Secure != tlsOn || c.SameSite != http.SameSiteStrictMode || c.MaxAge <= 0 || c.MaxAge > 43200 {
		t.Fatalf("bad cookie attributes: %v", c)
	}
	return c
}

func TestBrowserSessionTransportAndAuthority(t *testing.T) {
	for _, tlsOn := range []bool{false, true} {
		t.Run(map[bool]string{false: "loopback", true: "tls"}[tlsOn], func(t *testing.T) {
			s := authSessionServer(t)
			cookie := loginBrowser(t, s, tlsOn)
			for _, method := range []string{"GET", "POST"} {
				r := httptest.NewRequest(method, "http://localhost/private", nil)
				r.RemoteAddr = "127.0.0.1:3456"
				if tlsOn {
					r.TLS = &tls.ConnectionState{}
				}
				r.AddCookie(cookie)
				w := httptest.NewRecorder()
				s.withAuth(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })(w, r)
				if w.Code != 204 {
					t.Fatalf("authenticated %s=%d", method, w.Code)
				}
				r.Header.Set("Origin", "https://evil.example")
				w = httptest.NewRecorder()
				s.withAuth(func(w http.ResponseWriter, r *http.Request) { t.Fatal("cross-origin admitted") })(w, r)
				if w.Code != 401 {
					t.Fatalf("cross-origin status=%d", w.Code)
				}
			}
			w := sessionRequest(s, "GET", "/api/auth/status", "", cookie, tlsOn)
			if w.Code != 200 || !strings.Contains(w.Body.String(), `"authenticated":true`) || w.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatalf("status=%d %s", w.Code, w.Body)
			}
			// Explicit invalid bearer cannot fall back to a valid cookie.
			r := httptest.NewRequest("GET", "http://localhost/private", nil)
			r.RemoteAddr = "127.0.0.1:3456"
			r.AddCookie(cookie)
			r.Header.Set("Authorization", "Bearer invalid")
			if ok, err := s.authenticatedRequest(r); ok || err != nil {
				t.Fatalf("fallback=%v %v", ok, err)
			}
			// Native bearer remains usable without cookie transport assumptions.
			r = httptest.NewRequest("GET", "http://remote/private", nil)
			r.Header.Set("Authorization", "Bearer "+cookie.Value)
			if ok, err := s.authenticatedRequest(r); !ok || err != nil {
				t.Fatalf("bearer=%v %v", ok, err)
			}
			cookie.Value = "invalid"
			if w = sessionRequest(s, "GET", "/api/auth/status", "", cookie, tlsOn); strings.Contains(w.Body.String(), `"authenticated":true`) {
				t.Fatal("invalid cookie accepted")
			}
		})
	}
}

func TestBrowserLoginRejectsUnsafeRequests(t *testing.T) {
	s := authSessionServer(t)
	for _, tc := range []struct {
		name, method, body, content, host, peer, origin, site string
		status                                                int
	}{
		{name: "method", method: "GET", status: 405},
		{name: "form", method: "POST", content: "application/x-www-form-urlencoded", body: "code=123456", status: 415},
		{name: "bad-code", method: "POST", body: `{"code":"x"}`, status: 401},
		{name: "username", method: "POST", body: `{"code":"123456","username":"admin"}`, status: 400},
		{name: "extra-json", method: "POST", body: `{"code":"123456"}{}`, status: 400},
		{name: "oversize", method: "POST", body: `{"code":"` + strings.Repeat("x", 1100) + `"}`, status: 400},
		{name: "foreign-origin", method: "POST", origin: "https://evil.example", status: 403},
		{name: "null-origin", method: "POST", origin: "null", status: 403},
		{name: "same-site", method: "POST", site: "same-site", status: 403},
		{name: "cross-site", method: "POST", site: "cross-site", status: 403},
		{name: "remote-peer", method: "POST", peer: "192.0.2.4:123", status: 403},
		{name: "host-rebinding", method: "POST", host: "evil.example", status: 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://localhost/api/auth/session", strings.NewReader(tc.body))
			r.RemoteAddr = "127.0.0.1:123"
			r.Header.Set("Content-Type", "application/json")
			if tc.content != "" {
				r.Header.Set("Content-Type", tc.content)
			}
			if tc.host != "" {
				r.Host = tc.host
			}
			if tc.peer != "" {
				r.RemoteAddr = tc.peer
			}
			r.Header.Set("Origin", tc.origin)
			r.Header.Set("Sec-Fetch-Site", tc.site)
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			if w.Code != tc.status || len(w.Result().Cookies()) != 0 {
				t.Fatalf("status=%d cookies=%d", w.Code, len(w.Result().Cookies()))
			}
		})
	}
}

func TestBrowserSessionStateFailureDoesNotIssueCookie(t *testing.T) {
	s := authSessionServer(t)
	if err := os.WriteFile(filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json"), []byte(`{broken`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/auth/session", "/api/auth/status"} {
		method := "POST"
		if path == "/api/auth/status" {
			method = "GET"
		}
		w := sessionRequest(s, method, path, `{"code":"123456"}`, nil, false)
		if w.Code != 500 || len(w.Result().Cookies()) != 0 {
			t.Fatalf("%s=%d", path, w.Code)
		}
	}
}

func TestBrowserCookieOriginTransportAndRestart(t *testing.T) {
	s := authSessionServer(t)
	cookie := loginBrowser(t, s, false)
	// A new manager validates the same persisted token, not a process-local grant.
	s = New(nil, nil, s.cfg)
	for _, tc := range []struct {
		name, method, host, peer, origin, site string
		tlsOn, allowed                         bool
	}{
		{name: "plain-loopback", method: "GET", allowed: true},
		{name: "same-origin-write", method: "POST", origin: "http://localhost", site: "same-origin", allowed: true},
		{name: "same-origin-tls", method: "PATCH", host: "gi.example", peer: "192.0.2.2:42", origin: "https://gi.example", site: "same-origin", tlsOn: true, allowed: true},
		{name: "top-level-navigation", method: "GET", site: "none", allowed: true},
		{name: "tls-offload-not-trusted", method: "POST", host: "gi.example", origin: "https://gi.example"},
		{name: "plain-remote", method: "GET", host: "gi.example", peer: "192.0.2.2:42"},
		{name: "same-site-subdomain", method: "POST", site: "same-site"},
		{name: "cross-site-read", method: "GET", site: "cross-site"},
		{name: "null-origin", method: "GET", origin: "null"},
		{name: "wrong-port", method: "POST", origin: "http://localhost:8090"},
		{name: "wrong-scheme", method: "POST", origin: "https://localhost"},
		{name: "origin-userinfo", method: "POST", origin: "http://user@localhost"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://localhost/private", nil)
			r.RemoteAddr = "127.0.0.1:99"
			if tc.host != "" {
				r.Host = tc.host
			}
			if tc.peer != "" {
				r.RemoteAddr = tc.peer
			}
			if tc.tlsOn {
				r.TLS = &tls.ConnectionState{}
			}
			r.Header.Set("Origin", tc.origin)
			r.Header.Set("Sec-Fetch-Site", tc.site)
			r.Header.Set("X-Forwarded-Proto", "https")
			r.Header.Set("X-Forwarded-For", "127.0.0.1")
			r.AddCookie(cookie)
			called := false
			w := httptest.NewRecorder()
			s.withAuth(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(204) })(w, r)
			if called != tc.allowed {
				t.Fatalf("admitted=%v want=%v status=%d", called, tc.allowed, w.Code)
			}
			if w.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("private response can be cached")
			}
		})
	}
}
