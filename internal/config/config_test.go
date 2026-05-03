package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReadsWorkspacePiAndPiclawConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".piclaw"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".piclaw", "config.json"), []byte(`{"assistant":{"assistantName":"Neo","assistantAvatar":"a.png"},"user":{"userName":"Rui","userAvatar":"u.png"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"defaultProvider":"ollama","defaultModel":"gemma4:latest","defaultThinkingLevel":"medium","enabledModels":["gemma4:latest"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.AssistantName != "Neo" || cfg.UserName != "Rui" || cfg.DefaultModel != "gemma4:latest" || cfg.DefaultProvider != "ollama" || cfg.DefaultThinkingLevel != "medium" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if len(cfg.EnabledModels) != 1 || cfg.EnabledModels[0] != "gemma4:latest" {
		t.Fatalf("unexpected enabled models: %#v", cfg.EnabledModels)
	}
	for _, want := range []string{"agentic coding assistant", "## Tool environment", "`skills`", "## Path and safety policy"} {
		if !strings.Contains(cfg.SystemPrompt, want) {
			t.Fatalf("system prompt missing %q:\n%s", want, cfg.SystemPrompt)
		}
	}
}

func TestLoadWrapsAgentsInstructionsInRuntimePrompt(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("Project rule: keep APIs stable."), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if !strings.Contains(cfg.SystemPrompt, "## Workspace instructions") || !strings.Contains(cfg.SystemPrompt, "Project rule: keep APIs stable.") {
		t.Fatalf("workspace instructions missing from prompt:\n%s", cfg.SystemPrompt)
	}
	if !strings.Contains(cfg.SystemPrompt, "Use `tools` to list or inspect registered tools") {
		t.Fatalf("runtime guidance missing from prompt:\n%s", cfg.SystemPrompt)
	}
}
