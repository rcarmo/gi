package web

import "net/http"

// Advertise only commands/actions whose native paths exist. Startup-loaded
// skills share the native slash catalogue; no separate Skills group.
func (s *Server) handleQuickActions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	commands := []map[string]string{
		{"name": "/model", "description": "Show or select the session model", "source": "native"},
		{"name": "/compact", "description": "Compact the current session context", "source": "native"},
	}
	slash := []string{"/model", "/compact"}
	for _, skill := range s.skillQuickActions() {
		commands = append(commands, skill)
		slash = append(slash, skill["name"])
	}
	writeJSON(w, 200, map[string]any{
		"workspaceCommands": []string{"toggle-workspace", "open-explorer"},
		"slashCommands":     slash,
		"commands":          commands,
	})
}
