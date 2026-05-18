package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteToolListsAndReads(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".gi", "skills", "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "Name: demo\nDescription: Demo skill\n\nUse this skill."
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	list, err := ExecuteTool(root, map[string]any{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(list, "demo") || !strings.Contains(list, "Demo skill") {
		t.Fatalf("list=%s", list)
	}
	read, err := ExecuteTool(root, map[string]any{"name": "demo"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if read != body {
		t.Fatalf("read=%q", read)
	}
}
