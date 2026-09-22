package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
)

func modelLabel(provider, model string) string {
	if strings.Contains(model, "/") || provider == "" {
		return model
	}
	return provider + "/" + model
}

func (s *Server) modelCatalogue() []inference.ModelOption {
	_, options := inference.ListRuntimeOptions(s.cfg.DefaultProvider, s.cfg.DefaultModel, s.cfg.EnabledModels)
	return options
}

// Reject synthetic catalogue placeholders and missing credentials. Selection is
// local metadata only: no provider request, no agent turn, no global config write.
func usableSessionModel(option inference.ModelOption) bool {
	if option.Provider == "test" && (option.ID == "test-model" || option.ID == "bootstrap") {
		return option.Enabled
	}
	return (option.Enabled || option.Authenticated) && option.Authenticated && inference.ResolveModelContextWindow(option.Provider, option.ID) > 0
}

func (s *Server) sessionModelPayload(session *store.Session) map[string]any {
	current, _ := session.State["selected_model"].(string)
	if current == "" {
		current, _ = session.State["model"].(string)
	}
	if current == "" {
		current = s.cfg.DefaultModel
	}
	provider, _ := session.State["selected_provider"].(string)
	if provider == "" {
		provider, _ = session.State["provider"].(string)
	}
	if provider == "" {
		provider = s.cfg.DefaultProvider
	}
	current = modelLabel(provider, current)
	thinking, _ := session.State["thinking_level"].(string)
	options := s.modelCatalogue()
	var selected inference.ModelOption
	for _, option := range options {
		if option.Label == current {
			selected = option
			break
		}
	}
	return map[string]any{"model_options": options, "models": options, "model": current, "current": current, "thinking_level": thinking, "thinking_level_label": thinking, "supports_thinking": selected.Reasoning, "context_window": selected.ContextWindow, "context_usage": nil}
}

func (s *Server) selectSessionModel(r *http.Request, sessionID, requested string) (map[string]any, error) {
	requested = strings.TrimSpace(requested)
	var match *inference.ModelOption
	for _, option := range s.modelCatalogue() {
		if option.Label == requested || option.ID == requested {
			if match != nil && match.Label != option.Label {
				return nil, fmt.Errorf("ambiguous model; use provider/model")
			}
			value := option
			match = &value
		}
	}
	if len(requested) > 256 {
		return nil, fmt.Errorf("model selection too long")
	}
	if match == nil {
		return nil, fmt.Errorf("unknown model %q", requested)
	}
	if !usableSessionModel(*match) {
		return nil, fmt.Errorf("model %q is unavailable or lacks credentials", requested)
	}
	state := map[string]any{"model": match.Label, "provider": match.Provider, "selected_model": match.Label, "selected_provider": match.Provider}
	if match.Provider == "test" {
		state["model"] = match.ID
		state["selected_model"] = match.ID
	} // Native deterministic shell models retain their sentinel IDs.
	if !match.Reasoning {
		state["thinking_level"] = ""
	}
	if err := s.store.TouchSessionState(r.Context(), sessionID, state); err != nil {
		return nil, err
	}
	session, err := s.store.GetSession(r.Context(), sessionID)
	if err != nil {
		return nil, err
	}
	return s.sessionModelPayload(session), nil
}

func (s *Server) handleSessionModel(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := s.store.GetSession(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return
	}
	if r.Method == http.MethodGet {
		writeJSON(w, 200, s.sessionModelPayload(session))
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
	payload := s.sessionModelPayload(session)
	if len(fields) > 1 {
		payload, err = s.selectSessionModel(r, sessionID, strings.Join(fields[1:], " "))
	}
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return true
	}
	writeJSON(w, 200, map[string]any{"command": map[string]any{"model_label": payload["current"], "thinking_level": payload["thinking_level"], "supports_thinking": payload["supports_thinking"], "context_window": payload["context_window"], "message": "Current model: " + payload["current"].(string)}})
	return true
}
