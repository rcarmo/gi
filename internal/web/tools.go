package web

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/tools"
)

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	// Return the list of tools the agent can call.
	toolDefs := []map[string]any{
		s.scriptTool.Definition(),
		{
			"name":        "tools",
			"description": "List available tools or get details about a specific tool. Use with no arguments to list all tools (names + short descriptions). Pass a tool name via the `name` argument to get its full schema and usage. Use `query` to filter tools by keyword.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":  map[string]any{"type": "string", "description": "Exact tool name to get full details for"},
					"query": map[string]any{"type": "string", "description": "Filter tools by keyword in name or description"},
				},
			},
		},
		{
			"name":        "read",
			"description": "Read text content from a workspace file or managed VFS asset. Supports both workspace paths and `vfs://` paths.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Workspace-relative path or vfs://namespace/path",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			"name":        "write",
			"description": "Write text content to a workspace file or managed VFS asset. Creates parent directories for workspace paths.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Workspace-relative path or vfs://namespace/path",
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

type toolOutput struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

type readToolInput struct {
	Path string `json:"path"`
}

type writeToolInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type shellToolInput struct {
	Command string `json:"command"`
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

	case "read":
		var input readToolInput
		if err := json.Unmarshal(req.Input, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		output := executeReadTool(r.Context(), s, input.Path)
		writeJSON(w, http.StatusOK, output)

	case "write":
		var input writeToolInput
		if err := json.Unmarshal(req.Input, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if output, err := executeWriteTool(r.Context(), s, input.Path, input.Content); err != nil {
			writeJSON(w, http.StatusOK, toolOutput{Error: err.Error(), Result: output})
		} else {
			writeJSON(w, http.StatusOK, toolOutput{Result: output})
		}

	case "shell":
		var input shellToolInput
		if err := json.Unmarshal(req.Input, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		output := executeShellTool(r.Context(), input.Command)
		writeJSON(w, http.StatusOK, output)

	default:
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "unknown tool: " + req.Tool})
	}
}

func executeReadTool(_ context.Context, s *Server, path string) toolOutput {
	resolved, err := tools.ResolveToolPath(s.cfg.WorkspaceRoot, path, false)
	if err != nil {
		return toolOutput{Error: err.Error()}
	}
	if resolved.IsVFS() {
		_, raw, err := s.store.GetVFSFileContent(context.Background(), resolved.VFSNamespace, resolved.VFSPath)
		if err != nil {
			return toolOutput{Error: err.Error()}
		}
		return toolOutput{Result: string(raw)}
	}
	content, err := os.ReadFile(resolved.WorkspacePath)
	if err != nil {
		return toolOutput{Error: err.Error()}
	}
	return toolOutput{Result: string(content)}
}

func executeWriteTool(_ context.Context, s *Server, path string, content string) (string, error) {
	resolved, err := tools.ResolveToolPath(s.cfg.WorkspaceRoot, path, true)
	if err != nil {
		return "", err
	}
	if resolved.IsVFS() {
		_, err := s.store.SaveVFSFile(context.Background(), resolved.VFSNamespace, resolved.VFSPath,
			inferContentTypeFromFilename(resolved.VFSPath), []byte(content), map[string]any{})
		if err != nil {
			return "", err
		}
		return "written", nil
	}
	dir := filepath.Dir(resolved.WorkspacePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(resolved.WorkspacePath, []byte(content), 0o644); err != nil {
		return "", err
	}
	return "written", nil
}

func inferContentTypeFromFilename(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if extType := mime.TypeByExtension(ext); extType != "" {
		return extType
	}
	return "text/plain"
}

func executeShellTool(ctx context.Context, command string) toolOutput {
	command = strings.TrimSpace(command)
	if command == "" {
		return toolOutput{Error: "command is required"}
	}
	execCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(execCtx, "sh", "-lc", command)
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		if execCtx.Err() != nil {
			return toolOutput{Result: output, Error: fmt.Sprintf("command error: %v: %v", err, execCtx.Err())}
		}
		return toolOutput{Result: output, Error: err.Error()}
	}
	return toolOutput{Result: output}
}
