package web

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/rcarmo/gi/internal/store"
)

func (s *Server) handleMessageDelete(w http.ResponseWriter, r *http.Request, sessionID, messageID string) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		w.WriteHeader(405)
		return
	}
	if r.URL.Query().Has("cascade") && r.URL.Query().Get("cascade") != "false" {
		writeJSON(w, 400, map[string]any{"error": "Cascade deletion is not supported"})
		return
	}
	err := s.store.DeleteMessage(r.Context(), sessionID, messageID)
	if err != nil {
		code := 500
		if errors.Is(err, sql.ErrNoRows) {
			code = 404
		} else if errors.Is(err, store.ErrMessageDeleteBusy) || errors.Is(err, store.ErrMessageDeleteProtected) {
			code = 409
		}
		writeJSON(w, code, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "deleted": []string{messageID}})
}
