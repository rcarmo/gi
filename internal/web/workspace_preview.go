package web

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const workspacePreviewLimit = 256 * 1024

// OpenRoot confines the actual open, including symlinks and concurrent renames.
// A preflight lexical check alone would leave a check/open race.
func (s *Server) openWorkspaceFile(rawPath string) (*os.File, os.FileInfo, string, error) {
	root := s.cfg.WorkspaceRoot
	if root == "" {
		root = "/workspace"
	}
	path := strings.TrimSpace(rawPath)
	if path == "" {
		return nil, nil, "", fmt.Errorf("missing path")
	}
	if filepath.IsAbs(path) {
		var err error
		path, err = filepath.Rel(root, path)
		if err != nil {
			return nil, nil, "", err
		}
	}
	path = filepath.Clean(path)
	if !filepath.IsLocal(path) {
		return nil, nil, "", fmt.Errorf("path escapes workspace")
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, nil, "", err
	}
	defer dir.Close()
	entry, err := dir.Stat(path)
	if err != nil {
		return nil, nil, "", err
	}
	if !entry.Mode().IsRegular() {
		return nil, nil, "", fmt.Errorf("not a regular file")
	}
	f, err := dir.Open(path)
	if err != nil {
		return nil, nil, "", err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		if err == nil {
			err = fmt.Errorf("not a regular file")
		}
		return nil, nil, "", err
	}
	return f, info, filepath.ToSlash(path), nil
}

func workspaceContentType(path string, sample []byte) string {
	detected := http.DetectContentType(sample)
	switch detected {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp":
		return detected
	}
	if strings.HasPrefix(detected, "text/") || len(sample) == 0 {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".md", ".markdown":
			return "text/markdown"
		case ".html", ".htm":
			return "text/html"
		case ".svg":
			return "image/svg+xml"
		}
	}
	return strings.Split(detected, ";")[0]
}

func (s *Server) handleWorkspacePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	limit := 20000
	if raw := r.URL.Query().Get("max_bytes"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > workspacePreviewLimit {
			writeJSON(w, 400, map[string]any{"error": "invalid preview limit"})
			return
		}
		limit = n
	}
	f, info, path, err := s.openWorkspaceFile(r.URL.Query().Get("path"))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusInternalServerError
		} // preserve legacy file API status
		writeJSON(w, status, map[string]any{"error": "Unable to read workspace file"})
		return
	}
	defer f.Close()
	// Read at least a sniff window, never more than the preview bound plus a
	// short UTF-8 tail. Large files are not read in full for metadata/preview.
	raw, err := io.ReadAll(io.LimitReader(f, int64(max(512, limit)+utf8.UTFMax)))
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "Unable to read workspace file"})
		return
	}
	contentType := workspaceContentType(path, raw[:min(512, len(raw))])
	kind := "binary"
	switch contentType {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp":
		kind = "image"
	default:
		sample := raw
		if info.Size() > int64(len(raw)) {
			// A bounded sample may end inside a code point; only trim that suffix.
			for i := max(0, len(sample)-3); i < len(sample); i++ {
				if !utf8.FullRune(sample[i:]) {
					sample = sample[:i]
					break
				}
			}
		}
		if utf8.Valid(sample) && !strings.ContainsRune(string(sample), 0) && (strings.HasPrefix(contentType, "text/") || contentType == "image/svg+xml" || len(raw) == 0) {
			kind = "text"
		}
	}
	preview := map[string]any{"path": path, "kind": kind, "content_type": contentType, "size": info.Size(), "mtime": info.ModTime().UTC().Format(time.RFC3339Nano)}
	if kind == "text" {
		text := raw[:min(limit, len(raw))]
		for len(text) > 0 && !utf8.Valid(text) {
			text = text[:len(text)-1]
		}
		preview["text"], preview["content"] = string(text), string(text) // legacy content alias
		preview["truncated"] = info.Size() > int64(len(text))
	}
	if kind == "image" {
		preview["url"] = "/api/workspace/raw?path=" + url.QueryEscape(path)
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	writeJSON(w, 200, preview)
}

func (s *Server) handleWorkspaceRaw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	f, info, path, err := s.openWorkspaceFile(r.URL.Query().Get("path"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	sample := make([]byte, 512)
	n, err := f.Read(sample)
	if err != nil && err != io.EOF {
		http.Error(w, "Unable to read workspace file", 500)
		return
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Unable to read workspace file", 500)
		return
	}
	contentType := workspaceContentType(path, sample[:n])
	disposition := "attachment"
	switch contentType {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp":
		disposition = "inline"
	}
	if r.URL.Query().Get("download") == "1" {
		disposition = "attachment"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": filepath.Base(path)}))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}
