package web

import (
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleSessionSearch(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	if _, err := s.store.GetSession(r.Context(), sessionID); err != nil {
		writeJSON(w, 404, map[string]any{"error": "Session not found"})
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "current"
	}
	limit, offset := 50, 0
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": "Invalid limit"})
			return
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": "Invalid offset"})
			return
		}
	}
	if q == "" || len(q) > 512 || limit < 1 || limit > 100 || offset < 0 || offset > 10000 || (scope != "current" && scope != "root" && scope != "all") {
		writeJSON(w, 400, map[string]any{"error": "Invalid search query, scope or bounds"})
		return
	}
	messages, err := s.store.SearchMessages(r.Context(), sessionID, q, scope, limit, offset)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "Search unavailable"})
		return
	}
	writeJSON(w, 200, map[string]any{"messages": messages, "query": q, "scope": scope})
}
