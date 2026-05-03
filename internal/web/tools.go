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
	if s.turns != nil {
		active := map[string]bool{}
		for _, name := range s.turns.ActiveTools() {
			active[name] = true
		}
		toolDefs := []map[string]any{}
		for _, tool := range s.turns.ToolEntries() {
			var params any
			_ = json.Unmarshal(tool.Parameters, &params)
			toolDefs = append(toolDefs, map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  params,
				"source":      tool.Source,
				"kind":        tool.Kind,
				"weight":      tool.Weight,
				"activation":  tool.Activation,
				"active":      active[tool.Name],
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"tools": toolDefs})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tools": []map[string]any{}})
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
