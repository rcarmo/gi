package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/scripting"
	"github.com/rcarmo/gi/internal/store"
)

// ScriptTool is a tool that executes Joker/Clojure scripts with access
// to gi's bridge state. It can run inline scripts or script files from
// the workspace.
type ScriptTool struct {
	store  *store.Store
	cfg    config.RuntimeConfig
	runner scripting.Runner
}

// ScriptInput is what the agent sends to invoke the script tool.
type ScriptInput struct {
	// Script is inline script code to execute.
	Script string `json:"script,omitempty"`

	// Path is a workspace-relative path to a script file.
	Path string `json:"path,omitempty"`

	// SessionID is the active session (injected by the turn engine).
	SessionID string `json:"session_id,omitempty"`
}

// ScriptOutput is what the script tool returns.
type ScriptOutput struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

func NewScriptTool(s *store.Store, cfg config.RuntimeConfig) *ScriptTool {
	return &ScriptTool{
		store:  s,
		cfg:    cfg,
		runner: scripting.NewJokerRunner(),
	}
}

// Definition returns the tool metadata for the agent.
func (t *ScriptTool) Definition() map[string]any {
	return map[string]any{
		"name":        "script",
		"description": "Execute a Joker/Clojure script with access to gi's session state, config, and workspace files. Scripts receive *gi-bridge* with session-id, config, and session-state.",
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"script": map[string]any{
					"type":        "string",
					"description": "Inline Joker/Clojure script to execute",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "Workspace-relative path to a .joke or .clj script file",
				},
			},
		},
	}
}

// Execute runs the script and returns the output.
func (t *ScriptTool) Execute(ctx context.Context, input ScriptInput) ScriptOutput {
	bridge := t.buildBridge(input.SessionID)

	var result string
	var err error

	if input.Path != "" {
		fullPath := filepath.Join(t.cfg.WorkspaceRoot, input.Path)
		// Safety: ensure path doesn't escape workspace
		clean := filepath.Clean(fullPath)
		if !strings.HasPrefix(clean, filepath.Clean(t.cfg.WorkspaceRoot)+string(os.PathSeparator)) && clean != filepath.Clean(t.cfg.WorkspaceRoot) {
			return ScriptOutput{Error: "path escapes workspace"}
		}
		log.Printf("script: executing file %s", input.Path)
		result, err = t.runner.ExecuteFile(ctx, fullPath, bridge)
	} else if input.Script != "" {
		log.Printf("script: executing inline (%d chars)", len(input.Script))
		result, err = t.runner.Execute(ctx, input.Script, bridge)
	} else {
		return ScriptOutput{Error: "either script or path is required"}
	}

	if err != nil {
		log.Printf("script: error: %v", err)
		return ScriptOutput{Error: err.Error()}
	}

	log.Printf("script: result (%d chars)", len(result))
	return ScriptOutput{Result: result}
}

func (t *ScriptTool) buildBridge(sessionID string) *scripting.Bridge {
	return scripting.NewBridge(sessionID, scripting.BridgeFuncs{
		GetSessionState: func(ctx context.Context) (map[string]any, error) {
			session, err := t.store.GetSession(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			return session.State, nil
		},
		SetSessionState: func(ctx context.Context, patch map[string]any) error {
			return t.store.TouchSessionState(ctx, sessionID, patch)
		},
		ListMessages: func(ctx context.Context, limit int) ([]map[string]any, error) {
			msgs, err := t.store.ListMessages(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			var result []map[string]any
			start := 0
			if limit > 0 && len(msgs) > limit {
				start = len(msgs) - limit
			}
			for _, m := range msgs[start:] {
				result = append(result, map[string]any{
					"id":      m.ID,
					"role":    m.Role,
					"content": m.Content,
					"payload": m.Payload,
				})
			}
			return result, nil
		},
		AddMessage: func(ctx context.Context, role, content string) error {
			return t.store.AddMessage(ctx, store.NowID("msg"), sessionID, role, content, map[string]any{"kind": "script"})
		},
		ListTurnEvents: func(ctx context.Context, turnID string) ([]map[string]any, error) {
			events, err := t.store.ListTurnEvents(ctx, turnID)
			if err != nil {
				return nil, err
			}
			var result []map[string]any
			for _, e := range events {
				result = append(result, map[string]any{
					"seq":     e.Seq,
					"type":    e.Type,
					"payload": e.Payload,
				})
			}
			return result, nil
		},
		GetConfig: func(ctx context.Context) (map[string]any, error) {
			b, _ := json.Marshal(t.cfg)
			var m map[string]any
			json.Unmarshal(b, &m)
			return m, nil
		},
		ReadFile: func(ctx context.Context, path string) (string, error) {
			full := filepath.Join(t.cfg.WorkspaceRoot, path)
			data, err := os.ReadFile(full)
			if err != nil {
				return "", err
			}
			return string(data), nil
		},
		WriteFile: func(ctx context.Context, path, content string) error {
			full := filepath.Join(t.cfg.WorkspaceRoot, path)
			dir := filepath.Dir(full)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			return os.WriteFile(full, []byte(content), 0o644)
		},
		ListDir: func(ctx context.Context, path string) ([]map[string]any, error) {
			full := filepath.Join(t.cfg.WorkspaceRoot, path)
			entries, err := os.ReadDir(full)
			if err != nil {
				return nil, err
			}
			var result []map[string]any
			for _, e := range entries {
				info, _ := e.Info()
				result = append(result, map[string]any{
					"name":  e.Name(),
					"isDir": e.IsDir(),
					"size":  info.Size(),
				})
			}
			return result, nil
		},
		Exec: func(ctx context.Context, command string) (string, error) {
			return "", fmt.Errorf("exec: not yet available in script bridge")
		},
		Log: func(ctx context.Context, level, message string) {
			log.Printf("script[%s]: %s", level, message)
		},
	})
}
