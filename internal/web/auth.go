package web

import (
	"encoding/json"
	"net/http"

	giauth "github.com/rcarmo/gi/internal/auth"
)

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	status, err := s.auth.Status()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleAuthEnrollStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !giauth.LoopbackRequest(r) {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "initial enrollment is only allowed from loopback"})
		return
	}
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	pending, err := s.auth.StartEnrollment(req.Username)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"username": pending.Username, "secret": pending.Secret, "otpauth_url": pending.URL, "expires_in_seconds": 600})
}

func (s *Server) handleAuthEnrollVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !giauth.LoopbackRequest(r) {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "initial enrollment is only allowed from loopback"})
		return
	}
	var req struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	state, err := s.auth.VerifyEnrollment(req.Username, req.Code)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "username": state.Username, "totp_enabled": state.TOTPEnabled})
}

func (s *Server) handleAuthTOTPVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	token, expires, err := s.auth.VerifyLogin(req.Username, req.Code)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "token": token, "token_type": "bearer", "expires_at": expires})
}
