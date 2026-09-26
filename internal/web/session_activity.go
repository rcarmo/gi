package web

import (
	"encoding/json"
	"io"
	"net/http"
)

func (s *Server) handleSessionResume(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		StopTurnID string `json:"stop_turn_id"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if err := d.Decode(&req); err != nil || req.StopTurnID == "" {
		writeJSON(w, 400, map[string]any{"error": "Expected stop_turn_id"})
		return
	}
	if err := d.Decode(new(any)); err != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Expected one resume"})
		return
	}
	continued, err := s.turns.ResumeWebQueue(r.Context(), sessionID, req.StopTurnID)
	if err != nil {
		writeJSON(w, 409, map[string]any{"error": "Queue state changed; refresh before resuming"})
		return
	}
	writeJSON(w, 200, map[string]any{"continued": continued})
}

func (s *Server) handleSessionActivity(w http.ResponseWriter, r *http.Request, sessionID string) {
	if _, err := s.store.GetSession(r.Context(), sessionID); err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return
	}
	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		var req struct {
			TurnID string `json:"turn_id"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil || req.TurnID == "" {
			writeJSON(w, 400, map[string]any{"error": "Expected turn_id"})
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Expected one cancellation"})
			return
		}
		// StopWebActiveTurn validates the immutable session/turn pair under runner lock.
		// It cannot resolve a different run if this one has just finished.
		turn, err := s.store.GetTurn(r.Context(), req.TurnID)
		if err != nil || turn.SessionID != sessionID {
			writeJSON(w, 409, map[string]any{"error": "Active run changed"})
			return
		}
		active, _, err := s.store.GetSessionActiveTurn(r.Context(), sessionID)
		if err != nil || active != req.TurnID || turn.Status != "running" && turn.Status != "cancelling" {
			writeJSON(w, 409, map[string]any{"error": "Active run changed"})
			return
		}
		if err := s.turns.StopWebActiveTurn(r.Context(), sessionID, req.TurnID); err != nil {
			writeJSON(w, 409, map[string]any{"error": err.Error()})
			return
		}
	default:
		w.WriteHeader(405)
		return
	}
	activity, err := s.store.SessionActivity(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, activity)
}
