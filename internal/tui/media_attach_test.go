package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestAttachCommandStoresLocalMedia(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_attach", "Attach", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	root := t.TempDir()
	path := filepath.Join(root, "note.txt")
	if err := os.WriteFile(path, []byte("hello attach"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_attach", cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultModel: "bootstrap", AssistantName: "Neo"}}
	lines := strings.Join(c.attachCommand("/attach note.txt", []string{"/attach", "note.txt"}), "\n")
	if !strings.Contains(lines, "attach: note.txt as media:") {
		t.Fatalf("unexpected attach output:\n%s", lines)
	}
	items, err := s.ListMedia(ctx, "session_attach")
	if err != nil {
		t.Fatalf("list media: %v", err)
	}
	if len(items) != 1 || items[0].Filename != "note.txt" || items[0].Metadata["source"] != "tui" {
		t.Fatalf("unexpected media rows: %#v", items)
	}
}
