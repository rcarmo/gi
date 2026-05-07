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
}
