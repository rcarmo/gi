package scripting

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// JokerRunner executes Clojure scripts using the Joker interpreter baked into
// the gi binary itself.
//
// Upstream Joker ships primarily as a CLI and the module source on its own is
// not directly consumable as an import without generated files. We vendor a
// generated copy locally and execute it in-process so scripts can access live
// bridge state in the same runtime.
type JokerRunner struct{}

func NewJokerRunner() *JokerRunner {
	return &JokerRunner{}
}

func (r *JokerRunner) Name() string { return "joker" }

func (r *JokerRunner) Execute(ctx context.Context, script string, bridge *Bridge) (string, error) {
	return ExecuteEmbeddedJoker(ctx, script, bridge)
}

func (r *JokerRunner) ExecuteFile(ctx context.Context, path string, bridge *Bridge) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read script %s: %w", path, err)
	}
	return r.Execute(ctx, string(content), bridge)
}

func buildBridgeState(ctx context.Context, bridge *Bridge) map[string]any {
	state := map[string]any{
		"session-id": bridge.SessionID,
	}
	if bridge.Funcs.GetConfig != nil {
		cfg, _ := bridge.Funcs.GetConfig(ctx)
		if cfg != nil {
			state["config"] = cfg
			state["runtime-config"] = cfg
		}
	}
	if bridge.Funcs.GetSessionState != nil {
		ss, _ := bridge.Funcs.GetSessionState(ctx)
		if ss != nil {
			state["session-state"] = ss
		}
	}
	if bridge.Funcs.GetSessionInfo != nil {
		info, _ := bridge.Funcs.GetSessionInfo(ctx)
		if info != nil {
			state["session-info"] = info
		}
	}
	if bridge.Funcs.ListTurns != nil {
		turns, _ := bridge.Funcs.ListTurns(ctx, 0)
		if turns != nil {
			state["turns"] = turns
		}
	}
	if bridge.Funcs.ListMessages != nil {
		messages, _ := bridge.Funcs.ListMessages(ctx, 0)
		if messages != nil {
			state["messages"] = messages
		}
	}
	return state
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// FindScripts discovers script files in a directory.
func FindScripts(dir string, extensions []string) ([]string, error) {
	if extensions == nil {
		extensions = []string{".joke", ".clj"}
	}
	var scripts []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		for _, ext := range extensions {
			if strings.HasSuffix(e.Name(), ext) {
				scripts = append(scripts, e.Name())
				break
			}
		}
	}
	return scripts, err
}
