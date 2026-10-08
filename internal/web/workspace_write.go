package web

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// Workspace writes follow Piclaw's explorer API (runtime/src/channels/web/workspace/
// file-service.ts): the same request shapes, status codes, error strings and
// response bodies, under Gi's /api prefix. Every operation goes through os.Root
// so symlinks and concurrent renames cannot escape the workspace.

const (
	workspaceMaxEditBytes = 256 * 1024
)

type workspaceResult struct {
	status int
	body   map[string]any
}

func workspaceErr(status int, message string) workspaceResult {
	return workspaceResult{status, map[string]any{"error": message}}
}

func workspaceExists(message string) workspaceResult {
	return workspaceResult{http.StatusConflict, map[string]any{"error": message, "code": "file_exists"}}
}

func (s *Server) workspaceRootPath() string {
	if s.cfg.WorkspaceRoot != "" {
		return s.cfg.WorkspaceRoot
	}
	return "/workspace"
}

// workspaceRel resolves a request path to a clean workspace-relative path.
// Empty means the workspace root, as in Piclaw.
func (s *Server) workspaceRel(raw string) (string, bool) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return ".", true
	}
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(s.workspaceRootPath(), path)
		if err != nil {
			return "", false
		}
		path = rel
	}
	path = filepath.Clean(filepath.FromSlash(path))
	if !filepath.IsLocal(path) && path != "." {
		return "", false
	}
	return path, true
}

func workspaceEntryName(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) || filepath.Base(name) != name {
		return "", false
	}
	return name, true
}

func (s *Server) withWorkspaceRoot(w http.ResponseWriter, fn func(*os.Root) workspaceResult) {
	s.workspaceWriteMu.Lock()
	defer s.workspaceWriteMu.Unlock()
	root, err := os.OpenRoot(s.workspaceRootPath())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Workspace unavailable"})
		return
	}
	defer root.Close()
	result := fn(root)
	writeJSON(w, result.status, result.body)
}

func slashPath(path string) string { return filepath.ToSlash(path) }

func (s *Server) handleWorkspaceFileWrite(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			Path    string  `json:"path"`
			Name    string  `json:"name"`
			Content *string `json:"content"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, workspaceMaxEditBytes*2))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeJSON(w, 400, map[string]any{"error": "Invalid JSON"})
			return
		}
		if decoder.Decode(&struct{}{}) != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Invalid trailing JSON"})
			return
		}
		s.withWorkspaceRoot(w, func(root *os.Root) workspaceResult { return s.workspaceCreate(root, req.Path, req.Name, req.Content) })
	case http.MethodPut:
		var req struct {
			Path             string  `json:"path"`
			Content          *string `json:"content"`
			ExpectedRevision string  `json:"expected_revision"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, workspaceMaxEditBytes*2))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeJSON(w, 400, map[string]any{"error": "Invalid JSON"})
			return
		}
		if decoder.Decode(&struct{}{}) != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Invalid trailing JSON"})
			return
		}
		s.withWorkspaceRoot(w, func(root *os.Root) workspaceResult {
			return s.workspaceUpdate(root, req.Path, req.Content, req.ExpectedRevision)
		})
	case http.MethodDelete:
		path := r.URL.Query().Get("path")
		s.withWorkspaceRoot(w, func(root *os.Root) workspaceResult { return s.workspaceDelete(root, path) })
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) workspaceCreate(root *os.Root, dirParam, nameParam string, content *string) workspaceResult {
	dir, ok := s.workspaceRel(dirParam)
	if !ok {
		return workspaceErr(400, "Invalid path")
	}
	name, ok := workspaceEntryName(nameParam)
	if !ok {
		return workspaceErr(400, "Invalid filename")
	}
	if content == nil {
		return workspaceErr(400, "Missing file content")
	}
	if info, err := root.Stat(dir); err != nil {
		return workspaceErr(404, "Directory not found")
	} else if !info.IsDir() {
		return workspaceErr(400, "Path is not a directory")
	}
	if len(*content) > workspaceMaxEditBytes {
		return workspaceErr(400, "File too large to edit")
	}
	target := filepath.Join(dir, name)
	f, err := root.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return workspaceExists("File already exists")
	}
	if err != nil {
		return workspaceErr(500, "Failed to create file")
	}
	_, werr := f.WriteString(*content)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return workspaceErr(500, "Failed to create file")
	}
	return workspaceResult{200, map[string]any{"path": slashPath(target), "name": name}}
}

