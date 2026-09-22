package web

import (
	"encoding/json"
	"errors"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	"io"
	"net/http"
)

func (s *Server) handleSessionQueue(w http.ResponseWriter, r *http.Request, sessionID string, parts []string) {
	if _, err := s.store.GetSession(r.Context(), sessionID); err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return
	}
	var err error
	switch {
	case len(parts) == 0 && r.Method == http.MethodGet:
	case len(parts) == 0 && r.Method == http.MethodPatch:
		var req struct {
			Expected []string `json:"expected"`
			Order    []string `json:"order"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 65536))
		decoder.DisallowUnknownFields()
		if decodeErr := decoder.Decode(&req); decodeErr != nil {
			writeJSON(w, 400, map[string]any{"error": decodeErr.Error()})
			return
		}
		if decodeErr := decoder.Decode(new(any)); decodeErr != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Expected one queue mutation"})
			return
		}
		err = s.store.ReorderQueuedTurns(r.Context(), sessionID, req.Expected, req.Order)
	case len(parts) == 2 && parts[1] == "steer" && r.Method == http.MethodPost:
		var req struct {
			ActiveTurnID string `json:"active_turn_id"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if decodeErr := decoder.Decode(&req); decodeErr != nil || req.ActiveTurnID == "" {
			writeJSON(w, 400, map[string]any{"error": "Expected active_turn_id"})
			return
		}
		if decodeErr := decoder.Decode(new(any)); decodeErr != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Expected one queue mutation"})
			return
		}
		err = s.turns.SteerQueuedTurn(r.Context(), sessionID, parts[0], req.ActiveTurnID)
	case len(parts) == 1 && r.Method == http.MethodDelete:
		turn, getErr := s.store.GetTurn(r.Context(), parts[0])
		if getErr != nil || turn.SessionID != sessionID {
			writeJSON(w, 404, map[string]any{"error": "Queued turn not found in this session"})
			return
		}
		err = s.turns.CancelQueuedTurn(r.Context(), sessionID, parts[0])
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrQueueConflict) {
			status = http.StatusConflict
		}
		writeJSON(w, status, map[string]any{"error": err.Error()})
		return
	}
	if r.Method != http.MethodGet && s.turns.Topics() != nil {
		s.turns.Topics().Publish(topics.Envelope{Topic: "session.queue", SessionID: sessionID, Type: "notice", Payload: map[string]any{"type": "queue_changed"}})
	}
	items, err := s.store.ListQueuedTurns(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	activeID, _, activeErr := s.store.GetSessionActiveTurn(r.Context(), sessionID)
	if activeErr == nil {
		active, getErr := s.store.GetTurn(r.Context(), activeID)
		if getErr != nil || active.Status != "running" {
			activeID = ""
		}
	} else {
		activeID = ""
	}
	writeJSON(w, 200, map[string]any{"items": items, "active_turn_id": activeID})
}
