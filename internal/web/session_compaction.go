package web

import (
	"encoding/json"
	"errors"
	"github.com/rcarmo/gi/internal/store"
	"io"
	"net/http"
)

func (s *Server) handleSessionCompaction(w http.ResponseWriter, r *http.Request, sessionID string) {
	w.Header().Set("Cache-Control", "private, no-store")
	if _, err := s.store.GetSession(r.Context(), sessionID); err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return
	}
	if r.Method == http.MethodGet {
		state, err := s.turns.ManualCompactionState(r.Context(), sessionID)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, 200, state)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if err := d.Decode(&req); err != nil || req.Token == "" {
		writeJSON(w, 400, map[string]any{"error": "Expected context token"})
		return
	}
	if err := d.Decode(new(any)); err != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Expected one compaction request"})
		return
	}
	result, err := s.turns.SubmitManualCompaction(r.Context(), sessionID, req.Token)
	if err != nil {
		status := 500
		if errors.Is(err, store.ErrQueueConflict) || errors.Is(err, store.ErrContextChanged) {
			status = 409
		}
		writeJSON(w, status, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 202, result)
}
