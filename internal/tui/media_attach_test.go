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

func TestPasteImageStoresClipboardImage(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_paste", "@agent", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 13}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_paste", cfg: config.RuntimeConfig{DefaultModel: "bootstrap", AssistantName: "Neo"}}
	c.clipboardImageReader = func() ([]byte, string, error) { return pngHeader, "image/png", nil }
	lines := strings.Join(c.pasteImageCommand("/paste-image", []string{"/paste-image"}), "\n")
	if !strings.Contains(lines, "paste-image:") || !strings.Contains(lines, "image/png") {
		t.Fatalf("unexpected paste output:\n%s", lines)
	}
	items, err := s.ListMedia(ctx, "session_paste")
	if err != nil {
		t.Fatalf("list media: %v", err)
	}
	if len(items) != 1 || items[0].ContentType != "image/png" || items[0].Metadata["source"] != "tui-paste" {
		t.Fatalf("unexpected media rows: %#v", items)
	}
	if !strings.HasSuffix(items[0].Filename, ".png") {
		t.Fatalf("unexpected filename: %q", items[0].Filename)
	}
}

func TestPasteImageEmptyClipboard(t *testing.T) {
	s, _ := store.Open("file::memory:?cache=shared")
	defer s.Close()
	_, _ = s.CreateSession(context.Background(), "session_empty_paste", "@agent", map[string]any{"status": "idle"})
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_empty_paste", cfg: config.RuntimeConfig{DefaultModel: "bootstrap"}}
	c.clipboardImageReader = func() ([]byte, string, error) { return nil, "", nil }
	lines := strings.Join(c.pasteImageCommand("/paste-image", []string{"/paste-image"}), "\n")
	if !strings.Contains(lines, "clipboard has no image") {
		t.Fatalf("unexpected empty clipboard output: %s", lines)
	}
}

func TestDefaultClipboardImageReaderSelectsHelper(t *testing.T) {
	avail := map[string]bool{"wl-paste": true}
	look := func(name string) (string, error) {
		if avail[name] {
			return "/usr/bin/" + name, nil
		}
		return "", os.ErrNotExist
	}
	if r := defaultClipboardImageReader("linux", look); r == nil {
		t.Fatalf("expected wl-paste reader on linux")
	}
	none := func(string) (string, error) { return "", os.ErrNotExist }
	if r := defaultClipboardImageReader("linux", none); r != nil {
		t.Fatalf("expected nil reader when no helper available")
	}
}
