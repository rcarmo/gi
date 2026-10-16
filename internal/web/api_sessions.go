package web

import (
	"encoding/json"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"net/http"
)

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		sessions, err := s.store.ListSessions(ctx)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if sessions == nil {
			sessions = []store.Session{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
	case http.MethodPost:
		var req struct {
			Title    string `json:"title"`
			AgentID  string `json:"agent_id"`
			ForkFrom string `json:"fork_from"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		id := store.NowID("session")
		if req.ForkFrom != "" {
			agentID := req.AgentID
			title := req.Title
			if title == "" && agentID != "" {
				title = "@" + agentID
			}
			session, err := s.store.CloneSession(ctx, req.ForkFrom, id, title, agentID)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusCreated, session)
			return
		}
		agentID := req.AgentID
		if agentID == "" {
			agentID = "agent"
		}
		title := req.Title
		if title == "" {
			title = "@" + agentID
		}
		alloc := gisession.AllocateDefaultSession(agentID, "gi", "default", id)
		session, _, err := s.store.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: id, Title: title, State: map[string]any{"status": "idle", "queue_count": 0, "model": s.cfg.DefaultModel, "provider": s.cfg.DefaultProvider, "thinking_level": s.cfg.DefaultThinkingLevel}, Allocation: alloc})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, session)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
