package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"

	giauth "github.com/rcarmo/gi/internal/auth"
	"github.com/rcarmo/gi/internal/inference"
)

func providerWriteTransport(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if !giauth.LoopbackRequest(r) {
		return false
	}
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
func (s *Server) providerSettingsPayload(r *http.Request) (map[string]any, error) {
	snapshot, err := inference.ReadProviderSettings()
	if err != nil {
		return nil, err
	}
	return map[string]any{"providers": snapshot.Providers, "revision": snapshot.Revision, "scope": snapshot.Scope, "can_write": providerWriteTransport(r)}, nil
}
func (s *Server) handleProviderSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	switch r.Method {
	case http.MethodGet:
		payload, err := s.providerSettingsPayload(r)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": "Cannot read provider credentials; repair the native credential file before retrying."})
			return
		}
		writeJSON(w, 200, payload)
	case http.MethodPatch, http.MethodDelete:
		if !providerWriteTransport(r) {
			writeJSON(w, 403, map[string]any{"error": "Credential writes require HTTPS or a loopback connection"})
			return
		}
		if err := http.NewCrossOriginProtection().Check(r); err != nil {
			writeJSON(w, 403, map[string]any{"error": "Cross-origin credential writes are not allowed"})
			return
		}
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if media != "application/json" {
			writeJSON(w, 415, map[string]any{"error": "JSON content type required"})
			return
		}
		var body struct {
			Provider string  `json:"provider"`
			Revision string  `json:"revision"`
			Key      *string `json:"key"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			writeJSON(w, 400, map[string]any{"error": "Invalid provider credential request"})
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Expected one JSON object"})
			return
		}
		var err error
		if r.Method == http.MethodPatch {
			if body.Key == nil {
				writeJSON(w, 400, map[string]any{"error": "API key required"})
				return
			}
			err = inference.SaveProviderKey(body.Provider, body.Revision, *body.Key)
		} else {
			if body.Key != nil {
				writeJSON(w, 400, map[string]any{"error": "Remove accepts provider and revision only"})
				return
			}
			err = inference.RemoveProviderKey(body.Provider, body.Revision)
		}
		body.Key = nil
		if err != nil {
			status := 500
			message := "Credential change failed; refresh providers to verify the store before retrying."
			if errors.Is(err, inference.ErrCredentialConflict) {
				status = 409
				message = "Credential store changed or is busy; refresh providers and retry"
			}
			if errors.Is(err, inference.ErrCredentialInvalid) {
				status = 400
				message = "Use a supported API-key provider and a nonempty key of at most 4096 bytes without whitespace; OAuth/token entries are read-only"
			}
			writeJSON(w, status, map[string]any{"error": message})
			return
		}
		payload, err := s.providerSettingsPayload(r)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": "Credential change completed but metadata reload failed; refresh providers to verify"})
			return
		}
		writeJSON(w, 200, payload)
	default:
		w.Header().Set("Allow", "GET, PATCH, DELETE")
		writeJSON(w, 405, map[string]any{"error": "Method not allowed"})
	}
}