func (s *Server) workspaceUpdate(root *os.Root, pathParam string, content *string, expectedRevision string) workspaceResult {
	path, ok := s.workspaceRel(pathParam)
	if !ok || path == "." {
		return workspaceErr(400, "Invalid path")
	}
	if content == nil {
		return workspaceErr(400, "Missing file content")
	}
	if len(*content) > workspaceMaxEditBytes {
		return workspaceErr(400, "File too large to edit")
	}
	if !utf8.ValidString(*content) || strings.ContainsRune(*content, 0) {
		return workspaceErr(400, "File is not editable text")
	}
	info, err := root.Stat(path)
	if err != nil {
		return workspaceErr(404, "File not found")
	}
	if info.IsDir() {
		return workspaceErr(400, "Path is a directory")
	}
	if !info.Mode().IsRegular() {
		return workspaceErr(400, "File is not editable text")
	}
	if expectedRevision == "" {
		return workspaceResult{428, map[string]any{"error": "Complete edit revision required", "code": "revision_required"}}
	}
	f, err := root.Open(path)
	if err != nil {
		return workspaceErr(500, "Failed to read file")
	}
	info, err = f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > workspaceMaxEditBytes {
		f.Close()
		return workspaceErr(400, "File is not editable text")
	}
	previous, err := io.ReadAll(io.LimitReader(f, workspaceMaxEditBytes+1))
	f.Close()
	if err != nil {
		return workspaceErr(500, "Failed to read file")
	}
	if len(previous) > workspaceMaxEditBytes || !utf8.Valid(previous) || strings.ContainsRune(string(previous), 0) {
		return workspaceErr(400, "File is not editable text")
	}
	revision := workspaceRevision(info, previous)
	if revision != expectedRevision {
		return workspaceResult{409, map[string]any{"error": "File changed since loaded revision", "code": "revision_conflict", "revision": revision}}
	}
	if string(previous) == *content {
		return workspaceResult{200, map[string]any{"path": slashPath(path), "name": filepath.Base(path), "size": info.Size(), "mtime": info.ModTime().UTC().Format(time.RFC3339Nano), "revision": revision}}
	}
	if err := root.WriteFile(path, []byte(*content), info.Mode().Perm()); err != nil {
		return workspaceErr(500, "Failed to write file")
	}
	updated, err := root.Stat(path)
	if err != nil {
		return workspaceErr(500, "Failed to write file")
	}
	return workspaceResult{200, map[string]any{"path": slashPath(path), "name": filepath.Base(path), "size": updated.Size(), "mtime": updated.ModTime().UTC().Format(time.RFC3339Nano), "revision": workspaceRevision(updated, []byte(*content))}}
}

func (s *Server) workspaceDelete(root *os.Root, pathParam string) workspaceResult {
	path, ok := s.workspaceRel(pathParam)
	if !ok || path == "." {
		return workspaceErr(400, "Invalid path")
	}
	info, err := root.Lstat(path)
	if err != nil {
		return workspaceErr(404, "File not found")
	}
	if info.IsDir() {
		return workspaceErr(400, "Path is a directory")
	}
	if err := root.Remove(path); err != nil {
		return workspaceErr(500, "Failed to delete file")
	}
	return workspaceResult{200, map[string]any{"path": slashPath(path), "name": filepath.Base(path), "deleted": true}}
}

