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

func TestPersistModelSelectionUpdatesPiSettings(t *testing.T) {
	root := t.TempDir()
	if err := PersistModelSelection(root, "ollama", "gemma4:latest", "high", []string{"qwen3:latest"}); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.DefaultProvider != "ollama" || cfg.DefaultModel != "gemma4:latest" || cfg.DefaultThinkingLevel != "high" {
		t.Fatalf("unexpected persisted config: %#v", cfg)
	}
	if len(cfg.EnabledModels) != 2 || cfg.EnabledModels[1] != "gemma4:latest" {
		t.Fatalf("unexpected enabled models after persist: %#v", cfg.EnabledModels)
	}
}

func TestLoadFallsBackToGiDefaultsWhenNoPiSettingsExist(t *testing.T) {
	root := t.TempDir()
	cfg := Load(root)
	if cfg.DefaultProvider != "opencode-zen" || cfg.DefaultModel != "opencode-zen/minimax-m2.5-free" || cfg.DefaultThinkingLevel != "low" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
	if len(cfg.EnabledModels) != 1 || cfg.EnabledModels[0] != "opencode-zen/minimax-m2.5-free" {
		t.Fatalf("unexpected enabled-model defaults: %#v", cfg.EnabledModels)
	}
	if !cfg.InboundWork.Enabled || cfg.InboundWork.IntervalMS != 500 || cfg.InboundWork.BatchSize != 8 || cfg.InboundWork.WorkerID != "web-runtime" || cfg.InboundWork.LeaseTTLMS != 2000 {
		t.Fatalf("unexpected inbound-work defaults: %#v", cfg.InboundWork)
	}
}

func TestLoadPreservesExplicitInboundWorkDisable(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"inboundWork":{"enabled":false,"interval_ms":25,"batch_size":2,"worker_id":"test-worker","lease_ttl_ms":750}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.InboundWork.Enabled {
		t.Fatalf("expected inbound work dispatcher to remain disabled, got %#v", cfg.InboundWork)
	}
	if cfg.InboundWork.IntervalMS != 25 || cfg.InboundWork.BatchSize != 2 || cfg.InboundWork.WorkerID != "test-worker" || cfg.InboundWork.LeaseTTLMS != 750 {
		t.Fatalf("unexpected explicit inbound-work config: %#v", cfg.InboundWork)
	}
}

func TestPersistClipboardModeUpdatesPiSettings(t *testing.T) {
	root := t.TempDir()
	if err := PersistClipboardMode(root, "osc52"); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.TUIClipboardMode != "osc52" {
		t.Fatalf("unexpected clipboard mode: %#v", cfg)
	}
	if err := PersistClipboardMode(root, "bogus"); err != nil {
		t.Fatal(err)
	}
	if cfg := Load(root); cfg.TUIClipboardMode != "off" {
		t.Fatalf("unexpected normalized clipboard mode: %#v", cfg)
	}
}

func TestPersistScrollbackLimitUpdatesPiSettings(t *testing.T) {
	root := t.TempDir()
	if err := PersistScrollbackLimit(root, 250); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.ScrollbackLimit != 250 {
		t.Fatalf("unexpected scrollback limit: %#v", cfg)
	}
}

func TestLoadUsesCurrentWorkingDirectoryWhenWorkspaceRootEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"defaultProvider":"cwd-provider","defaultModel":"cwd-model","enabledModels":["cwd-model"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	cfg := Load("")
	if cfg.WorkspaceRoot != root {
		t.Fatalf("unexpected workspace root: %q", cfg.WorkspaceRoot)
	}
	if cfg.DefaultProvider != "cwd-provider" || cfg.DefaultModel != "cwd-model" {
		t.Fatalf("unexpected cwd-loaded config: %#v", cfg)
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
	if !strings.Contains(cfg.SystemPrompt, "Use `tools` for staged discovery") {
		t.Fatalf("runtime guidance missing from prompt:\n%s", cfg.SystemPrompt)
	}
}
