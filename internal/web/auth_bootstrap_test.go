package web

import (
	"bytes"
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
	"github.com/rcarmo/gi/internal/config"
)

func setupServer(t *testing.T) *Server {
	t.Helper()
	return New(nil, nil, config.Load(t.TempDir()))
}
func setupRequest(s *Server, op, body string, cookie *http.Cookie, secure bool) *httptest.ResponseRecorder {
	scheme := "http"
	if secure {
		scheme = "https"
	}
	r := httptest.NewRequest("POST", scheme+"://localhost"+browserSetupPath+"/"+op, strings.NewReader(body))
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("Origin", scheme+"://localhost")
	r.Header.Set("Content-Type", "application/json")
	if secure {
		r.TLS = &tls.ConnectionState{}
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
func startSetup(t *testing.T, s *Server, secure bool) (giauth.PendingEnrollment, *http.Cookie) {
	t.Helper()
	w := setupRequest(s, "start", "{}", nil, secure)
	if w.Code != 200 {
		t.Fatalf("start status=%d body=%s", w.Code, w.Body)
	}
	var pending giauth.PendingEnrollment
	if err := json.Unmarshal(w.Body.Bytes(), &pending); err != nil {
		t.Fatal(err)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("no setup cookie")
	}
	c := cookies[0]
	if c.Name != browserSetupCookie || len(c.Value) != 64 || c.Path != browserSetupPath || c.Domain != "" || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode || c.Secure != secure || c.MaxAge != 600 {
		t.Fatal("setup cookie scope")
	}
	if w.Header().Get("Cache-Control") != "private, no-store" || pending.Secret == "" || pending.Username != "admin" || !strings.HasPrefix(pending.URL, "otpauth://totp/") {
		t.Fatal("setup response contract")
	}
	if strings.Contains(w.Body.String(), c.Value) {
		t.Fatal("binding disclosed in JSON")
	}
	if _, err := os.Stat(filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")); !os.IsNotExist(err) {
		t.Fatal("start persisted secret")
	}
	return pending, c
}
func TestBrowserSetupHTTPCommitCookieAndReplay(t *testing.T) {
	for _, secure := range []bool{false, true} {
		t.Run(map[bool]string{false: "http-loopback", true: "tls-loopback"}[secure], func(t *testing.T) {
			s := setupServer(t)
			p, c := startSetup(t, s, secure)
			body := `{"code":"` + webTestTOTPCode(p.Secret) + `"}`
			w := setupRequest(s, "finish", body, c, secure)
			if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"ok":true}` {
				t.Fatalf("finish status=%d body=%s", w.Code, w.Body)
			}
			var owner *http.Cookie
			for _, cookie := range w.Result().Cookies() {
				switch cookie.Name {
				case browserSetupCookie:
					if cookie.MaxAge != -1 || cookie.Value != "" || cookie.Path != browserSetupPath || !cookie.HttpOnly || cookie.Secure != secure || cookie.SameSite != http.SameSiteStrictMode {
						t.Fatal("setup cookie not cleared")
					}
				case browserSessionCookie:
					owner = cookie
				default:
					t.Fatal("unexpected cookie")
				}
			}
			if owner == nil || len(w.Result().Cookies()) != 2 || owner.Path != "/" || owner.Domain != "" || !owner.HttpOnly || owner.Secure != secure || owner.SameSite != http.SameSiteStrictMode || owner.MaxAge <= 0 || owner.MaxAge > 43200 {
				t.Fatal("owner cookie contract")
			}
			proof, err := s.auth.BrowserSessionProof(owner.Value)
			if err != nil || proof.ReauthRequired {
				t.Fatal("atomic owner authority absent")
			}
			path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, op := range []string{"finish", "start"} {
				payload := "{}"
				want := 409
				if op == "finish" {
					payload = body
					want = 400
				}
				denied := setupRequest(s, op, payload, c, secure)
				if denied.Code != want {
					t.Fatalf("replay %s: %d", op, denied.Code)
				}
				for _, cookie := range denied.Result().Cookies() {
					if cookie.Name == browserSessionCookie {
						t.Fatal("replay issued owner cookie")
					}
				}
				after, _ := os.ReadFile(path)
				if !bytes.Equal(before, after) {
					t.Fatal("replay changed state")
				}
			}
		})
	}
}

func TestBrowserSetupHTTPAdmissionDoesNotConsumePending(t *testing.T) {
	for _, tc := range []struct {
		name, method, body, origin, host, peer, query, media, site, cookie string
		secure, authorization                                              bool
		want                                                               int
	}{
		{name: "method", method: "GET", want: 405},
		{name: "remote-peer", peer: "192.0.2.1:2", want: 403},
		{name: "remote-tls", peer: "192.0.2.1:2", secure: true, want: 403},
		{name: "foreign-host", host: "evil.example", want: 403},
		{name: "foreign-host-tls", host: "evil.example", secure: true, want: 403},
		{name: "missing-origin", origin: "none", want: 403},
		{name: "foreign-origin", origin: "https://evil.example", want: 403},
		{name: "cross-site", site: "cross-site", want: 403},
		{name: "authorization-even-empty", authorization: true, want: 403},
		{name: "query", query: "?account=admin", want: 403},
		{name: "no-cookie", cookie: "absent", want: 401},
		{name: "duplicate-cookie", cookie: "duplicate", want: 401},
		{name: "short-cookie", cookie: "short", want: 401},
		{name: "media", media: "text/plain", want: 415},
		{name: "null", body: "null", want: 400},
		{name: "array", body: "[]", want: 400},
		{name: "unknown", body: `{"code":"123456","username":"other"}`, want: 400},
		{name: "trailing", body: `{"code":"123456"} {}`, want: 400},
		{name: "numeric", body: `{"code":123456}`, want: 400},
		{name: "short-code", body: `{"code":"12345"}`, want: 400},
		{name: "letters", body: `{"code":"abcdef"}`, want: 400},
		{name: "oversize", body: `{"code":"` + strings.Repeat("1", 1100) + `"}`, want: 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := setupServer(t)
			p, c := startSetup(t, s, tc.secure)
			scheme := "http"
			if tc.secure {
				scheme = "https"
			}
			host := tc.host
			if host == "" {
				host = "localhost"
			}
			method := tc.method
			if method == "" {
				method = "POST"
			}
			body := tc.body
			if body == "" {
				body = `{"code":"` + webTestTOTPCode(p.Secret) + `"}`
			}
			origin := tc.origin
			if origin == "" {
				origin = scheme + "://" + host
			}
			if origin == "none" {
				origin = ""
			}
			peer := tc.peer
			if peer == "" {
				peer = "127.0.0.1:1234"
			}
			media := tc.media
			if media == "" {
				media = "application/json"
			}
			r := httptest.NewRequest(method, scheme+"://"+host+browserSetupPath+"/finish"+tc.query, strings.NewReader(body))
			r.RemoteAddr = peer
			r.Header.Set("Origin", origin)
			r.Header.Set("Content-Type", media)
			r.Header.Set("Sec-Fetch-Site", tc.site)
			r.Header.Set("X-Forwarded-For", "127.0.0.1")
			r.Header.Set("X-Forwarded-Host", "localhost")
			if tc.secure {
				r.TLS = &tls.ConnectionState{}
			}
			if tc.authorization {
				r.Header["Authorization"] = []string{""}
			}
			switch tc.cookie {
			case "absent":
			case "duplicate":
				r.AddCookie(c)
				r.AddCookie(c)
			case "short":
				r.AddCookie(&http.Cookie{Name: browserSetupCookie, Value: "short"})
			default:
				r.AddCookie(c)
			}
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tc.want, w.Body)
			}
			if len(w.Result().Cookies()) != 0 || w.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("admission changed cookies/cache")
			}
			if strings.Contains(w.Body.String(), p.Secret) || strings.Contains(w.Body.String(), c.Value) {
				t.Fatal("error leaked setup")
			}
			if _, err := os.Stat(filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")); !os.IsNotExist(err) {
				t.Fatal("refusal created owner")
			}
			// A malformed or unauthorized request must not consume a valid pending setup.
			good := setupRequest(s, "finish", `{"code":"`+webTestTOTPCode(p.Secret)+`"}`, c, tc.secure)
			if good.Code != 200 {
				t.Fatalf("pending consumed by admission refusal: %d %s", good.Code, good.Body)
			}
		})
	}
}

func TestBrowserSetupHTTPInvalidCodeCancelAndStaleFinish(t *testing.T) {
	for _, kind := range []string{"wrong-code", "cancel", "restart", "other-browser", "legacy-winner", "write-failure"} {
		t.Run(kind, func(t *testing.T) {
			s := setupServer(t)
			p, c := startSetup(t, s, false)
			var before []byte
			path := filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.json")
			switch kind {
			case "wrong-code":
				wrong := ""
				for _, candidate := range []string{"000000", "111111", "222222", "333333"} {
					if !giauth.VerifyTOTP(p.Secret, candidate, time.Now(), 1) {
						wrong = candidate
						break
					}
				}
				w := setupRequest(s, "finish", `{"code":"`+wrong+`"}`, c, false)
				if w.Code != 400 || len(w.Result().Cookies()) != 1 || w.Result().Cookies()[0].MaxAge != -1 {
					t.Fatal("invalid attempt not consumed")
				}
			case "cancel":
				w := setupRequest(s, "cancel", "{}", c, false)
				if w.Code != 200 || len(w.Result().Cookies()) != 1 || w.Result().Cookies()[0].MaxAge != -1 {
					t.Fatal("cancel failed")
				}
			case "restart":
				s.auth = giauth.NewManager(s.cfg.WorkspaceRoot)
			case "other-browser":
				second, secondCookie := startSetup(t, s, false)
				w := setupRequest(s, "finish", `{"code":"`+webTestTOTPCode(second.Secret)+`"}`, &http.Cookie{Name: browserSetupCookie, Value: strings.Repeat("f", 64)}, false)
				if w.Code != 400 {
					t.Fatal("foreign binding accepted")
				}
				// The real second binding is still usable. This also makes the first stale.
				w = setupRequest(s, "finish", `{"code":"`+webTestTOTPCode(second.Secret)+`"}`, secondCookie, false)
				if w.Code != 200 {
					t.Fatal("foreign request consumed other pending")
				}
				before, _ = os.ReadFile(path)
			case "legacy-winner":
				pending, err := s.auth.StartEnrollment("legacy-owner")
				if err != nil {
					t.Fatal(err)
				}
				if _, err = s.auth.VerifyEnrollment("legacy-owner", webTestTOTPCode(pending.Secret)); err != nil {
					t.Fatal(err)
				}
				before, _ = os.ReadFile(path)
			case "write-failure":
				if err := os.MkdirAll(filepath.Join(s.cfg.WorkspaceRoot, ".gi", "auth.lock"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			w := setupRequest(s, "finish", `{"code":"`+webTestTOTPCode(p.Secret)+`"}`, c, false)
			want := 400
			if kind == "legacy-winner" || kind == "other-browser" {
				want = 409
			}
			if kind == "write-failure" {
				want = 500
			}
			if w.Code != want {
				t.Fatalf("status=%d want=%d body=%s", w.Code, want, w.Body)
			}
			for _, cookie := range w.Result().Cookies() {
				if cookie.Name == browserSessionCookie {
					t.Fatal("refused finish issued session")
				}
			}
			if strings.Contains(w.Body.String(), p.Secret) || strings.Contains(w.Body.String(), s.cfg.WorkspaceRoot) {
				t.Fatal("error leaked secret/storage")
			}
			after, err := os.ReadFile(path)
			if before != nil {
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("existing owner changed")
				}
			} else if !os.IsNotExist(err) {
				t.Fatal("failed finish persisted owner")
			}
		})
	}
}

func TestBrowserSetupTransportLoopbackHostAndPeer(t *testing.T) {
	for _, host := range []string{"localhost", "LOCALHOST:9876", "127.0.0.1:9876", "[::1]:9876", "[::1]"} {
		r := httptest.NewRequest("POST", "http://localhost", nil)
		r.Host = host
		r.RemoteAddr = "[::1]:1234"
		if !browserSetupTransport(r) {
			t.Fatalf("loopback host rejected: %s", host)
		}
	}
}
