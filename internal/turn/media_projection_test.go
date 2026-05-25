package turn

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func TestUserMessageWithProviderSafeMediaProjectsImageBlocks(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_projection", "Media", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	rawPNG := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	media, err := s.CreateMedia(ctx, "session_projection", "shot.png", "image/png", rawPNG, nil)
	if err != nil {
		t.Fatalf("create media: %v", err)
	}
	runner := &sessionRunner{store: s, engine: New(s)}
	msg := runner.userMessageWithProviderSafeMedia(ctx, "session_projection", "describe", map[string]any{"media": []any{map[string]any{"id": store.MediaRefID(media.ID)}}})
	if len(msg.Content) != 2 {
		t.Fatalf("expected text+image blocks, got %#v", msg.Content)
	}
	if msg.Content[0].Type != "text" || msg.Content[0].Text != "describe" {
		t.Fatalf("unexpected text block: %#v", msg.Content[0])
	}
	if msg.Content[1].Type != "image" || msg.Content[1].MimeType != "image/png" || msg.Content[1].Data != base64.StdEncoding.EncodeToString(rawPNG) {
		t.Fatalf("unexpected image block: %#v", msg.Content[1])
	}
}

func TestUserMessageWithProviderSafeMediaFallsBackForUnsupportedMedia(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_projection_fallback", "Media", map[string]any{"status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	media, err := s.CreateMedia(ctx, "session_projection_fallback", "notes.txt", "text/plain", []byte("not image"), nil)
	if err != nil {
		t.Fatalf("create media: %v", err)
	}
	runner := &sessionRunner{store: s, engine: New(s)}
	msg := runner.userMessageWithProviderSafeMedia(ctx, "session_projection_fallback", "describe", map[string]any{"media": []string{store.MediaRefID(media.ID)}})
	if len(msg.Content) != 1 || msg.Content[0].Type != "text" {
		t.Fatalf("expected text fallback only, got %#v", msg.Content)
	}
	if msg.Content[0].Text != "describe\n\n[media attachments included]" {
		t.Fatalf("unexpected fallback text: %q", msg.Content[0].Text)
	}
}
