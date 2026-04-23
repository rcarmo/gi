package scripting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// JokerRunner executes Clojure scripts via the Joker CLI binary.
//
// We use subprocess execution rather than compiling Joker/glojure in-process
// because:
//   - Joker's core package requires code generation and doesn't compile as a library
//   - glojure's NewEnvironment panics during init when called from an HTTP handler
//     in a long-running server (global state corruption)
//
// The Joker CLI binary must be installed (brew install candid82/brew/joker).
// Bridge state is injected as *gi-bridge* via read-string of an EDN literal.
type JokerRunner struct {
	// JokerPath overrides PATH lookup. Auto-detected if empty.
	JokerPath string
}

func NewJokerRunner() *JokerRunner {
	return &JokerRunner{}
}

func (r *JokerRunner) Name() string { return "joker" }

func (r *JokerRunner) resolveJoker() (string, error) {
	if r.JokerPath != "" {
		return r.JokerPath, nil
	}
	path, err := exec.LookPath("joker")
	if err != nil {
		return "", fmt.Errorf("joker not found in PATH (install: brew install candid82/brew/joker): %w", err)
	}
	return path, nil
}

func (r *JokerRunner) Execute(ctx context.Context, script string, bridge *Bridge) (string, error) {
	joker, err := r.resolveJoker()
	if err != nil {
		return "", err
	}

	preamble := buildPreamble(ctx, bridge)
	fullScript := preamble + "\n(println (do\n" + script + "\n))"

	cmd := exec.CommandContext(ctx, joker, "-")
	cmd.Stdin = strings.NewReader(fullScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("joker: %s", strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("joker: %w", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (r *JokerRunner) ExecuteFile(ctx context.Context, path string, bridge *Bridge) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read script %s: %w", path, err)
	}
	return r.Execute(ctx, string(content), bridge)
}

func buildPreamble(ctx context.Context, bridge *Bridge) string {
	state := map[string]any{
		"session-id": bridge.SessionID,
	}
	if bridge.Funcs.GetConfig != nil {
		cfg, _ := bridge.Funcs.GetConfig(ctx)
		if cfg != nil {
			state["config"] = cfg
		}
	}
	if bridge.Funcs.GetSessionState != nil {
		ss, _ := bridge.Funcs.GetSessionState(ctx)
		if ss != nil {
			state["session-state"] = ss
		}
	}
	return fmt.Sprintf("(require '[joker.json :as json] '[joker.walk :as walk])\n(def ^:private *gi-bridge* (walk/keywordize-keys (json/read-string %q)))", mustJSON(state))
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
