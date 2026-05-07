package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func TestReadFTSQueryMessagesAndWorkspace(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "note.txt"), []byte("queue overflow happened here"), 0o644); err != nil {
		t.Fatalf("seed workspace file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal", "tools"), 0o755); err != nil {
		t.Fatalf("mkdir internal/tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "tools", "resolver.go"), []byte("package tools\n// ResolveToolPath helper"), 0o644); err != nil {
		t.Fatalf("seed tooling file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "third_party", "joker"), 0o755); err != nil {
		t.Fatalf("mkdir third_party/joker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "third_party", "joker", "core.joke"), []byte("(println :hook)"), 0o644); err != nil {
		t.Fatalf("seed joker file: %v", err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	if _, err := s.CreateSession(ctx, "session_fts", "FTS", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_fts", "session_fts", "user", "queue overflow in tool execution", map[string]any{"kind": "chat"}); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_fts", "session_fts", "completed", "check queue overflow", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}

	msgDoc, err := ReadFTSQuery(ctx, root, s, "messages?q=queue+overflow&limit=5")
	if err != nil {
		t.Fatalf("fts messages: %v", err)
	}
	if !strings.Contains(msgDoc, "vfs://chat/sessions/session_fts/messages/msg_fts.md") {
		t.Fatalf("expected chat vfs source link in messages search: %q", msgDoc)
	}

	wsDoc, err := ReadFTSQuery(ctx, root, s, "workspace?q=queue+overflow&limit=5")
	if err != nil {
		t.Fatalf("fts workspace: %v", err)
	}
	if !strings.Contains(wsDoc, "note.txt") {
		t.Fatalf("expected workspace hit in workspace search: %q", wsDoc)
	}

	allDoc, err := ReadFTSQuery(ctx, root, s, "all?q=queue+overflow&limit=5")
	if err != nil {
		t.Fatalf("fts all: %v", err)
	}
	if !strings.Contains(allDoc, "## Messages") || !strings.Contains(allDoc, "## Workspace") {
		t.Fatalf("expected sectioned all-search output: %q", allDoc)
	}

	nsDoc, err := ReadFTSQuery(ctx, root, s, "tooling?q=ResolveToolPath&limit=5")
	if err != nil {
		t.Fatalf("fts tooling namespace: %v", err)
	}
	if !strings.Contains(nsDoc, "# Workspace namespace: tooling") || !strings.Contains(nsDoc, "Hints") {
		t.Fatalf("expected namespace heading+hints, got: %q", nsDoc)
	}
	if !strings.Contains(nsDoc, "internal/tools/resolver.go") {
		t.Fatalf("expected tooling namespace to include tooling file, got: %q", nsDoc)
	}

	jokerDoc, err := ReadFTSQuery(ctx, root, s, "go-joker?q=hook&limit=5")
	if err != nil {
		t.Fatalf("fts go-joker namespace: %v", err)
	}
	if !strings.Contains(jokerDoc, "# Workspace namespace: go-joker") {
		t.Fatalf("expected go-joker namespace heading, got: %q", jokerDoc)
	}

	helpDoc, err := ReadFTSQuery(ctx, root, s, "help")
	if err != nil {
		t.Fatalf("fts help: %v", err)
	}
	for _, want := range []string{"gi", "go-joker", "tooling", "hint:"} {
		if !strings.Contains(helpDoc, want) {
			t.Fatalf("expected help to mention %q, got: %q", want, helpDoc)
		}
	}
}
