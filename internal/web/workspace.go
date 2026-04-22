package web

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type workspaceNode struct {
	Name     string          `json:"name"`
	Path     string          `json:"path"`
	Type     string          `json:"type"`
	Size     int64           `json:"size,omitempty"`
	Children []workspaceNode `json:"children,omitempty"`
}

func (s *Server) handleWorkspaceTree(w http.ResponseWriter, r *http.Request) {
	root := s.cfg.WorkspaceRoot
	if root == "" {
		root = "/workspace"
	}
	node, err := buildWorkspaceTree(root, root, 0, 2)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (s *Server) handleWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	root := s.cfg.WorkspaceRoot
	if root == "" {
		root = "/workspace"
	}
	rel := strings.TrimSpace(r.URL.Query().Get("path"))
	if rel == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing path"})
		return
	}
	full := filepath.Clean(filepath.Join(root, rel))
	rootClean := filepath.Clean(root)
	if !strings.HasPrefix(full, rootClean+string(os.PathSeparator)) && full != rootClean {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "path escapes workspace"})
		return
	}
	data, err := os.ReadFile(full)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"path": rel, "content": string(data)})
}

func buildWorkspaceTree(root, current string, depth, maxDepth int) (workspaceNode, error) {
	info, err := os.Stat(current)
	if err != nil {
		return workspaceNode{}, err
	}
	rel := "."
	if current != root {
		rel, _ = filepath.Rel(root, current)
	}
	node := workspaceNode{Name: info.Name(), Path: filepath.ToSlash(rel), Type: "file", Size: info.Size()}
	if info.IsDir() {
		node.Type = "dir"
		node.Size = 0
		if depth < maxDepth {
			entries, err := os.ReadDir(current)
			if err != nil {
				return workspaceNode{}, err
			}
			children := make([]workspaceNode, 0, len(entries))
			for _, entry := range entries {
				name := entry.Name()
				if strings.HasPrefix(name, ".git") || name == "node_modules" {
					continue
				}
				child, err := buildWorkspaceTree(root, filepath.Join(current, name), depth+1, maxDepth)
				if err != nil {
					continue
				}
				children = append(children, child)
			}
			sort.Slice(children, func(i, j int) bool {
				if children[i].Type != children[j].Type {
					return children[i].Type == "dir"
				}
				return children[i].Name < children[j].Name
			})
			node.Children = children
		}
	}
	if node.Path == "." {
		node.Name = filepath.Base(root)
	}
	return node, nil
}
