package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (s *Server) handleSessionReset(w http.ResponseWriter, r *http.Request, sessionID string) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]any{"error": "method not allowed"})
		return
	}
	if !settingsWriteAllowed(w, r, "Reset API") {
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid reset body"})
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "invalid trailing data"})
		return
	}
	queued, err := s.turns.ResetSession(r.Context(), sessionID, body.Text)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, 404, map[string]any{"error": "session not found"})
		return
	}
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	status := http.StatusOK
	if queued {
		status = http.StatusAccepted
	}
	writeJSON(w, status, map[string]any{"ok": true, "queued": queued})
}
