package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	giauth "github.com/rcarmo/gi/internal/auth"
)

const passkeyLoginCookie = "gi_passkey_login"

func passkeyOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
func (s *Server) passkeyRequest(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Cache-Control", "private, no-store")
	if _, present := r.Header["Authorization"]; present || r.URL.Query().Has("auth_token") || !providerWriteTransport(r) || !browserSameOrigin(r) || !s.auth.PasskeysAvailable(passkeyOrigin(r)) {
		writeJSON(w, 403, map[string]any{"error": "Passkeys require a configured same-origin browser connection"})
		return false
	}
	return true
}
func writePasskeyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, giauth.ErrBrowserSessionRequired), errors.Is(err, giauth.ErrRecentProofRequired), errors.Is(err, giauth.ErrStateConflict):
		writeBrowserProofError(w, err)
	case errors.Is(err, giauth.ErrPasskeysUnavailable):
		writeJSON(w, 403, map[string]any{"error": "Passkeys unavailable"})
	case errors.Is(err, giauth.ErrLastFactor):
		writeJSON(w, 409, map[string]any{"error": err.Error()})
	case errors.Is(err, giauth.ErrPasskeyNotFound):
		writeJSON(w, 404, map[string]any{"error": err.Error()})
	case errors.Is(err, giauth.ErrDuplicateCredential):
		writeJSON(w, 409, map[string]any{"error": err.Error()})
	case errors.Is(err, giauth.ErrCeremony), errors.Is(err, giauth.ErrCredential), errors.Is(err, giauth.ErrPasskeyName):
		writeJSON(w, 400, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, 500, map[string]any{"error": "Cannot complete passkey operation; refresh before retrying"})
	}
}
func (s *Server) handlePasskeyList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	if !s.passkeyRequest(w, r) {
		return
	}
	token, ok := browserOwnerCookie(w, r)
	if !ok {
		return
	}
	list, err := s.auth.ListPasskeys(token, passkeyOrigin(r))
	if err != nil {
		writePasskeyError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"passkeys": list})
}
func (s *Server) handlePasskeyCeremony(w http.ResponseWriter, r *http.Request, operation string, finish bool) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	if !s.passkeyRequest(w, r) {
		return
	}
	// Unlike ordinary cookie reads, ceremony POSTs require the browser Origin
	// header as well as exact configured RP origin and transport verification.
	if r.Header.Get("Origin") != passkeyOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "Browser origin required"})
		return
	}
	var binding string
	if operation != "login" {
		var ok bool
		binding, ok = browserOwnerCookie(w, r)
		if !ok {
			return
		}
	} else if finish {
		count := 0
		for _, c := range r.Cookies() {
			if c.Name == passkeyLoginCookie {
				binding = c.Value
				count++
			}
		}
		if count != 1 || len(binding) != 64 {
			writePasskeyError(w, giauth.ErrCeremony)
			return
		}
	} else {
		var value [32]byte
		if _, err := rand.Read(value[:]); err != nil {
			writePasskeyError(w, err)
			return
		}
		binding = hex.EncodeToString(value[:])
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		w.WriteHeader(415)
		return
	}
	var body struct {
		Name       string          `json:"name"`
		ID         string          `json:"ceremony_id"`
		Credential json.RawMessage `json:"credential"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil || decoder.Decode(new(any)) != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Invalid passkey request"})
		return
	}
	if !finish {
		if body.ID != "" || len(body.Credential) > 0 || (operation != "register" && body.Name != "") {
			writePasskeyError(w, giauth.ErrCeremony)
			return
		}
		result, err := s.auth.BeginPasskey(operation, binding, passkeyOrigin(r), body.Name)
		if err != nil {
			writePasskeyError(w, err)
			return
		}
		if operation == "login" {
			http.SetCookie(w, &http.Cookie{Name: passkeyLoginCookie, Value: binding, Path: "/api/auth/passkeys/login", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, MaxAge: 300})
		}
		writeJSON(w, 200, result)
		return
	}
	if body.Name != "" || len(body.ID) != 64 || len(body.Credential) == 0 {
		writePasskeyError(w, giauth.ErrCeremony)
		return
	}
	if operation == "register" {
		if err := s.auth.FinishPasskeyRegistration(body.ID, binding, passkeyOrigin(r), body.Credential); err != nil {
			writePasskeyError(w, err)
			return
		}
	} else {
		token, expires, err := s.auth.FinishPasskeyAssertion(body.ID, operation, binding, passkeyOrigin(r), body.Credential)
		if err != nil {
			writePasskeyError(w, err)
			return
		}
		if operation == "login" {
			http.SetCookie(w, &http.Cookie{Name: browserSessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: int(time.Until(expires).Seconds())})
			http.SetCookie(w, &http.Cookie{Name: passkeyLoginCookie, Path: "/api/auth/passkeys/login", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, MaxAge: -1})
		}
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handlePasskeyMutation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	if !s.passkeyRequest(w, r) {
		return
	}
	if r.Header.Get("Origin") != passkeyOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "Browser origin required"})
		return
	}
	token, ok := browserOwnerCookie(w, r)
	if !ok {
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		w.WriteHeader(415)
		return
	}
	var body struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2048))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&body); err != nil || decoder.Decode(new(any)) != io.EOF || body.ID == "" {
		writeJSON(w, 400, map[string]any{"error": "Invalid passkey request"})
		return
	}
	if r.URL.Path == "/api/auth/passkeys/rename" {
		err = s.auth.RenamePasskey(token, passkeyOrigin(r), body.ID, body.Name)
	} else {
		if body.Name != "" {
			writeJSON(w, 400, map[string]any{"error": "Invalid passkey request"})
			return
		}
		err = s.auth.RemovePasskey(token, passkeyOrigin(r), body.ID)
	}
	if err != nil {
		writePasskeyError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
