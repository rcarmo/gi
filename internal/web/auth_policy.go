package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	giauth "github.com/rcarmo/gi/internal/auth"
)

func (s *Server) handleLoginPolicy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != "GET" && r.Method != "POST" {
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(405)
		return
	}
	token, ok := browserOwnerCookie(w, r)
	if !ok {
		return
	}
	var result giauth.LoginPolicy
	var err error
	if r.Method == "GET" {
		result, err = s.auth.ReadLoginPolicy(token, passkeyOrigin(r))
	} else {
		if r.Header.Get("Origin") != passkeyOrigin(r) {
			writeJSON(w, 403, map[string]any{"error": "Browser origin required"})
			return
		}
		mediaType, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if e != nil || mediaType != "application/json" {
			writeJSON(w, 415, map[string]any{"error": "application/json required"})
			return
		}
		var body struct {
			Policy   string `json:"policy"`
			Revision string `json:"revision"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if e = decoder.Decode(&body); e != nil || decoder.Decode(new(any)) != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Invalid policy request"})
			return
		}
		result, err = s.auth.UpdateLoginPolicy(token, passkeyOrigin(r), body.Policy, body.Revision)
	}
	if err != nil {
		switch {
		case errors.Is(err, giauth.ErrLoginPolicy):
			writeJSON(w, 400, map[string]any{"error": err.Error()})
		case errors.Is(err, giauth.ErrPolicyLockout):
			writeJSON(w, 409, map[string]any{"error": err.Error()})
		default:
			writeBrowserProofError(w, err)
		}
		return
	}
	writeJSON(w, 200, result)
}
