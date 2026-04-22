package config

import (
	"os"
	"path/filepath"
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
}
