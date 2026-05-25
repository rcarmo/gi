package tui

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func (c *chatTUI) attachCommand(text string, fields []string) []string {
	if len(fields) < 2 {
		return []string{"sys: usage /attach <path> [prompt]"}
	}
	if c.store == nil || c.engine == nil || strings.TrimSpace(c.sessionID) == "" {
		return []string{"error: /attach requires an active session"}
	}
	path := fields[1]
	if !filepath.IsAbs(path) && strings.TrimSpace(c.cfg.WorkspaceRoot) != "" {
		path = filepath.Join(c.cfg.WorkspaceRoot, path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("error: attach: %v", err)}
	}
	if len(raw) > 10<<20 {
		return []string{"error: attach: media exceeds 10 MiB limit"}
	}
	filename := filepath.Base(path)
	contentType := http.DetectContentType(raw)
	media, err := c.store.CreateMedia(c.contextOrBackground(), c.sessionID, filename, contentType, raw, map[string]any{"source": "tui", "path": path})
	if err != nil {
		return []string{fmt.Sprintf("error: attach: %v", err)}
	}
	ref := store.MediaRef{ID: store.MediaRefID(media.ID), MediaID: media.ID, SessionID: media.SessionID, Filename: media.Filename, ContentType: media.ContentType, Size: media.OriginalSize, SHA256: stringMetadata(media.Metadata, "sha256"), Source: "tui", CreatedAt: media.CreatedAt}
	prompt := strings.TrimSpace(strings.TrimPrefix(text, fields[0]+" "+fields[1]))
	if prompt == "" {
		return []string{fmt.Sprintf("attach: %s as %s (%s, %d bytes)", filename, ref.ID, contentType, len(raw))}
	}
	c.submitWithMetadata(prompt, map[string]any{"media": []store.MediaRef{ref}})
	return []string{fmt.Sprintf("attach: submitted %s as %s (%s, %d bytes)", filename, ref.ID, contentType, len(raw))}
}

func (c *chatTUI) contextOrBackground() context.Context {
	return context.Background()
}

func stringMetadata(metadata map[string]any, key string) string {
	if value, ok := metadata[key].(string); ok {
		return value
	}
	return ""
}
