package scripting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// JokerRunner executes Clojure scripts via the Joker CLI.
//
// Scripts receive the bridge state as a JSON blob on stdin and can call
// bridge functions by writing JSON-RPC-like requests to a named pipe
// (future) or by using the injected helper functions.
//
// For v1, we use a simpler model: the bridge state is injected as a
// Clojure map bound to *gi-bridge*, and scripts return their result
// as the last expression.
type JokerRunner struct {
	// JokerPath is the path to the joker binary. Auto-detected if empty.
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
		return "", fmt.Errorf("joker not found in PATH: %w", err)
	}
	return path, nil
}

func (r *JokerRunner) Execute(ctx context.Context, script string, bridge *Bridge) (string, error) {
	joker, err := r.resolveJoker()
	if err != nil {
		return "", err
	}

	// Build the preamble that injects bridge state
	preamble, err := buildJokerPreamble(ctx, bridge)
	if err != nil {
		return "", fmt.Errorf("build preamble: %w", err)
	}

	fullScript := preamble + "\n(println (do \n" + script + "\n))"

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

// buildJokerPreamble creates Clojure code that sets up the bridge state
// as a var accessible to the script.
func buildJokerPreamble(ctx context.Context, bridge *Bridge) (string, error) {
	state := map[string]any{
		"session-id": bridge.SessionID,
	}

	if bridge.Funcs.GetConfig != nil {
		cfg, err := bridge.Funcs.GetConfig(ctx)
		if err == nil {
			state["config"] = cfg
		}
	}

	if bridge.Funcs.GetSessionState != nil {
		ss, err := bridge.Funcs.GetSessionState(ctx)
		if err == nil {
			state["session-state"] = ss
		}
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`(require '[joker.json :as json] '[joker.walk :as walk])
(def ^:private *gi-bridge* (walk/keywordize-keys (json/read-string %q)))
`, string(stateJSON)), nil
}

// FindScripts discovers .joke and .clj scripts in a directory.
func FindScripts(dir string, extensions []string) ([]string, error) {
	if extensions == nil {
		extensions = []string{".joke", ".clj"}
	}
	var scripts []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				scripts = append(scripts, path)
				break
			}
		}
		return nil
	})
	return scripts, err
}
