package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	giauth "github.com/rcarmo/gi/internal/auth"
)

// Unlike ordinary API auth, owner proof never accepts bearer/query credentials
// or falls back between credential sources. Stored purpose is checked by auth.
func browserOwnerCookie(w http.ResponseWriter, r *http.Request) (string, bool) {
	w.Header().Set("Cache-Control", "private, no-store")
	if _, explicit := r.Header["Authorization"]; explicit || r.URL.Query().Has("auth_token") {
		writeJSON(w, 403, map[string]any{"error": "browser cookie authentication required"})
		return "", false
	}
	if !providerWriteTransport(r) || !browserSameOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "Same-origin TLS or loopback is required"})
		return "", false
	}
	var token string
	count := 0
	for _, cookie := range r.Cookies() {
		if cookie.Name == browserSessionCookie {
			token = cookie.Value
			count++
		}
	}
	if count != 1 || token == "" {
		writeJSON(w, 401, map[string]any{"error": "browser owner sign-in required"})
		return "", false
	}
	return token, true
}

func writeBrowserProofError(w http.ResponseWriter, err error) {
	status, message := 500, "Cannot verify authentication"
	switch {
	case errors.Is(err, giauth.ErrBrowserSessionRequired):
		status, message = 401, err.Error()
	case errors.Is(err, giauth.ErrRecentProofRequired):
		status, message = 403, err.Error()
	case errors.Is(err, giauth.ErrInvalidFactorProof):
		status, message = 401, err.Error()
	case errors.Is(err, giauth.ErrStateConflict):
		status, message = 409, "Authentication state changed; retry"
	}
	writeJSON(w, status, map[string]any{"error": message})
}

func (s *Server) handleBrowserProof(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	token, ok := browserOwnerCookie(w, r)
	if !ok {
		return
	}
	proof, err := s.auth.BrowserSessionProof(token)
	if err != nil {
		writeBrowserProofError(w, err)
		return
	}
	writeJSON(w, 200, proof)
}

func (s *Server) handleBrowserReauthTOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	token, ok := browserOwnerCookie(w, r)
	if !ok {
		return
	}
	if _, err := s.auth.BrowserSessionProof(token); err != nil {
		writeBrowserProofError(w, err)
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeJSON(w, 415, map[string]any{"error": "application/json required"})
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil {
		writeJSON(w, 400, map[string]any{"error": "Invalid authentication request"})
		return
	}
	if decoder.Decode(new(any)) != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Invalid authentication request"})
		return
	}
	// Revalidates session purpose/expiry/revocation under the auth writer lock.
	proof, err := s.auth.ReauthenticateBrowserTOTP(token, body.Code)
	if err != nil {
		writeBrowserProofError(w, err)
		return
	}
	writeJSON(w, 200, proof)
}
