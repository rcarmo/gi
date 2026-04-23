package web

import (
	"encoding/json"
	"net/http"

	"github.com/rcarmo/gi/internal/tools"
)

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	// Return the list of tools the agent can call
	toolDefs := []map[string]any{
		s.scriptTool.Definition(),
		{
			"name":        "read",
			"description": "Read a workspace file.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Workspace-relative file path",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			"name":        "write",
			"description": "Write content to a workspace file. Creates parent directories.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Workspace-relative file path",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "File content to write",
					},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			"name":        "shell",
			"description": "Execute a shell command and return stdout/stderr.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Shell command to execute",
					},
				},
				"required": []string{"command"},
			},
		},
	}
	writeJSON(w, http.StatusOK, map[string]any{"tools": toolDefs})
}

func (s *Server) handleToolExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tool      string          `json:"tool"`
		Input     json.RawMessage `json:"input"`
		SessionID string          `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	switch req.Tool {
	case "script":
		var input tools.ScriptInput
		if err := json.Unmarshal(req.Input, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		input.SessionID = req.SessionID
		output := s.scriptTool.Execute(r.Context(), input)
		writeJSON(w, http.StatusOK, output)

	default:
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "unknown tool: " + req.Tool})
	}
}
