package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func strictSendReceiptQuery(r *http.Request) (string, error) {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(q) != 1 {
		return "", fmt.Errorf("invalid query")
	}
	values := q["client_request_id"]
	if len(values) != 1 || !store.ValidWebSendToken(values[0]) {
		return "", fmt.Errorf("invalid token")
	}
	return values[0], nil
}

func (s *Server) handleSendReceipt(w http.ResponseWriter, r *http.Request, sessionID string) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	query, err := strictSendReceiptQuery(r)
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": "one client_request_id is required"})
		return
	}
	if _, err = s.store.GetSession(r.Context(), sessionID); err != nil {
		writeJSON(w, 404, map[string]any{"error": "session not found"})
		return
	}
	raw, confirmed, err := s.store.GetWebSendReceipt(r.Context(), sessionID, query)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "send receipt unavailable"})
		return
	}
	if !confirmed {
		writeJSON(w, 200, map[string]any{"confirmed": false})
		return
	}
	var result turn.SubmitResult
	if err = json.Unmarshal(raw, &result); err != nil || result.TurnID == "" || result.SessionID == "" {
		writeJSON(w, 500, map[string]any{"error": "send receipt unavailable"})
		return
	}
	// Protect against stale/deleted targets and malformed persisted references.
	target, err := s.store.GetTurn(r.Context(), result.TurnID)
	if err != nil || target.SessionID != result.SessionID || (result.SessionID != sessionID && result.SourceSessionID != sessionID) {
		writeJSON(w, 200, map[string]any{"confirmed": false})
		return
	}
	writeJSON(w, 200, map[string]any{"confirmed": true, "client_request_id": query, "source_session_id": sessionID, "result": result})
}
