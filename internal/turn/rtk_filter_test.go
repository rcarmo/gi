package turn

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func loadExample(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "examples", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example %s: %v", name, err)
	}
	return string(data)
}

func TestRTKJokerFilterRewritesShellCoverage(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".gi", "extensions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rtk-tool-filter.joke"), []byte(loadExample(t, "rtk-tool-filter.joke")), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cfg := config.RuntimeConfig{WorkspaceRoot: root, DefaultModel: "bootstrap", MaxIterations: 64, Agents: routing.AgentsConfig{List: []routing.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}}}
	e := NewWithRuntimeConfig(s, cfg, "")

	covered := []string{
		"ls -la",
		"tree .",
		"cat README.md",
		"head -20 README.md",
		"tail -20 log.txt",
		"grep -R TODO .",
		"rg TODO .",
		"find . -name '*.go'",
		"diff a b",
		"git status",
		"gh pr list",
		"go test ./...",
		"npm test",
		"pnpm test",
		"yarn test",
		"bun test",
		"jest",
		"vitest run",
		"pytest",
		"cargo test",
		"ruff check .",
		"golangci-lint run",
		"eslint .",
		"biome check .",
		"tsc --noEmit",
		"docker ps",
		"kubectl get pods",
		"make test",
		"cmake --build build",
	}
	for _, command := range covered {
		call := goai.ToolCall{Type: "toolCall", ID: "tc", Name: "shell", Arguments: map[string]any{"command": command}}
		resp, err := e.emitHook(context.Background(), HookRequest{Name: HookToolCall, ToolCall: &call})
		if err != nil {
			t.Fatalf("hook for %q: %v", command, err)
		}
		if resp.ToolCall == nil {
			t.Fatalf("expected rewrite for %q", command)
		}
		got, _ := resp.ToolCall.Arguments["command"].(string)
		if !strings.Contains(got, "command -v rtk") || !strings.Contains(got, "rtk "+command) || !strings.Contains(got, "else "+command) {
			t.Fatalf("bad rewrite for %q: %q", command, got)
		}
	}
}

func TestRTKJokerFilterRewritesReadToShell(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".gi", "extensions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rtk-tool-filter.joke"), []byte(loadExample(t, "rtk-tool-filter.joke")), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cfg := config.RuntimeConfig{WorkspaceRoot: root, DefaultModel: "bootstrap", MaxIterations: 64, Agents: routing.AgentsConfig{List: []routing.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}}}
	e := NewWithRuntimeConfig(s, cfg, "")
	call := goai.ToolCall{Type: "toolCall", ID: "read1", Name: "read", Arguments: map[string]any{"path": "README.md"}}
	resp, err := e.emitHook(context.Background(), HookRequest{Name: HookToolCall, ToolCall: &call})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ToolCall == nil || resp.ToolCall.Name != "shell" {
		t.Fatalf("expected read->shell rewrite: %#v", resp.ToolCall)
	}
	cmd, _ := resp.ToolCall.Arguments["command"].(string)
	if !strings.Contains(cmd, "rtk read 'README.md'") || !strings.Contains(cmd, "cat 'README.md'") {
		t.Fatalf("bad read rewrite: %q", cmd)
	}

	vfsCall := goai.ToolCall{Type: "toolCall", ID: "read2", Name: "read", Arguments: map[string]any{"path": "vfs://notes/file.md"}}
	resp, err = e.emitHook(context.Background(), HookRequest{Name: HookToolCall, ToolCall: &vfsCall})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ToolCall != nil {
		t.Fatalf("vfs read should not be rewritten: %#v", resp.ToolCall)
	}
}
