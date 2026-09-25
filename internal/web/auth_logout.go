package web

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"
)

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	token, ok := browserOwnerCookie(w, r)
	if !ok {
		return
	}
	if r.Header.Get("Origin") != passkeyOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "Browser origin required"})
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		w.WriteHeader(415)
		return
	}
	var body map[string]json.RawMessage
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	if err = decoder.Decode(&body); err != nil || body == nil || len(body) != 0 || decoder.Decode(new(any)) != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Invalid sign-out request"})
		return
	}
	if err = s.auth.LogoutBrowserSession(token); err != nil {
		writeBrowserProofError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: browserSessionCookie, Value: "", Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, Expires: time.Unix(1, 0), MaxAge: -1})
	writeJSON(w, 200, map[string]any{"ok": true})
}
