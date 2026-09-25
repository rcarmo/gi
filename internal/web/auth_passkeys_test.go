package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	giauth "github.com/rcarmo/gi/internal/auth"
)

// This route matrix uses seeded credential records, not a WebAuthn verifier
// bypass. UI tests separately prove real credential creation and sign-in.
func TestBrowserPasskeyManagementAuthority(t *testing.T) {
	for _, scenario := range []struct {
		name, method, path, authority string
		status                        int
	}{
		{"anonymous-list", "GET", "/api/auth/passkeys", "anonymous", 401},
		{"automation-register", "POST", "/api/auth/passkeys/register/start", "bearer", 403},
		{"automation-cookie-register", "POST", "/api/auth/passkeys/register/start", "automation-cookie", 401},
		{"expired-rename", "POST", "/api/auth/passkeys/rename", "expired", 401},
		{"revoked-rename", "POST", "/api/auth/passkeys/rename", "revoked", 401},
		{"foreign-origin-remove", "POST", "/api/auth/passkeys/remove", "foreign", 403},
		{"account-selected-list", "GET", "/api/auth/passkeys?account=another", "owner", 400},
		{"username-selected-list", "GET", "/api/auth/passkeys?username=another", "owner", 400},
		{"empty-account-list", "GET", "/api/auth/passkeys?account=", "owner", 400},
		{"duplicate-account-list", "GET", "/api/auth/passkeys?account=admin&account=another", "owner", 400},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			s := authSessionServer(t)
			owner := loginBrowser(t, s, false)
			s.auth = giauth.NewManagerWithPasskeys(s.cfg.WorkspaceRoot, giauth.PasskeyConfig{RPID: "localhost", Origins: []string{"http://localhost"}})
			state := authState(t, s)
			automation, _, err := s.auth.VerifyLogin("", webTestTOTPCode(state.TOTPSecret))
			if err != nil {
				t.Fatal(err)
			}
			state = authState(t, s)
			state.LoginPolicy = "either"
			state.Passkeys = []giauth.Passkey{{Name: "Private owner key", RPID: "localhost", CreatedAt: time.Now().UTC(), Credential: webauthn.Credential{ID: []byte("private-owner-credential"), PublicKey: []byte("private-owner-public-key")}}}
			if scenario.authority == "expired" {
				for i := range state.Sessions {
					state.Sessions[i].ExpiresAt = time.Now().Add(-time.Hour)
				}
			}
			path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
			data, err := json.Marshal(state)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			if scenario.authority == "revoked" {
				if err = s.auth.RevokeToken(owner.Value); err != nil {
					t.Fatal(err)
				}
			}
			request := func(method, target string, cookie *http.Cookie, origin, bearer string) *httptest.ResponseRecorder {
				body := map[string]string{"id": base64.RawURLEncoding.EncodeToString(state.Passkeys[0].Credential.ID)}
				if strings.Contains(target, "rename") {
					body["name"] = "Unauthorised name"
				}
				if strings.Contains(target, "register") {
					body = map[string]string{"name": "Unauthorised key"}
				}
				payload, _ := json.Marshal(body)
				r := httptest.NewRequest(method, "http://localhost"+target, bytes.NewReader(payload))
				r.RemoteAddr = "127.0.0.1:12"
				r.Header.Set("Origin", origin)
				r.Header.Set("Content-Type", "application/json")
				if cookie != nil {
					r.AddCookie(cookie)
				}
				if bearer != "" {
					r.Header.Set("Authorization", "Bearer "+bearer)
				}
				w := httptest.NewRecorder()
				s.Handler().ServeHTTP(w, r)
				return w
			}
			cookie, origin, bearer := owner, "http://localhost", ""
			switch scenario.authority {
			case "anonymous":
				cookie = nil
			case "bearer":
				cookie, bearer = nil, automation
			case "automation-cookie":
				cookie = &http.Cookie{Name: browserSessionCookie, Value: automation}
			case "foreign":
				origin = "https://evil.example"
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			denied := request(scenario.method, scenario.path, cookie, origin, bearer)
			if denied.Code != scenario.status {
				t.Fatalf("status=%d want=%d body=%s", denied.Code, scenario.status, denied.Body)
			}
			if denied.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("cache policy")
			}
			if len(denied.Result().Cookies()) != 0 {
				t.Fatal("refusal changed cookies")
			}
			var result map[string]any
			if err = json.Unmarshal(denied.Body.Bytes(), &result); err != nil || len(result) != 1 || result["error"] == nil {
				t.Fatalf("unexpected refusal: %s", denied.Body)
			}
			for _, value := range []string{state.Passkeys[0].Name, base64.RawURLEncoding.EncodeToString(state.Passkeys[0].Credential.ID), base64.StdEncoding.EncodeToString(state.Passkeys[0].Credential.PublicKey), state.TOTPSecret, owner.Value, automation} {
				if value == "" || strings.Contains(denied.Body.String(), value) {
					t.Fatal("missing canary or inventory/secret disclosed")
				}
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("refused request changed authentication state")
			}
			// Positive control: the same server exposes this actual inventory to a
			// fresh owner, so refusal is not caused by an unconfigured/empty store.
			if scenario.authority == "expired" || scenario.authority == "revoked" {
				owner = loginBrowser(t, s, false)
			}
			accepted := request("GET", "/api/auth/passkeys", owner, "http://localhost", "")
			if accepted.Code != 200 || !strings.Contains(accepted.Body.String(), state.Passkeys[0].Name) {
				t.Fatalf("positive control failed: %d %s", accepted.Code, accepted.Body)
			}
		})
	}
}

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
