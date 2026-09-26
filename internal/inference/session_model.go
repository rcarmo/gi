package inference

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

// SessionModelChoice separates durable user selection from runtime turn metadata.
// Model retains the native deterministic-shell sentinel when applicable.
type SessionModelChoice struct{ Model, Provider, Thinking string }

func SessionModel(state map[string]any, fallback SessionModelChoice) SessionModelChoice {
	choice := fallback
	if value, _ := state["selected_model"].(string); value != "" {
		choice.Model = value
	} else if value, _ := state["model"].(string); value != "" {
		choice.Model = value
	}
	if value, ok := state["selected_provider"].(string); ok {
		choice.Provider = value
	} else if value, ok := state["provider"].(string); ok {
		choice.Provider = value
	}
	if value, ok := state["thinking_level"].(string); ok {
		choice.Thinking = value
	}
	return choice
}

func (choice SessionModelChoice) Label() string {
	if strings.Contains(choice.Model, "/") || choice.Provider == "" {
		return choice.Model
	}
	return choice.Provider + "/" + choice.Model
}

func UsableSessionModel(option ModelOption) bool {
	if option.Provider == "test" && (option.ID == "test-model" || option.ID == "bootstrap") {
		return option.Enabled
	}
	return (option.Enabled || option.Authenticated) && option.Authenticated && ResolveModelContextWindow(option.Provider, option.ID) > 0
}

func ResolveSessionModel(options []ModelOption, requested string) (ModelOption, error) {
	requested = strings.TrimSpace(requested)
	if len(requested) > 256 {
		return ModelOption{}, fmt.Errorf("model selection too long")
	}
	var match *ModelOption
	for _, option := range options {
		if option.Label != requested && option.ID != requested {
			continue
		}
		if match != nil && match.Label != option.Label {
			return ModelOption{}, fmt.Errorf("ambiguous model; use provider/model")
		}
		value := option
		match = &value
	}
	if match == nil {
		return ModelOption{}, fmt.Errorf("unknown model %q", requested)
	}
	if !UsableSessionModel(*match) {
		return ModelOption{}, fmt.Errorf("model %q is unavailable or lacks credentials", requested)
	}
	return *match, nil
}

// SelectSessionModel writes only the addressed session. No provider request,
// prompt admission, global settings write, or replacement of an in-flight turn.
func SelectSessionModel(ctx context.Context, s *store.Store, sessionID string, options []ModelOption, requested string) (*store.Session, error) {
	match, err := ResolveSessionModel(options, requested)
	if err != nil {
		return nil, err
	}
	if err := CheckSessionModelContext(ctx, s, sessionID, match); err != nil {
		return nil, err
	}
	model := match.Label
	if match.Provider == "test" {
		model = match.ID
	}
	state := map[string]any{"model": model, "provider": match.Provider, "selected_model": model, "selected_provider": match.Provider}
	// Model Apply atomically resets thinking to provider default; never carry a
	// stale read of a concurrently changed choice across model selection.
	state["thinking_level"] = ""
	state["thinking_model"] = ""
	if err := s.TouchSessionState(ctx, sessionID, state); err != nil {
		return nil, err
	}
	return s.GetSession(ctx, sessionID)
}
