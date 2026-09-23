package web

import "net/http"

// Advertise only commands/actions whose native paths exist. Skills, settings,
// chat-only and terminal/VNC panes must not become plausible no-op actions.
func (s *Server) handleQuickActions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, 200, map[string]any{
		"workspaceCommands": []string{"toggle-workspace", "open-explorer"},
		"slashCommands":     []string{"/model", "/compact"},
		"commands": []map[string]string{
			{"name": "/model", "description": "Show or select the session model", "source": "native"},
			{"name": "/compact", "description": "Compact the current session context", "source": "native"},
		},
	})
}
