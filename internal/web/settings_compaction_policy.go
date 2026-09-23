package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/rcarmo/gi/internal/config"
)

func (s *Server) compactionPolicyPayload(saved config.CompactionPolicySnapshot) map[string]any {
	active := s.turns.CompactionPolicy()
	return map[string]any{"saved": saved, "active": active, "restart_required": saved.Policy != active}
}

func (s *Server) handleCompactionPolicy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	switch r.Method {
	case http.MethodGet:
		saved, err := config.ReadCompactionPolicy(s.cfg.WorkspaceRoot)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": "Cannot read saved compaction policy; check .pi/settings.json without replacing it."})
			return
		}
		writeJSON(w, 200, s.compactionPolicyPayload(saved))
	case http.MethodPatch:
		if err := http.NewCrossOriginProtection().Check(r); err != nil {
			writeJSON(w, 403, map[string]any{"error": "Cross-origin settings writes are not allowed"})
			return
		}
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if media != "application/json" {
			writeJSON(w, 415, map[string]any{"error": "JSON content type required"})
			return
		}
		var body struct {
			Revision         string `json:"revision"`
			Enabled          *bool  `json:"enabled"`
			ContextWindow    *int   `json:"context_window"`
			ReserveTokens    *int   `json:"reserve_tokens"`
			KeepRecentTokens *int   `json:"keep_recent_tokens"`
			ThresholdTokens  *int   `json:"threshold_tokens"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			writeJSON(w, 400, map[string]any{"error": "Invalid compaction policy request"})
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Expected one JSON object"})
			return
		}
		if body.Enabled == nil || body.ContextWindow == nil || body.ReserveTokens == nil || body.KeepRecentTokens == nil || body.ThresholdTokens == nil {
			writeJSON(w, 400, map[string]any{"error": "All compaction policy fields are required"})
			return
		}
		saved, err := config.SaveCompactionPolicy(s.cfg.WorkspaceRoot, body.Revision, config.CompactionPolicyEdit{Enabled: *body.Enabled, ContextWindow: *body.ContextWindow, ReserveTokens: *body.ReserveTokens, KeepRecentTokens: *body.KeepRecentTokens, ThresholdTokens: *body.ThresholdTokens})
		if err != nil {
			status := 500
			message := "Cannot save compaction policy; reload saved policy to verify settings before retrying."
			if errors.Is(err, config.ErrSettingsConflict) {
				status = 409
				message = err.Error()
			}
			if errors.Is(err, config.ErrCompactionPolicyInvalid) {
				status = 400
				message = err.Error()
			}
			writeJSON(w, status, map[string]any{"error": message})
			return
		}
		writeJSON(w, 200, s.compactionPolicyPayload(saved))
	default:
		w.Header().Set("Allow", "GET, PATCH")
		writeJSON(w, 405, map[string]any{"error": "Method not allowed"})
	}
}
