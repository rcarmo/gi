package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
)

func (s *Server) modelCatalogue() []inference.ModelOption {
	_, options := inference.ListRuntimeOptions(s.cfg.DefaultProvider, s.cfg.DefaultModel, s.cfg.EnabledModels)
	return options
}

func (s *Server) sessionModelPayload(ctx context.Context, session *store.Session) (map[string]any, error) {
	choice := inference.SessionModel(session.State, inference.SessionModelChoice{Model: s.cfg.DefaultModel, Provider: s.cfg.DefaultProvider})
	current, thinking := choice.Label(), choice.Thinking
	options := s.modelCatalogue()
	var selected inference.ModelOption
	for _, option := range options {
		if option.Label == current {
			selected = option
			break
		}
	}
	usage, err := inference.SessionContextUsage(ctx, s.store, session.ID, selected.ContextWindow)
	if err != nil {
		return nil, err
	}
	return map[string]any{"model_options": options, "models": options, "model": current, "current": current, "thinking_level": thinking, "thinking_level_label": thinking, "supports_thinking": selected.Reasoning, "context_window": selected.ContextWindow, "context_usage": usage}, nil
}

func (s *Server) selectSessionModel(r *http.Request, sessionID, requested string) (map[string]any, error) {
	session, err := inference.SelectSessionModel(r.Context(), s.store, sessionID, s.modelCatalogue(), requested)
	if err != nil {
		return nil, err
	}
	return s.sessionModelPayload(r.Context(), session)
}

func (s *Server) handleSessionModel(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := s.store.GetSession(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return
	}
	if r.Method == http.MethodGet {
		payload, err := s.sessionModelPayload(r.Context(), session)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, 200, payload)
		return
	}
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Model string `json:"model"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "Invalid model selection"})
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Expected one model selection"})
		return
	}
	payload, err := s.selectSessionModel(r, sessionID, req.Model)
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, payload)
}

func (s *Server) handleModelCommand(w http.ResponseWriter, r *http.Request, sessionID, prompt string) bool {
	fields := strings.Fields(prompt)
	if len(fields) == 0 || fields[0] != "/model" {
		return false
	}
	session, err := s.store.GetSession(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return true
	}
	payload, err := s.sessionModelPayload(r.Context(), session)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return true
	}
	if len(fields) > 1 {
		payload, err = s.selectSessionModel(r, sessionID, strings.Join(fields[1:], " "))
	}
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return true
	}
	writeJSON(w, 200, map[string]any{"command": map[string]any{"model_label": payload["current"], "thinking_level": payload["thinking_level"], "supports_thinking": payload["supports_thinking"], "context_window": payload["context_window"], "context_usage": payload["context_usage"], "message": "Current model: " + payload["current"].(string)}})
	return true
}
