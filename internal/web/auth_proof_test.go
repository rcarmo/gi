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
	"time"

	giauth "github.com/rcarmo/gi/internal/auth"
)

func authState(t *testing.T, s *Server) giauth.State {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	var state giauth.State
	if err = json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestBrowserProofRoutesReauthAndIsolation(t *testing.T) {
	s := authSessionServer(t)
	a, b := loginBrowser(t, s, false), loginBrowser(t, s, false)
	state := authState(t, s)
	for i := range state.Sessions {
		state.Sessions[i].CreatedAt = time.Now().Add(-time.Hour)
		state.Sessions[i].AuthenticatedAt = time.Now().Add(-6 * time.Minute)
	}
	// Isolated fixture ageing only; production routes never accept client time.
	data, _ := json.Marshal(state)
	path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	before := string(data)
	for _, cookie := range []*http.Cookie{a, b} {
		w := sessionRequest(s, "GET", "/api/auth/session/proof", "", cookie, false)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"reauth_required":true`) {
			t.Fatalf("proof %d %s", w.Code, w.Body)
		}
	}
	after, _ := os.ReadFile(path)
	if string(after) != before {
		t.Fatal("reads mutated auth")
	}
	w := sessionRequest(s, "POST", "/api/auth/session/reauth/totp", `{"code":"`+webTestTOTPCode(state.TOTPSecret)+`"}`, a, false)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"reauth_required":false`) || len(w.Result().Cookies()) != 0 {
		t.Fatalf("reauth %d %s", w.Code, w.Body)
	}
	for _, secret := range []string{a.Value, b.Value, state.TOTPSecret, "token_hash", "auth_factor"} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatal("proof leaked auth material")
		}
	}
	s = New(nil, nil, s.cfg)
	for i, cookie := range []*http.Cookie{a, b} {
		w = sessionRequest(s, "GET", "/api/auth/session/proof", "", cookie, false)
		want := `"reauth_required":false`
		if i == 1 {
			want = `"reauth_required":true`
		}
		if w.Code != 200 || !strings.Contains(w.Body.String(), want) {
			t.Fatalf("reopen %d %s", w.Code, w.Body)
		}
	}
	persisted := authState(t, s)
	for i := range state.Sessions {
		if state.Sessions[i].TokenHash != persisted.Sessions[i].TokenHash || !state.Sessions[i].ExpiresAt.Equal(persisted.Sessions[i].ExpiresAt) {
			t.Fatal("reauth changed token/expiry")
		}
	}
	if err := s.auth.RevokeToken(a.Value); err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []struct{ method, path string }{{"GET", "/api/auth/session/proof"}, {"POST", "/api/auth/session/reauth/totp"}} {
		w = sessionRequest(s, endpoint.method, endpoint.path, `{"code":"`+webTestTOTPCode(state.TOTPSecret)+`"}`, a, false)
		if w.Code != 401 {
			t.Fatalf("revoked %d %s", w.Code, w.Body)
		}
	}
}

