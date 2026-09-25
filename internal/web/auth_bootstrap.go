package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"

	giauth "github.com/rcarmo/gi/internal/auth"
)

const browserSetupCookie = "gi_setup"
const browserSetupPath = "/api/auth/setup"

func browserSetupTransport(r *http.Request) bool {
	if !giauth.LoopbackRequest(r) {
		return false
	}
	host := r.Host
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || ip != nil && ip.IsLoopback()
}

func setBrowserSetupCookie(w http.ResponseWriter, r *http.Request, token string) {
	age := 600
	expires := time.Now().Add(10 * time.Minute)
	if token == "" {
		age = -1
		expires = time.Unix(1, 0)
	}
	http.SetCookie(w, &http.Cookie{Name: browserSetupCookie, Value: token, Path: browserSetupPath, HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, MaxAge: age, Expires: expires})
}

func (s *Server) handleBrowserSetup(w http.ResponseWriter, r *http.Request, operation string) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(405)
		return
	}
	_, authorization := r.Header["Authorization"]
	if authorization || r.URL.RawQuery != "" || !browserSetupTransport(r) || !browserSameOrigin(r) || r.Header.Get("Origin") != passkeyOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "Browser setup requires same-origin loopback access"})
		return
	}
	var binding string
	count := 0
	for _, c := range r.Cookies() {
		if c.Name == browserSetupCookie {
			binding = c.Value
			count++
		}
	}
	if count > 1 || count == 1 && len(binding) != 64 || operation != "start" && count != 1 {
		writeJSON(w, 401, map[string]any{"error": "Browser setup cookie required"})
		return
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		w.WriteHeader(415)
		return
	}
	var body map[string]json.RawMessage
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	if err = decoder.Decode(&body); err != nil || body == nil || decoder.Decode(new(any)) != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Invalid browser setup request"})
		return
	}
	code := ""
	if operation == "finish" {
		if len(body) != 1 || json.Unmarshal(body["code"], &code) != nil || len(code) != 6 {
			writeJSON(w, 400, map[string]any{"error": "Invalid browser setup code"})
			return
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				writeJSON(w, 400, map[string]any{"error": "Invalid browser setup code"})
				return
			}
		}
	} else if len(body) != 0 {
		writeJSON(w, 400, map[string]any{"error": "Invalid browser setup request"})
		return
	}
	switch operation {
	case "start":
		pending, token, err := s.auth.BeginBrowserSetup(binding)
		if err != nil {
			s.writeBrowserSetupError(w, err)
			return
		}
		setBrowserSetupCookie(w, r, token)
		writeJSON(w, 200, map[string]any{"username": pending.Username, "secret": pending.Secret, "url": pending.URL, "expires_in_seconds": 600})
	case "cancel":
		s.auth.CancelBrowserSetup(binding)
		setBrowserSetupCookie(w, r, "")
		writeJSON(w, 200, map[string]any{"ok": true})
	case "finish":
		token, expires, err := s.auth.FinishBrowserSetup(binding, code)
		// Finish attempts with a parsed code always consume their pending binding.
		setBrowserSetupCookie(w, r, "")
		if err != nil {
			s.writeBrowserSetupError(w, err)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: browserSessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: int(time.Until(expires).Seconds())})
		writeJSON(w, 200, map[string]any{"ok": true})
	}
}

func (s *Server) writeBrowserSetupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, giauth.ErrOwnerExists):
		writeJSON(w, 409, map[string]any{"error": err.Error()})
	case errors.Is(err, giauth.ErrBrowserSetupLimit):
		writeJSON(w, 429, map[string]any{"error": err.Error()})
	case errors.Is(err, giauth.ErrBrowserSetup), errors.Is(err, giauth.ErrInvalidFactorProof):
		writeJSON(w, 400, map[string]any{"error": err.Error()})
	case errors.Is(err, giauth.ErrStateConflict):
		writeJSON(w, 409, map[string]any{"error": "Authentication state changed; start setup again"})
	default:
		writeJSON(w, 500, map[string]any{"error": "Cannot complete browser setup; check status before starting again"})
	}
}
