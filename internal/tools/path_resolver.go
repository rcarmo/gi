package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	storevfs "github.com/rcarmo/gi/internal/store/vfs"
)

// ToolPath represents a resolved tool input path after validation.
// It can point at either the workspace filesystem or a managed VFS namespace.
type ToolPath struct {
	WorkspacePath string
	VFSNamespace  string
	VFSPath       string
	isVFS         bool
}

// IsVFS reports whether this path resolves into a managed VFS namespace.
func (p ToolPath) IsVFS() bool { return p.isVFS }

func resolveToolPath(root, raw string, writable bool) (resolvedPath, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return resolvedPath{}, fmt.Errorf("empty path")
	}
	if strings.HasPrefix(trimmed, "fts://") {
		if writable {
			return resolvedPath{}, fmt.Errorf("fts namespace is read-only")
		}
		locator := strings.TrimLeft(strings.TrimPrefix(trimmed, "fts://"), "/")
		if locator == "" {
			locator = "help"
		}
		return resolvedPath{workspacePath: "", vfsNamespace: "fts", vfsPath: locator, isVFS: true}, nil
	}
	if strings.HasPrefix(trimmed, "vfs://") {
		ns, vpath, err := storevfs.ParsePath(trimmed)
		if err != nil {
			return resolvedPath{}, err
		}
		if writable && (ns == "reference" || ns == "chat") {
			return resolvedPath{}, fmt.Errorf("vfs namespace is read-only: %s", ns)
		}
		return resolvedPath{workspacePath: "", vfsNamespace: ns, vfsPath: vpath, isVFS: true}, nil
	}
	full := filepath.Join(root, trimmed)
	clean := filepath.Clean(full)
	rootClean := filepath.Clean(root)
	rootResolved := rootClean
	if rp, err := filepath.EvalSymlinks(rootClean); err == nil {
		rootResolved = filepath.Clean(rp)
	}
	targetResolved := clean
	if tp, err := filepath.EvalSymlinks(clean); err == nil {
		targetResolved = filepath.Clean(tp)
	} else if os.IsNotExist(err) {
		parent := filepath.Dir(clean)
		if pp, perr := filepath.EvalSymlinks(parent); perr == nil {
			targetResolved = filepath.Clean(filepath.Join(pp, filepath.Base(clean)))
		}
	} else {
		return resolvedPath{}, err
	}
	if !pathWithinRoot(targetResolved, rootResolved) {
		return resolvedPath{}, fmt.Errorf("path escapes workspace")
	}
	return resolvedPath{workspacePath: clean, isVFS: false}, nil
}

func pathWithinRoot(path, root string) bool {
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	return path == root || strings.HasPrefix(path, root+string(os.PathSeparator))
}

type resolvedPath struct {
	workspacePath string
	vfsNamespace  string
	vfsPath       string
	isVFS         bool
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
