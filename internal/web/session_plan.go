package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rcarmo/gi/internal/plan"
	"github.com/rcarmo/gi/internal/store"
)

func (s *Server) handleSessionPlan(w http.ResponseWriter, r *http.Request, session string) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method == http.MethodGet {
		saved, err := s.store.SessionPlan(r.Context(), session)
		if err != nil {
			code := 500
			if errors.Is(err, sql.ErrNoRows) {
				code = 404
			}
			writeJSON(w, code, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "plan": saved})
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(405)
		return
	}
	if !settingsWriteAllowed(w, r, "Plan API") {
		return
	}
	var body struct {
		Markdown         *string `json:"markdown"`
		Action           string  `json:"action"`
		ExpectedRevision string  `json:"expected_revision"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2*plan.MaxBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid plan body"})
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "invalid trailing plan body"})
		return
	}
	action := body.Action
	if action == "" {
		action = "write"
	}
	if action != "write" && action != "reset" {
		writeJSON(w, 400, map[string]any{"error": "action must be write or reset"})
		return
	}
	saved, err := s.store.MutateSessionPlanConditional(r.Context(), session, plan.Mutation{Action: action, Markdown: body.Markdown}, body.ExpectedRevision)
	if err != nil {
		if errors.Is(err, store.ErrPlanRevisionRequired) {
			writeJSON(w, 428, map[string]any{"error": err.Error(), "code": "revision_required"})
			return
		}
		if errors.Is(err, store.ErrPlanRevisionConflict) {
			writeJSON(w, 409, map[string]any{"error": err.Error(), "code": "revision_conflict", "revision": saved.Revision})
			return
		}
		code := 500
		if errors.Is(err, sql.ErrNoRows) {
			code = 404
		} else if errors.Is(err, plan.ErrInvalid) {
			code = 400
		}
		writeJSON(w, code, map[string]any{"error": err.Error()})
		return
	}
	if s.turns != nil {
		s.turns.PublishPlanChanged(saved, "api", action)
	}
	writeJSON(w, 200, map[string]any{"ok": true, "plan": saved})
}
