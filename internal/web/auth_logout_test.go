package web

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestBrowserLogoutGuardsAndCookieClearing(t *testing.T) {
	for _, tc := range []struct {
		name, method, body, origin, query, media, peer string
		cookie, bearer, duplicate, secure              bool
		want                                           int
	}{
		{name: "success", cookie: true, want: 200},
		{name: "tls-success", cookie: true, secure: true, want: 200},
		{name: "no-cookie", want: 401},
		{name: "bearer", bearer: true, want: 403},
		{name: "cookie-and-bearer", cookie: true, bearer: true, want: 403},
		{name: "query", cookie: true, query: "?auth_token=x", want: 403},
		{name: "duplicate-cookie", cookie: true, duplicate: true, want: 401},
		{name: "foreign", cookie: true, origin: "https://evil.example", want: 403},
		{name: "missing-origin", cookie: true, origin: "none", want: 403},
		{name: "remote-http", cookie: true, peer: "192.0.2.1:2", want: 403},
		{name: "method", cookie: true, method: "GET", want: 405},
		{name: "media", cookie: true, media: "text/plain", want: 415},
		{name: "null", cookie: true, body: "null", want: 400},
		{name: "array", cookie: true, body: "[]", want: 400},
		{name: "unknown", cookie: true, body: `{"all":true}`, want: 400},
		{name: "trailing", cookie: true, body: "{} {}", want: 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := authSessionServer(t)
			owner := loginBrowser(t, s, tc.secure)
			other := loginBrowser(t, s, tc.secure)
			state := authState(t, s)
			automation, _, err := s.auth.VerifyLogin("", webTestTOTPCode(state.TOTPSecret))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			method, body, origin, media, peer := tc.method, tc.body, tc.origin, tc.media, tc.peer
			scheme := "http"
			if tc.secure {
				scheme = "https"
			}
			if method == "" {
				method = "POST"
			}
			if body == "" {
				body = "{}"
			}
			if origin == "" {
				origin = scheme + "://localhost"
			}
			if origin == "none" {
				origin = ""
			}
			if media == "" {
				media = "application/json"
			}
			if peer == "" {
				peer = "127.0.0.1:12"
			}
			r := httptest.NewRequest(method, scheme+"://localhost/api/auth/session/logout"+tc.query, bytes.NewBufferString(body))
			r.RemoteAddr = peer
			r.Header.Set("Content-Type", media)
			r.Header.Set("Origin", origin)
			if tc.secure {
				r.TLS = &tls.ConnectionState{}
			}
			if tc.cookie {
				r.AddCookie(owner)
			}
			if tc.duplicate {
				r.AddCookie(owner)
			}
			if tc.bearer {
				r.Header.Set("Authorization", "Bearer "+automation)
			}
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("status=%d body=%s", w.Code, w.Body)
			}
			if w.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("cache")
			}
			if tc.want == 200 {
				var result map[string]any
				if err = json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result) != 1 || result["ok"] != true {
					t.Fatal("unexpected success body")
				}
				cookies := w.Result().Cookies()
				if len(cookies) != 1 {
					t.Fatal("clearing cookie absent")
				}
				c := cookies[0]
				if c.Name != browserSessionCookie || c.Value != "" || c.Path != "/" || c.Domain != "" || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode || c.Secure != tc.secure || c.MaxAge != -1 {
					t.Fatalf("clearing cookie %+v", c)
				}
				if s.auth.ValidateToken(owner.Value) || !s.auth.ValidateToken(other.Value) || !s.auth.ValidateToken(automation) {
					t.Fatal("wrong token revoked")
				}
			} else {
				after, _ := os.ReadFile(path)
				if !bytes.Equal(before, after) || len(w.Result().Cookies()) != 0 {
					t.Fatal("refusal mutated authority")
				}
			}
		})
	}
}
