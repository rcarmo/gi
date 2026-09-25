package web

import (
	giauth "github.com/rcarmo/gi/internal/auth"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBrowserPasskeyRoutesDisabledAndGuarded(t *testing.T) {
	s := authSessionServer(t)
	cookie := loginBrowser(t, s, false)
	for _, path := range []string{"/api/auth/passkeys", "/api/auth/passkeys/register/start", "/api/auth/passkeys/login/start", "/api/auth/passkeys/reauth/start", "/api/auth/passkeys/remove", "/api/auth/passkeys/rename"} {
		method := "POST"
		if path == "/api/auth/passkeys" {
			method = "GET"
		}
		r := httptest.NewRequest(method, "http://localhost"+path, strings.NewReader(`{}`))
		r.RemoteAddr = "127.0.0.1:12"
		r.AddCookie(cookie)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", "http://localhost")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatalf("unconfigured %s %d %s", path, w.Code, w.Body)
		}
	}
	s.auth = giauth.NewManagerWithPasskeys(s.cfg.WorkspaceRoot, giauth.PasskeyConfig{RPID: "localhost", Origins: []string{"http://localhost"}})
	state := authState(t, s)
	bearer, _, err := s.auth.VerifyLogin("", webTestTOTPCode(state.TOTPSecret))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, origin, body, content, authorization, query, host, peer string
		cookie                                                        *http.Cookie
		want                                                          int
	}{
		{name: "valid", origin: "http://localhost", body: `{"name":"Laptop"}`, cookie: cookie, want: 200},
		{name: "no-cookie", origin: "http://localhost", body: `{"name":"Laptop"}`, want: 401},
		{name: "bearer-cookie", origin: "http://localhost", body: `{"name":"Laptop"}`, cookie: &http.Cookie{Name: browserSessionCookie, Value: bearer}, want: 401},
		{name: "explicit-bearer", origin: "http://localhost", body: `{"name":"Laptop"}`, cookie: cookie, authorization: bearer, want: 403},
		{name: "query", origin: "http://localhost", cookie: cookie, query: "?auth_token=", want: 403},
		{name: "missing-origin", cookie: cookie, body: `{"name":"Laptop"}`, want: 403},
		{name: "wrong-origin", origin: "https://evil.example", cookie: cookie, want: 403},
		{name: "wrong-rp-host", origin: "http://other.localhost", host: "other.localhost", cookie: cookie, want: 403},
		{name: "remote-http", origin: "http://localhost", peer: "192.0.2.1:2", cookie: cookie, want: 403},
		{name: "unknown-fields", origin: "http://localhost", body: `{"name":"Laptop","account":"another"}`, cookie: cookie, want: 400},
		{name: "bad-name", origin: "http://localhost", body: `{"name":""}`, cookie: cookie, want: 400},
		{name: "wrong-media", origin: "http://localhost", content: "text/plain", body: `{}`, cookie: cookie, want: 415},
		{name: "oversize", origin: "http://localhost", body: `{"name":"` + strings.Repeat("a", 128*1024) + `"}`, cookie: cookie, want: 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := authState(t, s)
			r := httptest.NewRequest("POST", "http://localhost/api/auth/passkeys/register/start"+tc.query, strings.NewReader(tc.body))
			r.RemoteAddr = "127.0.0.1:12"
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Origin", tc.origin)
			if tc.cookie != nil {
				r.AddCookie(tc.cookie)
			}
			if tc.content != "" {
				r.Header.Set("Content-Type", tc.content)
			}
			if tc.authorization != "" {
				r.Header.Set("Authorization", "Bearer "+tc.authorization)
			}
			if tc.host != "" {
				r.Host = tc.host
			}
			if tc.peer != "" {
				r.RemoteAddr = tc.peer
			}
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("%d want%d %s", w.Code, tc.want, w.Body)
			}
			if w.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("cache policy")
			}
			after := authState(t, s)
			if tc.want != 200 && len(before.Ceremonies) != len(after.Ceremonies) {
				t.Fatal("failed begin modified ceremonies")
			}
		})
	}
}
