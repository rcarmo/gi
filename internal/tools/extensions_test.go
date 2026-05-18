package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
)

func TestDiscoverExtensionScripts(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".gi", "extensions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "smart.joke"), []byte("nil"), 0o644); err != nil {
		t.Fatal(err)
	}
	ext := DiscoverExtensionScripts(root)
	if len(ext) != 1 || ext[0].Engine != "joker" || ext[0].Path != filepath.Join(".gi", "extensions", "smart.joke") {
		t.Fatalf("extensions: %#v", ext)
	}
}

func TestLoadWorkspaceExtensions(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".gi", "extensions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ok.joke"), []byte("nil"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	scriptTool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: root})
	res := LoadWorkspaceExtensions(context.Background(), root, scriptTool)
	if len(res) != 1 || res[0].Path != filepath.Join(".gi", "extensions", "ok.joke") {
		t.Fatalf("load results: %#v", res)
	}
}
