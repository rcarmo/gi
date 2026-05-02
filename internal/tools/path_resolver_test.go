package tools

import (
	"strings"
	"testing"
)

func TestResolveToolPathWorkspace(t *testing.T) {
	root := t.TempDir()
	resolved, err := ResolveToolPath(root, "docs/test.txt", false)
	if err != nil {
		t.Fatalf("resolve workspace path: %v", err)
	}
	if resolved.IsVFS() {
		t.Fatalf("expected workspace path, got vfs namespace=%q", resolved.VFSNamespace)
	}
	if resolved.WorkspacePath == "" {
		t.Fatalf("empty workspace path")
	}
}

func TestResolveToolPathTraversal(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveToolPath(root, "../evil.txt", false); err == nil {
		t.Fatalf("expected traversal error")
	} else if !strings.Contains(err.Error(), "path escapes workspace") {
		t.Fatalf("expected traversal error, got %q", err)
	}
}

func TestResolveToolPathEmptyPath(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveToolPath(root, "", false); err == nil {
		t.Fatalf("expected empty path error")
	} else if err.Error() != "empty path" {
		t.Fatalf("expected empty path, got %q", err)
	}
}

func TestResolveToolPathMalformedVFS(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveToolPath(root, "vfs://", false); err == nil {
		t.Fatalf("expected malformed vfs error")
	} else if !strings.Contains(err.Error(), "invalid vfs path") {
		t.Fatalf("expected invalid vfs path error, got %q", err)
	}

	if _, err := ResolveToolPath(root, "vfs://scripts/../evil", false); err == nil {
		t.Fatalf("expected traversal vfs error")
	} else if !strings.Contains(err.Error(), "traversal outside namespace") {
		t.Fatalf("expected traversal outside namespace error, got %q", err)
	}
}

func TestResolveToolPathVFSReadOnlyReference(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveToolPath(root, "vfs://reference/system/readme.md", true); err == nil {
		t.Fatalf("expected reference namespace write error")
	} else if !strings.Contains(err.Error(), "vfs namespace is read-only") {
		t.Fatalf("expected read-only error, got %q", err)
	}
}

func TestResolveToolPathVFSWritableNamespace(t *testing.T) {
	root := t.TempDir()
	resolved, err := ResolveToolPath(root, "vfs://scripts/doc.md", false)
	if err != nil {
		t.Fatalf("resolve vfs path: %v", err)
	}
	if !resolved.IsVFS() {
		t.Fatal("expected vfs resolution")
	}
	if resolved.VFSNamespace != "scripts" || resolved.VFSPath != "doc.md" {
		t.Fatalf("unexpected vfs resolution: %#v", resolved)
	}
}
