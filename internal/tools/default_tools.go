package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rcarmo/gi/internal/rtk"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func ScriptWithPayload(engine, name string, payload map[string]any, script string) string {
	if strings.TrimSpace(script) == "" {
		return script
	}
	b, _ := json.Marshal(payload)
	if engine == "joker" || engine == "" && strings.HasPrefix(strings.TrimSpace(script), "(") {
		return fmt.Sprintf("(def *gi-%s* (walk/keywordize-keys (json/read-string %q)))\n%s", name, string(b), script)
	}
	return fmt.Sprintf("gi.%s = %s; gi.%sPayload = gi.%s; gi.toolArgs = (gi.tool && gi.tool.arguments) || {};\n%s", name, string(b), name, name, script)
}

func ExecuteRead(ctx context.Context, workspaceRoot string, s *store.Store, call goai.ToolCall) (string, error) {
	path, _ := call.Arguments["path"].(string)
	if path == "" {
		return "", fmt.Errorf("read: path is required")
	}
	resolved, err := ResolveToolPath(workspaceRoot, path, false)
	if err != nil {
		return "", err
	}
	if resolved.IsVFS() {
		if resolved.VFSNamespace == "fts" {
			return ReadFTSQuery(ctx, workspaceRoot, s, resolved.VFSPath)
		}
		_, raw, err := s.GetVFSFileContent(ctx, resolved.VFSNamespace, resolved.VFSPath)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
	content, err := os.ReadFile(resolved.WorkspacePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func ExecuteWrite(ctx context.Context, workspaceRoot string, s *store.Store, call goai.ToolCall) (string, error) {
	path, _ := call.Arguments["path"].(string)
	content, _ := call.Arguments["content"].(string)
	if path == "" {
		return "", fmt.Errorf("write: path is required")
	}
	resolved, err := ResolveToolPath(workspaceRoot, path, true)
	if err != nil {
		return "", err
	}
	if resolved.IsVFS() {
		_, err := s.SaveVFSFile(ctx, resolved.VFSNamespace, resolved.VFSPath, "text/plain", []byte(content), map[string]any{})
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

func ExecuteRTK(ctx context.Context, workspaceRoot string, call goai.ToolCall) (string, error) {
	command, _ := call.Arguments["command"].(string)
	if command == "" {
		return "", fmt.Errorf("rtk: command is required")
	}
	filterOnly, _ := call.Arguments["filter_only"].(bool)
	output, _ := call.Arguments["output"].(string)
	var err error
	if !filterOnly {
		cmd := exec.CommandContext(ctx, "sh", "-lc", command)
		cmd.Dir = workspaceRoot
		out, runErr := cmd.CombinedOutput()
		output = string(out)
		err = runErr
	}
	res := rtk.Filter(command, output)
	b, _ := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return string(b), fmt.Errorf("exit: %w", err)
	}
	return string(b), nil
}

func ExecuteShell(ctx context.Context, workspaceRoot string, call goai.ToolCall) (string, error) {
	command, _ := call.Arguments["command"].(string)
	if command == "" {
		return "", fmt.Errorf("shell: command is required")
	}
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	cmd.Dir = workspaceRoot
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return output, fmt.Errorf("exit: %w", err)
	}
	return output, nil
}
