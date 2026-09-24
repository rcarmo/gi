package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	giauth "github.com/rcarmo/gi/internal/auth"
)

const browserSessionCookie = "gi_session"

// Explicit credentials take precedence, including an invalid bearer. Browser
// cookies never grant connectivity/script endpoints a different authority.
func (s *Server) authenticatedRequest(r *http.Request) (bool, error) {
	if strings.TrimSpace(r.Header.Get("Authorization")) != "" || r.URL.Query().Has("auth_token") {
		return s.auth.ValidateBearerRequestWithError(r)
	}
	cookie, err := r.Cookie(browserSessionCookie)
	if err != nil || !providerWriteTransport(r) || !browserSameOrigin(r) {
		return false, nil
	}
	return s.auth.ValidateTokenWithError(cookie.Value)
}

// Unlike CrossOriginProtection alone, also fence cookie-authenticated GET/SSE
// requests. Native bearer clients retain their existing contract.
func browserSameOrigin(r *http.Request) bool {
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if err != nil || u.Scheme != scheme || !strings.EqualFold(u.Host, r.Host) || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return false
		}
	}
	return http.NewCrossOriginProtection().Check(r) == nil
}

// Separate from the bearer API: browser callers never receive a token in JSON.
// Only direct TLS or a loopback peer AND Host may carry browser credentials.
func (s *Server) handleAuthSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(405)
		return
	}
	if !providerWriteTransport(r) || !browserSameOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "Browser sign-in requires same-origin HTTPS or localhost"})
		return
	}
	media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if media != "application/json" {
		writeJSON(w, 415, map[string]any{"error": "JSON content type required"})
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, 400, map[string]any{"error": "Invalid sign-in request"})
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Expected one JSON object"})
		return
	}
	token, expires, err := s.auth.VerifyLogin("", body.Code)
	if err != nil {
		status := 500
		message := "Cannot verify sign-in; try again"
		switch err.Error() {
		case "invalid user", "TOTP is not enrolled", "invalid TOTP code":
			status, message = 401, "Invalid authentication code"
		}
		if errors.Is(err, giauth.ErrStateConflict) {
			status, message = 409, "Sign-in state changed; try again"
		}
		writeJSON(w, status, map[string]any{"error": message})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: browserSessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: int(time.Until(expires).Seconds())})
	writeJSON(w, 200, map[string]any{"ok": true, "expires_at": expires})
}
