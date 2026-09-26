package tui

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func (c *chatTUI) attachCommand(text string, fields []string) []string {
	if len(fields) < 2 {
		return []string{"sys: usage /attach <path> [prompt]"}
	}
	if c.store == nil || c.engine == nil || strings.TrimSpace(c.sessionID) == "" {
		return []string{"error: /attach requires an active session"}
	}
	if err := c.refreshPendingMedia(); err != nil {
		return []string{"error: attach: pending state unavailable: " + err.Error()}
	}
	if c.mediaSlotsUsed() >= maxPendingMedia {
		return []string{"error: attach: pending limit 6; /attachments or /detach first"}
	}
	path := fields[1]
	if !filepath.IsAbs(path) && strings.TrimSpace(c.cfg.WorkspaceRoot) != "" {
		path = filepath.Join(c.cfg.WorkspaceRoot, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return []string{fmt.Sprintf("error: attach: %v", err)}
	}
	if !info.Mode().IsRegular() {
		return []string{"error: attach: regular file required"}
	}
	if info.Size() > 10<<20 {
		return []string{"error: attach: media exceeds 10 MiB limit"}
	}
	file, err := os.Open(path)
	if err != nil {
		return []string{fmt.Sprintf("error: attach: %v", err)}
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, (10<<20)+1))
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
	if err := c.stageMedia(ref); err != nil {
		return []string{fmt.Sprintf("error: attach: stored %s but staging failed: %v; prompt not sent", ref.ID, err)}
	}
	prompt := strings.TrimSpace(strings.TrimPrefix(text, fields[0]+" "+fields[1]))
	if prompt != "" {
		c.submitWithMetadata(prompt, nil)
	}
	return []string{fmt.Sprintf("attach: %s as %s (%s, %d bytes); staged for prompt admission", filename, ref.ID, contentType, len(raw))}
}

// pasteImageCommand reads an image from the system clipboard, stores it in the
// session media store, and optionally attaches it to a submitted prompt. The
// clipboard reader is injectable for tests; the default uses a per-platform
// helper (wl-paste / xclip / pngpaste|osascript / powershell).
func (c *chatTUI) pasteImageCommand(text string, fields []string) []string {
	if c.store == nil || c.engine == nil || strings.TrimSpace(c.sessionID) == "" {
		return []string{"error: /paste-image requires an active session"}
	}
	if err := c.refreshPendingMedia(); err != nil {
		return []string{"error: paste-image: pending state unavailable: " + err.Error()}
	}
	if c.mediaSlotsUsed() >= maxPendingMedia {
		return []string{"error: paste-image: pending limit 6; /attachments or /detach first"}
	}
	reader := c.clipboardImageReader
	if reader == nil {
		reader = defaultClipboardImageReader(runtime.GOOS, c.lookupClipboardHelper)
	}
	if reader == nil {
		return []string{"error: paste-image: no clipboard image helper available (install wl-paste, xclip, pngpaste, or use powershell)"}
	}
	raw, contentType, err := reader()
	if err != nil {
		return []string{fmt.Sprintf("error: paste-image: %v", err)}
	}
	if len(raw) == 0 {
		return []string{"sys: paste-image: clipboard has no image"}
	}
	if len(raw) > 10<<20 {
		return []string{"error: paste-image: image exceeds 10 MiB limit"}
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = http.DetectContentType(raw)
	}
	filename := fmt.Sprintf("clipboard-%d%s", time.Now().UnixNano(), imageExtForMime(contentType))
	media, err := c.store.CreateMedia(c.contextOrBackground(), c.sessionID, filename, contentType, raw, map[string]any{"source": "tui-paste"})
	if err != nil {
		return []string{fmt.Sprintf("error: paste-image: %v", err)}
	}
	ref := store.MediaRef{ID: store.MediaRefID(media.ID), MediaID: media.ID, SessionID: media.SessionID, Filename: media.Filename, ContentType: media.ContentType, Size: media.OriginalSize, SHA256: stringMetadata(media.Metadata, "sha256"), Source: "tui-paste", CreatedAt: media.CreatedAt}
	if err := c.stageMedia(ref); err != nil {
		return []string{fmt.Sprintf("error: paste-image: stored %s but staging failed: %v; prompt not sent", ref.ID, err)}
	}
	prompt := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	if prompt != "" {
		c.submitWithMetadata(prompt, nil)
	}
	return []string{fmt.Sprintf("paste-image: %s as %s (%s, %d bytes); staged for prompt admission", filename, ref.ID, contentType, len(raw))}
}

func imageExtForMime(mime string) string {
	switch {
	case strings.Contains(mime, "png"):
		return ".png"
	case strings.Contains(mime, "jpeg"), strings.Contains(mime, "jpg"):
		return ".jpg"
	case strings.Contains(mime, "webp"):
		return ".webp"
	case strings.Contains(mime, "gif"):
		return ".gif"
	default:
		return ".img"
	}
}

// defaultClipboardImageReader returns a reader that pulls a PNG image from the
// clipboard using whichever helper is available, or nil when none is present.
func defaultClipboardImageReader(goos string, lookPath func(string) (string, error)) func() ([]byte, string, error) {
	type helper struct {
		name string
		args []string
		mime string
	}
	var candidates []helper
	switch goos {
	case "darwin":
		candidates = []helper{{"pngpaste", []string{"-"}, "image/png"}}
	case "windows":
		candidates = []helper{{"powershell", []string{"-NoProfile", "-Command", "Add-Type -AssemblyName System.Windows.Forms; $img=[System.Windows.Forms.Clipboard]::GetImage(); if($img){$ms=New-Object System.IO.MemoryStream; $img.Save($ms,[System.Drawing.Imaging.ImageFormat]::Png); [Console]::OpenStandardOutput().Write($ms.ToArray(),0,$ms.Length)}"}, "image/png"}}
	default:
		candidates = []helper{
			{"wl-paste", []string{"--type", "image/png"}, "image/png"},
			{"xclip", []string{"-selection", "clipboard", "-t", "image/png", "-o"}, "image/png"},
		}
	}
	for _, cand := range candidates {
		if _, err := lookPath(cand.name); err == nil {
			h := cand
			return func() ([]byte, string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				out, err := exec.CommandContext(ctx, h.name, h.args...).Output()
				if err != nil {
					return nil, "", fmt.Errorf("%s: %w", h.name, err)
				}
				return out, h.mime, nil
			}
		}
	}
	return nil
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