func TestBrowserProofRejectsBearerLegacyAndUnsafeRequests(t *testing.T) {
	s := authSessionServer(t)
	owner := loginBrowser(t, s, false)
	state := authState(t, s)
	w := sessionRequest(s, "POST", "/api/auth/totp/verify", `{"code":"`+webTestTOTPCode(state.TOTPSecret)+`"}`, nil, false)
	var bearer struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &bearer); err != nil || w.Code != 200 || bearer.Token == "" {
		t.Fatalf("bearer %d %s", w.Code, w.Body)
	}
	legacy := &http.Cookie{Name: browserSessionCookie, Value: bearer.Token}
	// Ordinary API use remains backward compatible, but owner proof fails closed.
	if !s.auth.ValidateToken(legacy.Value) {
		t.Fatal("ordinary bearer invalid")
	}
	for _, endpoint := range []struct{ method, path string }{{"GET", "/api/auth/session/proof"}, {"POST", "/api/auth/session/reauth/totp"}} {
		for _, tc := range []struct {
			name                                    string
			cookie                                  *http.Cookie
			header, query, origin, host, peer, site string
			tls, duplicate                          bool
			status                                  int
		}{
			{name: "owner", cookie: owner, status: 200},
			{name: "remote-tls", cookie: owner, tls: true, host: "gi.example", peer: "192.0.2.1:45", origin: "https://gi.example", status: 200},
			{name: "missing", status: 401},
			{name: "legacy-cookie", cookie: legacy, status: 401},
			{name: "explicit-bearer", cookie: owner, header: bearer.Token, status: 403},
			{name: "explicit-invalid", cookie: owner, header: "invalid", status: 403},
			{name: "query", cookie: owner, query: "?auth_token=" + bearer.Token, status: 403},
			{name: "empty-query", cookie: owner, query: "?auth_token=", status: 403},
			{name: "other-origin", cookie: owner, origin: "https://evil.example", status: 403},
			{name: "null-origin", cookie: owner, origin: "null", status: 403},
			{name: "cross-site", cookie: owner, site: "cross-site", status: 403},
			{name: "remote-http", cookie: owner, peer: "192.0.2.1:45", status: 403},
			{name: "host-rebinding", cookie: owner, host: "evil.example", status: 403},
			{name: "duplicate-cookie", cookie: owner, duplicate: true, status: 401},
		} {
			t.Run(endpoint.method+"/"+tc.name, func(t *testing.T) {
				path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
				before, _ := os.ReadFile(path)
				r := httptest.NewRequest(endpoint.method, "http://localhost"+endpoint.path+tc.query, strings.NewReader(`{"code":"`+webTestTOTPCode(state.TOTPSecret)+`"}`))
				r.RemoteAddr = "127.0.0.1:123"
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Origin", tc.origin)
				r.Header.Set("Sec-Fetch-Site", tc.site)
				if tc.cookie != nil {
					r.AddCookie(tc.cookie)
				}
				if tc.duplicate {
					r.AddCookie(tc.cookie)
				}
				if tc.header != "" {
					r.Header.Set("Authorization", "Bearer "+tc.header)
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
				r.Header.Set("X-Forwarded-Proto", "https")
				r.Header.Set("X-Forwarded-For", "127.0.0.1")
				w := httptest.NewRecorder()
				s.Handler().ServeHTTP(w, r)
				if w.Code != tc.status || w.Header().Get("Cache-Control") != "private, no-store" || len(w.Result().Cookies()) != 0 {
					t.Fatalf("status=%d want=%d %s", w.Code, tc.status, w.Body)
				}
				if tc.status != 200 || endpoint.method == "GET" {
					after, _ := os.ReadFile(path)
					if string(before) != string(after) {
						t.Fatal("rejected/read request mutated state")
					}
				}
			})
		}
	}
}

func TestBrowserReauthRejectsMalformedAndBrokenState(t *testing.T) {
	s := authSessionServer(t)
	cookie := loginBrowser(t, s, false)
	for _, tc := range []struct {
		method, body, content string
		status                int
	}{
		{"GET", "", "application/json", 405}, {"POST", "code=123456", "application/x-www-form-urlencoded", 415},
		{"POST", `{"code":"bad"}`, "application/json", 401}, {"POST", `{"code":"123456","authenticated_at":"tomorrow"}`, "application/json", 400},
		{"POST", `{"code":"123456"}{}`, "application/json", 400}, {"POST", `{"code":"` + strings.Repeat("x", 1100) + `"}`, "application/json", 400},
	} {
		path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
		before, _ := os.ReadFile(path)
		r := httptest.NewRequest(tc.method, "http://localhost/api/auth/session/reauth/totp", strings.NewReader(tc.body))
		r.RemoteAddr = "127.0.0.1:45"
		r.AddCookie(cookie)
		r.Header.Set("Content-Type", tc.content)
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("status%d want%d: %s", w.Code, tc.status, w.Body)
		}
		after, _ := os.ReadFile(path)
		if string(before) != string(after) {
			t.Fatal("bad input mutated state")
		}
	}
	if err := os.WriteFile(filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json"), []byte("{broken-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []struct{ method, path string }{{"GET", "/api/auth/session/proof"}, {"POST", "/api/auth/session/reauth/totp"}} {
		w := sessionRequest(s, endpoint.method, endpoint.path, `{"code":"123456"}`, cookie, false)
		if w.Code != 500 || strings.Contains(w.Body.String(), "broken-secret") || len(w.Result().Cookies()) != 0 {
			t.Fatalf("broken %d %s", w.Code, w.Body)
		}
	}
}
