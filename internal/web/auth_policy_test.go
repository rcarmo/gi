package web

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrowserLoginPolicyGuardsAndSafeDefaults(t *testing.T) {
	s := authSessionServer(t)
	cookie := loginBrowser(t, s, false)
	state := authState(t, s)
	bearer, _, err := s.auth.VerifyLogin("", webTestTOTPCode(state.TOTPSecret))
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"GET", "POST"} {
		for _, tc := range []struct {
			name, origin, auth, query, host, peer, body, content string
			cookie                                               *http.Cookie
			tls                                                  bool
			want                                                 int
		}{
			{name: "owner", origin: "http://localhost", cookie: cookie, want: 200},
			{name: "tls", origin: "https://gi.example", host: "gi.example", peer: "192.0.2.1:3", cookie: cookie, tls: true, want: 200},
			{name: "missing", origin: "http://localhost", want: 401},
			{name: "bearer-cookie", origin: "http://localhost", cookie: &http.Cookie{Name: browserSessionCookie, Value: bearer}, want: 401},
			{name: "explicit", origin: "http://localhost", cookie: cookie, auth: bearer, want: 403},
			{name: "query", origin: "http://localhost", cookie: cookie, query: "?auth_token=", want: 403},
			{name: "wrong-origin", origin: "https://evil.example", cookie: cookie, want: 403},
			{name: "remote-http", origin: "http://localhost", cookie: cookie, peer: "192.0.2.1:3", want: 403},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				prior, err := s.auth.ReadLoginPolicy(cookie.Value, "http://localhost")
				if err != nil {
					t.Fatal(err)
				}
				body := `{"policy":"either","revision":"` + prior.Revision + `"}`
				r := httptest.NewRequest(method, "http://localhost/api/auth/policy"+tc.query, strings.NewReader(body))
				r.RemoteAddr = "127.0.0.1:12"
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Origin", tc.origin)
				if tc.cookie != nil {
					r.AddCookie(tc.cookie)
				}
				if tc.auth != "" {
					r.Header.Set("Authorization", "Bearer "+tc.auth)
				}
				if tc.host != "" {
					r.Host = tc.host
				}
				if tc.peer != "" {
					r.RemoteAddr = tc.peer
				}
				if tc.tls {
					r.TLS = &tls.ConnectionState{}
				}
				path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
				before, _ := os.ReadFile(path)
				w := httptest.NewRecorder()
				s.Handler().ServeHTTP(w, r)
				if w.Code != tc.want || w.Header().Get("Cache-Control") != "private, no-store" || len(w.Result().Cookies()) != 0 {
					t.Fatalf("%d want%d %s", w.Code, tc.want, w.Body)
				}
				if strings.Contains(w.Body.String(), state.TOTPSecret) || strings.Contains(w.Body.String(), "token_hash") {
					t.Fatal("secret exposed")
				}
				if method == "GET" || tc.want != 200 {
					after, _ := os.ReadFile(path)
					if string(before) != string(after) {
						t.Fatal("read/rejection wrote state")
					}
				}
			})
		}
	}
}
func TestBrowserLoginPolicyValidatesCASAndBodies(t *testing.T) {
	s := authSessionServer(t)
	cookie := loginBrowser(t, s, false)
	for _, tc := range []struct {
		body, origin, content string
		want                  int
	}{
		{`{"policy":"passkey-only","revision":"initial"}`, "http://localhost", "application/json", 409},
		{`{"policy":"off","revision":"initial"}`, "http://localhost", "application/json", 400},
		{`{"policy":"totp-only","revision":"stale"}`, "http://localhost", "application/json", 409},
		{`{"policy":"totp-only","revision":"initial","account":"another"}`, "http://localhost", "application/json", 400},
		{`{"policy":"totp-only","revision":"initial"}{}`, "http://localhost", "application/json", 400},
		{`{"policy":"totp-only","revision":"initial"}`, "", "application/json", 403},
		{`{}`, "http://localhost", "text/plain", 415},
		{`{"policy":"` + strings.Repeat("x", 1100) + `"}`, "http://localhost", "application/json", 400},
	} {
		r := httptest.NewRequest("POST", "http://localhost/api/auth/policy", strings.NewReader(tc.body))
		r.RemoteAddr = "127.0.0.1:12"
		r.AddCookie(cookie)
		r.Header.Set("Content-Type", tc.content)
		r.Header.Set("Origin", tc.origin)
		path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
		before, _ := os.ReadFile(path)
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("%d want%d %s", w.Code, tc.want, w.Body)
		}
		after, _ := os.ReadFile(path)
		if string(before) != string(after) {
			t.Fatal("rejection wrote state")
		}
	}
}
