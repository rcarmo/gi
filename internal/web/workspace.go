package web

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type workspaceNode struct {
	Name     string          `json:"name"`
	Path     string          `json:"path"`
	Type     string          `json:"type"`
	Size     int64           `json:"size,omitempty"`
	Children []workspaceNode `json:"children"`
}

const workspaceTreeNodeLimit = 10000

var errWorkspaceTreeLimit = errors.New("workspace subtree exceeds entry limit")

func (s *Server) handleWorkspaceTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	root := s.cfg.WorkspaceRoot
	if root == "" {
		root = "/workspace"
	}
	q := r.URL.Query()
	path := q.Get("path")
	if path == "" {
		path = "."
	}
	path = filepath.Clean(path)
	if !filepath.IsLocal(path) {
		writeJSON(w, 400, map[string]any{"error": "path escapes workspace"})
		return
	}
	depth := 2 // Preserve the original no-argument tree response.
	if q.Has("depth") {
		n, err := strconv.Atoi(q.Get("depth"))
		if err != nil || n < 0 || n > 8 {
			writeJSON(w, 400, map[string]any{"error": "invalid tree depth (0–8)"})
			return
		}
		depth = n
	}
	showHidden := true // Legacy callers receive .pi/.piclaw as before.
	if q.Has("show_hidden") {
		switch q.Get("show_hidden") {
		case "true":
			showHidden = true
		case "false":
			showHidden = false
		default:
			writeJSON(w, 400, map[string]any{"error": "invalid hidden-files flag"})
			return
		}
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "Unable to read workspace tree"})
		return
	}
	defer dir.Close()
	remaining := workspaceTreeNodeLimit
	legacy := !q.Has("show_hidden")
	node, err := readWorkspaceTree(dir, path, depth, showHidden, legacy, &remaining)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errWorkspaceTreeLimit) {
			status = http.StatusRequestEntityTooLarge
		}
		writeJSON(w, status, map[string]any{"error": "Unable to read workspace subtree", "detail": func() string {
			if errors.Is(err, errWorkspaceTreeLimit) {
				return errWorkspaceTreeLimit.Error()
			}
			return "Path unavailable"
		}()})
		return
	}
	if node.Path == "." {
		node.Name = filepath.Base(root)
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeJSON(w, http.StatusOK, node)
}

func (s *Server) handleWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	s.handleWorkspacePreview(w, r)
}

// A global node budget and bounded per-directory reads prevent partial trees
// masquerading as complete snapshots. Rooted opens never follow outside links.
func readWorkspaceTree(root *os.Root, path string, depth int, showHidden, legacy bool, remaining *int) (workspaceNode, error) {
	if *remaining <= 0 {
		return workspaceNode{}, errWorkspaceTreeLimit
	}
	*remaining = *remaining - 1
	info, err := root.Stat(path)
	if err != nil {
		return workspaceNode{}, err
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return workspaceNode{}, fmt.Errorf("unsupported workspace entry")
	}
	node := workspaceNode{Name: filepath.Base(path), Path: filepath.ToSlash(path), Type: "file", Size: info.Size()}
	if !info.IsDir() {
		return node, nil
	}
	node.Type = "dir"
	node.Size = 0
	if depth == 0 {
		return node, nil
	}
	// [] means a fetched empty directory. null means an unexpanded stub; the
	// supplied explorer deliberately preserves cached children for stubs.
	node.Children = []workspaceNode{}
	dir, err := root.Open(path)
	if err != nil {
		return workspaceNode{}, err
	}
	entries, readErr := dir.ReadDir(workspaceTreeNodeLimit + 1)
	dir.Close()
	if readErr != nil && readErr != io.EOF {
		return workspaceNode{}, readErr
	}
	if len(entries) > workspaceTreeNodeLimit {
		return workspaceNode{}, errWorkspaceTreeLimit
	}
	for _, entry := range entries {
		name := entry.Name()
		if (!showHidden && strings.HasPrefix(name, ".")) || (legacy && (strings.HasPrefix(name, ".git") || name == "node_modules")) {
			continue
		}
		child, err := readWorkspaceTree(root, filepath.Join(path, name), depth-1, showHidden, legacy, remaining)
		if errors.Is(err, errWorkspaceTreeLimit) {
			return workspaceNode{}, err
		}
		if err != nil {
			continue
		} // Broken/outside symlinks never enter a returned tree.
		node.Children = append(node.Children, child)
	}
	sort.Slice(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]
		if a.Type != b.Type {
			return a.Type == "dir"
		}
		return a.Name < b.Name
	})
	return node, nil
}
