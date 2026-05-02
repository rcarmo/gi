package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

// ToolPath represents a resolved tool input path after validation.
// It can point at either the workspace filesystem or a managed VFS namespace.
type ToolPath struct {
	WorkspacePath string
	VFSNamespace string
	VFSPath      string
	isVFS        bool
}

// IsVFS reports whether this path resolves into a managed VFS namespace.
func (p ToolPath) IsVFS() bool { return p.isVFS }

func resolveToolPath(root, raw string, writable bool) (resolvedPath, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return resolvedPath{}, fmt.Errorf("empty path")
	}
	if strings.HasPrefix(trimmed, "vfs://") {
		ns, vpath, err := store.ParseVFSPath(trimmed)
		if err != nil {
			return resolvedPath{}, err
		}
		if writable && ns == "reference" {
			return resolvedPath{}, fmt.Errorf("vfs namespace is read-only: %s", ns)
		}
		return resolvedPath{workspacePath: "", vfsNamespace: ns, vfsPath: vpath, isVFS: true}, nil
	}
	full := filepath.Join(root, trimmed)
	clean := filepath.Clean(full)
	rootClean := filepath.Clean(root)
	if !strings.HasPrefix(clean, rootClean+string(os.PathSeparator)) && clean != rootClean {
		return resolvedPath{}, fmt.Errorf("path escapes workspace")
	}
	return resolvedPath{workspacePath: clean, isVFS: false}, nil
}

type resolvedPath struct {
	workspacePath string
	vfsNamespace string
	vfsPath      string
	isVFS        bool
}

// ResolveToolPath exposes the shared path resolution strategy to other packages.
func ResolveToolPath(root, raw string, writable bool) (ToolPath, error) {
	resolved, err := resolveToolPath(root, raw, writable)
	if err != nil {
		return ToolPath{}, err
	}
	return ToolPath{
		WorkspacePath: resolved.workspacePath,
		VFSNamespace:  resolved.vfsNamespace,
		VFSPath:       resolved.vfsPath,
		isVFS:         resolved.isVFS,
	}, nil
}