func (s *Server) handleWorkspaceUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Settings -> General "Upload limit (MB)" (Piclaw's workspaceUploadLimitMb).
	workspaceMaxUploadBytes := int64(s.generalSettings(r.Context()).WorkspaceUploadLimitMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, workspaceMaxUploadBytes+1<<20)
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Missing file"})
		return
	}
	defer file.Close()
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	q := r.URL.Query()
	overwrite := q.Get("overwrite") == "1" || q.Get("overwrite") == "true"
	s.withWorkspaceRoot(w, func(root *os.Root) workspaceResult {
		dir, ok := s.workspaceRel(q.Get("path"))
		if !ok {
			return workspaceErr(400, "Invalid path")
		}
		if info, err := root.Stat(dir); err != nil {
			return workspaceErr(404, "Directory not found")
		} else if !info.IsDir() {
			return workspaceErr(400, "Path is not a directory")
		}
		name := filepath.Base(filepath.FromSlash(strings.ReplaceAll(header.Filename, `\`, "/")))
		if name == "" || name == "." || name == ".." || name == string(filepath.Separator) {
			return workspaceErr(400, "Missing filename")
		}
		if header.Size > workspaceMaxUploadBytes {
			return workspaceErr(400, "File too large to upload")
		}
		target := filepath.Join(dir, name)
		existing, statErr := root.Stat(target)
		existed := statErr == nil
		if existed && !overwrite {
			return workspaceExists("File already exists")
		}
		if existed && existing.IsDir() {
			return workspaceErr(400, "Path is a directory")
		}
		out, err := root.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return workspaceErr(500, "Failed to upload file")
		}
		size, werr := io.Copy(out, io.LimitReader(file, workspaceMaxUploadBytes+1))
		if cerr := out.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil || size > workspaceMaxUploadBytes {
			_ = root.Remove(target)
			if werr == nil {
				return workspaceErr(400, "File too large to upload")
			}
			return workspaceErr(500, "Failed to upload file")
		}
		return workspaceResult{200, map[string]any{"path": slashPath(target), "name": name, "size": size, "overwritten": existed}}
	})
}

func (s *Server) handleWorkspaceRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64*1024)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid JSON"})
		return
	}
	s.withWorkspaceRoot(w, func(root *os.Root) workspaceResult {
		path, ok := s.workspaceRel(req.Path)
		if !ok {
			return workspaceErr(400, "Invalid path")
		}
		if path == "." {
			return workspaceErr(400, "Cannot rename workspace root")
		}
		name, ok := workspaceEntryName(req.Name)
		if !ok {
			return workspaceErr(400, "Invalid filename")
		}
		if _, err := root.Lstat(path); err != nil {
			return workspaceErr(404, "File not found")
		}
		next := filepath.Join(filepath.Dir(path), name)
		if next == path {
			return workspaceResult{200, map[string]any{"path": slashPath(path), "name": name}}
		}
		if _, err := root.Lstat(next); err == nil {
			return workspaceExists("File already exists")
		}
		if err := root.Rename(path, next); err != nil {
			return workspaceErr(500, "Failed to rename file")
		}
		return workspaceResult{200, map[string]any{"path": slashPath(next), "name": name, "old_path": slashPath(path)}}
	})
}

func (s *Server) handleWorkspaceMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path   string `json:"path"`
		Target string `json:"target"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64*1024)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid JSON"})
		return
	}
	s.withWorkspaceRoot(w, func(root *os.Root) workspaceResult {
		source, ok := s.workspaceRel(req.Path)
		if !ok {
			return workspaceErr(400, "Invalid path")
		}
		if source == "." {
			return workspaceErr(400, "Cannot move workspace root")
		}
		target, ok := s.workspaceRel(req.Target)
		if !ok {
			return workspaceErr(400, "Invalid target")
		}
		if info, err := root.Stat(target); err != nil {
			return workspaceErr(404, "Target directory not found")
		} else if !info.IsDir() {
			return workspaceErr(400, "Target is not a directory")
		}
		info, err := root.Lstat(source)
		if err != nil {
			return workspaceErr(404, "File not found")
		}
		name := filepath.Base(source)
		next := filepath.Join(target, name)
		if next == source {
			return workspaceResult{200, map[string]any{"path": slashPath(source), "name": name}}
		}
		if info.IsDir() && (target == source || strings.HasPrefix(target, source+string(filepath.Separator))) {
			return workspaceErr(400, "Cannot move a folder into itself")
		}
		if _, err := root.Lstat(next); err == nil {
			return workspaceExists("Target already exists")
		}
		if err := root.Rename(source, next); err != nil {
			return workspaceErr(500, "Failed to move file")
		}
		return workspaceResult{200, map[string]any{"path": slashPath(next), "name": name, "old_path": slashPath(source)}}
	})
}

func (s *Server) handleWorkspaceStat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.withWorkspaceRoot(w, func(root *os.Root) workspaceResult {
		path, ok := s.workspaceRel(r.URL.Query().Get("path"))
		if !ok || path == "." {
			return workspaceErr(400, "Invalid path")
		}
		info, err := root.Stat(path)
		if err != nil {
			return workspaceErr(404, "File not found")
		}
		return workspaceResult{200, map[string]any{"path": slashPath(path), "mtime": info.ModTime().UTC().Format(time.RFC3339Nano), "size": info.Size()}}
	})
}
