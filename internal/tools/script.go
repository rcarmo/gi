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
	store *store.Store
	cfg   config.RuntimeConfig
	joker scripting.Runner
	js    scripting.Runner
}

// ScriptInput is what the agent sends to invoke the script tool.
type ScriptInput struct {
	Script    string `json:"script,omitempty"`
	Path      string `json:"path,omitempty"`
	Engine    string `json:"engine,omitempty"` // "joker" or "js" (default: auto-detect)
	SessionID string `json:"session_id,omitempty"`
}

// ScriptOutput is what the script tool returns.
type ScriptOutput struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

func NewScriptTool(s *store.Store, cfg config.RuntimeConfig) *ScriptTool {
	return &ScriptTool{
		store: s,
		cfg:   cfg,
		joker: scripting.NewJokerRunner(),
		js:    scripting.NewGojaRunner(),
	}
}

// Definition returns the tool metadata for the agent.
func (t *ScriptTool) Definition() map[string]any {
	return map[string]any{
		"name":        "script",
		"description": "Execute a script with access to gi's session state, config, and workspace files. Supports Joker/Clojure (.joke, .clj) and JavaScript (.js). The gi bridge object provides sessionId, config, sessionState, listMessages(), readFile(), writeFile(), listDir(), and log().",
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"script": map[string]any{
					"type":        "string",
					"description": "Inline script to execute",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "Workspace-relative path to a script file",
				},
				"engine": map[string]any{
					"type":        "string",
					"description": "Script engine: 'js' (JavaScript/goja, compiled-in) or 'joker' (Clojure, requires joker binary). Default: js for .js files, joker for .joke/.clj, js for inline.",
					"enum":        []string{"js", "joker"},
				},
			},
		},
	}
}

// Execute runs the script and returns the output.
func (t *ScriptTool) Execute(ctx context.Context, input ScriptInput) ScriptOutput {
	bridge := t.buildBridge(input.SessionID)
	runner := t.resolveRunner(input.Engine, input.Path)

	var result string
	var err error

	if input.Path != "" {
		fullPath := filepath.Join(t.cfg.WorkspaceRoot, input.Path)
		clean := filepath.Clean(fullPath)
		if !strings.HasPrefix(clean, filepath.Clean(t.cfg.WorkspaceRoot)+string(os.PathSeparator)) && clean != filepath.Clean(t.cfg.WorkspaceRoot) {
			return ScriptOutput{Error: "path escapes workspace"}
		}
		log.Printf("script[%s]: executing file %s", runner.Name(), input.Path)
		result, err = runner.ExecuteFile(ctx, fullPath, bridge)
	} else if input.Script != "" {
		log.Printf("script[%s]: executing inline (%d chars)", runner.Name(), len(input.Script))
		result, err = runner.Execute(ctx, input.Script, bridge)
	} else {
		return ScriptOutput{Error: "either script or path is required"}
	}

	if err != nil {
		log.Printf("script[%s]: error: %v", runner.Name(), err)
		return ScriptOutput{Result: result, Error: err.Error()}
	}

	log.Printf("script[%s]: result (%d chars)", runner.Name(), len(result))
	return ScriptOutput{Result: result}
}

func (t *ScriptTool) resolveRunner(engine, path string) scripting.Runner {
	if engine == "joker" {
		return t.joker
	}
	if engine == "js" || engine == "javascript" {
		return t.js
	}
	// Auto-detect from file extension
	if strings.HasSuffix(path, ".joke") || strings.HasSuffix(path, ".clj") {
		return t.joker
	}
	if strings.HasSuffix(path, ".js") {
		return t.js
	}
	// Default: JS (compiled-in, no external dependency)
	return t.js
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
